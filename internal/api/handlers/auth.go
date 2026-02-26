package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"dermify-api/config"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
)

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Message  string `json:"message"`
}

// HandleRegister creates a new user account with a hashed password.
func HandleRegister(db *sql.DB, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if req.Username == "" || req.Email == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "username, email, and password are required"})
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to process password"})
			return
		}

		var userID int64
		err = db.QueryRow(
			`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
			req.Username, req.Email, hash,
		).Scan(&userID)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "username or email already exists"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(registerResponse{
			ID:       userID,
			Username: req.Username,
			Email:    req.Email,
			Message:  "user registered successfully",
		})
	}
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// HandleLogout revokes a refresh token.
func HandleLogout(db *sql.DB, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req logoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if req.RefreshToken != "" {
			tokenHash := auth.HashToken(req.RefreshToken)
			_ = auth.RevokeRefreshToken(db, tokenHash)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
	}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// HandleRefreshToken validates an existing refresh token, revokes it, and issues
// a new access token and refresh token (token rotation).
func HandleRefreshToken(db *sql.DB, cfg *config.Configuration, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if req.RefreshToken == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "refresh_token is required"})
			return
		}

		oldHash := auth.HashToken(req.RefreshToken)
		userID, err := auth.ValidateRefreshToken(db, oldHash)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired refresh token"})
			return
		}

		_ = auth.RevokeRefreshToken(db, oldHash)

		var email string
		err = db.QueryRow(`SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to look up user"})
			return
		}

		accessToken, err := auth.GenerateAccessToken(userID, email, cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate access token"})
			return
		}

		newRefreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate refresh token"})
			return
		}

		newHash := auth.HashToken(newRefreshToken)
		expiresAt := time.Now().Add(cfg.Auth.RefreshTokenExpiry)
		if err := auth.StoreRefreshToken(db, userID, newHash, expiresAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to store refresh token"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(refreshResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    int(cfg.Auth.AccessTokenExpiry.Seconds()),
		})
	}
}

type profileResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
}

// HandleGetProfile returns the authenticated user's profile.
func HandleGetProfile(db *sql.DB, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
			return
		}

		var resp profileResponse
		var bio sql.NullString
		err := db.QueryRow(
			`SELECT id, username, email, bio, created_at FROM users WHERE id = $1`,
			claims.UserID,
		).Scan(&resp.ID, &resp.Username, &resp.Email, &bio, &resp.CreatedAt)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
			return
		}

		if bio.Valid {
			resp.Bio = bio.String
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
