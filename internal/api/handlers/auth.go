package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"dermify-api/config"
	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/service"
)

type registerRequest struct {
	Username string `json:"username" example:"johndoe"`
	Email    string `json:"email" example:"johndoe@example.com"`
	Password string `json:"password" example:"secretpassword"`
}

type registerResponse struct {
	ID       int64  `json:"id" example:"1"`
	Username string `json:"username" example:"johndoe"`
	Email    string `json:"email" example:"johndoe@example.com"`
	Message  string `json:"message" example:"user registered successfully"`
}

// HandleRegister creates a new user account with a hashed password.
//
//	@Summary		Register a new user
//	@Description	Creates a new user account with a hashed password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		registerRequest		true	"Registration details"
//	@Success		201		{object}	registerResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/auth/register [post]
func HandleRegister(authSvc *service.AuthService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.Username == "" || req.Email == "" || req.Password == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "username, email, and password are required")
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalPasswordProcessing, "failed to process password")
			return
		}

		user, err := authSvc.Register(r.Context(), req.Username, req.Email, hash)
		if err != nil {
			handleAuthError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(registerResponse{ //nolint:errcheck // response write
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Message:  "user registered successfully",
		})
	}
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" example:"dGhpcyBpcyBhIHJlZnJl..."`
}

// HandleLogout revokes a refresh token.
//
//	@Summary		Logout
//	@Description	Revokes the provided refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		logoutRequest	true	"Logout request"
//	@Success		200		{object}	MessageResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Router			/auth/logout [post]
func HandleLogout(authSvc *service.AuthService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req logoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.RefreshToken != "" {
			tokenHash := auth.HashToken(req.RefreshToken)
			_ = authSvc.RevokeRefreshToken(r.Context(), tokenHash)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"}) //nolint:errcheck // response write
	}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"dGhpcyBpcyBhIHJlZnJl..."`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"dGhpcyBpcyBhIHJlZnJl..."`
	ExpiresIn    int    `json:"expires_in" example:"3600"`
}

// HandleRefreshToken validates an existing refresh token, revokes it, and issues
// a new access token and refresh token (token rotation).
//
//	@Summary		Refresh token
//	@Description	Validates an existing refresh token, revokes it, and issues a new access/refresh token pair
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		refreshRequest	true	"Refresh token request"
//	@Success		200		{object}	refreshResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		401		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/auth/refresh [post]
func HandleRefreshToken(authSvc *service.AuthService, cfg *config.Configuration, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.RefreshToken == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.AuthRefreshTokenRequired, "refresh_token is required")
			return
		}

		oldHash := auth.HashToken(req.RefreshToken)

		userID, err := authSvc.ValidateRefreshToken(r.Context(), oldHash)
		if err != nil {
			handleAuthError(w, err)
			return
		}

		_ = authSvc.RevokeRefreshToken(r.Context(), oldHash)

		user, err := authSvc.GetUserByID(r.Context(), userID)
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalUserLookup, "failed to look up user")
			return
		}

		accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalTokenGeneration, "failed to generate access token")
			return
		}

		newRefreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalRefreshTokenGeneration, "failed to generate refresh token")
			return
		}

		newHash := auth.HashToken(newRefreshToken)
		expiresAt := time.Now().Add(cfg.Auth.RefreshTokenExpiry)

		if err := authSvc.StoreRefreshToken(r.Context(), user.ID, newHash, expiresAt); err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalRefreshTokenStorage, "failed to store refresh token")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(refreshResponse{ //nolint:errcheck // response write
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    int(cfg.Auth.AccessTokenExpiry.Seconds()),
		})
	}
}

type profileResponse struct {
	ID                 int64  `json:"id" example:"1"`
	Username           string `json:"username" example:"johndoe"`
	Email              string `json:"email" example:"johndoe@example.com"`
	Bio                string `json:"bio" example:"Software developer"`
	Role               string `json:"role" example:"doctor"`
	Language           string `json:"language" example:"en"`
	Timezone           string `json:"timezone" example:"UTC"`
	MustChangePassword bool   `json:"must_change_password"`
	CreatedAt          string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// HandleGetProfile returns the authenticated user's profile.
//
//	@Summary		Get user profile
//	@Description	Returns the authenticated user's profile
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	profileResponse
//	@Failure		401	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/auth/me [get]
func HandleGetProfile(authSvc *service.AuthService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		user, err := authSvc.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			handleAuthError(w, err)
			return
		}

		resp := profileResponse{
			ID:                 user.ID,
			Username:           user.Username,
			Email:              user.Email,
			Role:               user.Role,
			Language:           user.Language,
			Timezone:           user.Timezone,
			MustChangePassword: user.MustChangePassword,
			CreatedAt:          user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if user.Bio != nil {
			resp.Bio = *user.Bio
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // response write
	}
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// HandleChangePassword updates the authenticated user's password and clears must-change flag.
func HandleChangePassword(authSvc *service.AuthService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		var req changePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.CurrentPassword == "" || req.NewPassword == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "current_password and new_password are required")
			return
		}
		if len(req.NewPassword) < 8 {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "new password must be at least 8 characters")
			return
		}

		user, err := authSvc.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			handleAuthError(w, err)
			return
		}
		if !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthInvalidCredentials, "current password is incorrect")
			return
		}

		hash, err := auth.HashPassword(req.NewPassword)
		if err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalPasswordProcessing, "failed to process password")
			return
		}

		if err := authSvc.UpdatePassword(r.Context(), claims.UserID, hash, true); err != nil {
			slog.Error("failed to update password", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalUserLookup, "failed to update password")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "password updated"}) //nolint:errcheck // response write
	}
}

type updatePreferencesRequest struct {
	Language string `json:"language" example:"en"`
	Timezone string `json:"timezone" example:"UTC"`
}

// HandleUpdatePreferences updates the authenticated user's language and timezone.
//
//	@Summary		Update user preferences
//	@Description	Updates the authenticated user's language and timezone
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		updatePreferencesRequest	true	"Preferences"
//	@Success		200		{object}	MessageResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		401		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/auth/preferences [put]
func HandleUpdatePreferences(userSvc *service.UserService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized, apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		var req updatePreferencesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if req.Language == "" && req.Timezone == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.PreferencesInvalidData, "at least one of language or timezone is required")
			return
		}

		if err := userSvc.UpdatePreferences(r.Context(), claims.UserID, req.Language, req.Timezone); err != nil {
			slog.Error("failed to update preferences", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.PreferencesUpdateFailed, "failed to update preferences")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "preferences updated"}) //nolint:errcheck // response write
	}
}

// handleAuthError maps service auth errors to HTTP responses.
func handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserAlreadyExists):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.UserAlreadyExists, "username or email already exists")
	case errors.Is(err, service.ErrInvalidCredentials):
		apierrors.WriteError(w, http.StatusUnauthorized,
			apierrors.AuthInvalidCredentials, "invalid credentials")
	case errors.Is(err, service.ErrRefreshTokenInvalid):
		apierrors.WriteError(w, http.StatusUnauthorized,
			apierrors.AuthInvalidRefreshToken, "invalid or expired refresh token")
	case errors.Is(err, service.ErrUserNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.UserNotFound, "user not found")
	default:
		slog.Error("auth operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.InternalUserLookup, "internal error")
	}
}
