package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"dermify-api/config"
	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/service"
)

type loginRequest struct {
	Email    string `json:"email" example:"johndoe@example.com"`
	Password string `json:"password" example:"secretpassword"`
}

type loginResponse struct {
	AccessToken        string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken       string `json:"refresh_token" example:"dGhpcyBpcyBhIHJlZnJl..."`
	ExpiresIn          int    `json:"expires_in" example:"3600"`
	Message            string `json:"message" example:"login successful"`
	MustChangePassword bool   `json:"must_change_password"`
}

// HandleLogin authenticates a user with email and password, returning JWT access
// and refresh tokens.
//
//	@Summary		Login
//	@Description	Authenticates a user with email and password, returning JWT access and refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginRequest	true	"Login credentials"
//	@Success		200		{object}	loginResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		401		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/auth/login [post]
func HandleLogin(authSvc *service.AuthService, cfg *config.Configuration, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.Email == "" || req.Password == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "email and password are required")
			return
		}

		user, err := authSvc.Authenticate(r.Context(), req.Email)
		if err != nil {
			m.IncrementLoginFailureCount()
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthInvalidCredentials, "invalid credentials")
			return
		}

		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			m.IncrementLoginFailureCount()
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthInvalidCredentials, "invalid credentials")
			return
		}

		accessToken, err := auth.GenerateAccessToken(
			user.ID, user.Email, user.Role, cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry,
		)
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalTokenGeneration, "failed to generate token")
			return
		}

		refreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalRefreshTokenGeneration, "failed to generate refresh token")
			return
		}

		tokenHash := auth.HashToken(refreshToken)
		expiresAt := time.Now().Add(cfg.Auth.RefreshTokenExpiry)

		if err := authSvc.StoreRefreshToken(r.Context(), user.ID, tokenHash, expiresAt); err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalRefreshTokenStorage, "failed to store refresh token")
			return
		}

		m.IncrementLoginSuccessCount()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResponse{ //nolint:errcheck // response write
			AccessToken:        accessToken,
			RefreshToken:       refreshToken,
			ExpiresIn:          int(cfg.Auth.AccessTokenExpiry.Seconds()),
			Message:            "login successful",
			MustChangePassword: user.MustChangePassword,
		})
	}
}
