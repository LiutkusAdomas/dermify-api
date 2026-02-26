package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"dermify-api/config"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/metrics"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Message      string `json:"message"`
}

// HandleLogin authenticates a user with email and password, returning JWT access
// and refresh tokens.
func HandleLogin(db *sql.DB, cfg *config.Configuration, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if req.Email == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "email and password are required"})
			return
		}

		var userID int64
		var passwordHash string
		err := db.QueryRow(
			`SELECT id, password_hash FROM users WHERE email = $1`, req.Email,
		).Scan(&userID, &passwordHash)
		if err != nil {
			m.IncrementLoginFailureCount()
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
			return
		}

		if !auth.CheckPassword(req.Password, passwordHash) {
			m.IncrementLoginFailureCount()
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
			return
		}

		accessToken, err := auth.GenerateAccessToken(
			userID, req.Email, cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"})
			return
		}

		refreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate refresh token"})
			return
		}

		tokenHash := auth.HashToken(refreshToken)
		expiresAt := time.Now().Add(cfg.Auth.RefreshTokenExpiry)
		if err := auth.StoreRefreshToken(db, userID, tokenHash, expiresAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to store refresh token"})
			return
		}

		m.IncrementLoginSuccessCount()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int(cfg.Auth.AccessTokenExpiry.Seconds()),
			Message:      "login successful",
		})
	}
}
