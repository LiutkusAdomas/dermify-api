package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"dermify-api/config"
	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/auth"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/service"
)

type invitationStatusResponse struct {
	OrgID                int64  `json:"org_id"`
	OrgName              string `json:"org_name"`
	Email                string `json:"email"`
	Role                 string `json:"role"`
	ExpiresAt            string `json:"expires_at"`
	RequiresAccountSetup bool   `json:"requires_account_setup"`
}

type completeInvitationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type completeInvitationResponse struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	ExpiresIn          int    `json:"expires_in"`
	Message            string `json:"message"`
	OrgID              int64  `json:"org_id"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"must_change_password"`
}

// HandleGetInvitationStatus returns public invitation details by token.
func HandleGetInvitationStatus(orgSvc *service.OrganizationService, authSvc *service.AuthService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := getURLParam(r, "token")
		if token == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "token is required")
			return
		}

		inv, err := orgSvc.GetInvitationByToken(r.Context(), token)
		if err != nil {
			handleOrgError(w, err)
			return
		}

		requiresSetup := false
		if _, err := authSvc.GetUserByEmail(r.Context(), inv.Email); err != nil {
			if errors.Is(err, service.ErrUserNotFound) {
				requiresSetup = true
			} else {
				handleAuthError(w, err)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invitationStatusResponse{ //nolint:errcheck // response write
			OrgID:                inv.OrgID,
			OrgName:              inv.OrgName,
			Email:                inv.Email,
			Role:                 inv.Role,
			ExpiresAt:            inv.ExpiresAt.Format(time.RFC3339),
			RequiresAccountSetup: requiresSetup,
		})
	}
}

// HandleCompleteInvitation accepts invitation and creates account if needed.
func HandleCompleteInvitation(orgSvc *service.OrganizationService, authSvc *service.AuthService, cfg *config.Configuration, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := getURLParam(r, "token")
		if token == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "token is required")
			return
		}

		var req completeInvitationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		inv, err := orgSvc.GetInvitationByToken(r.Context(), token)
		if err != nil {
			handleOrgError(w, err)
			return
		}

		user, err := authSvc.GetUserByEmail(r.Context(), inv.Email)
		if err != nil {
			if !errors.Is(err, service.ErrUserNotFound) {
				handleAuthError(w, err)
				return
			}

			if strings.TrimSpace(req.Username) == "" || req.Password == "" {
				apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "username and password are required for new users")
				return
			}
			if len(req.Password) < 8 {
				apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "password must be at least 8 characters")
				return
			}

			hash, hashErr := auth.HashPassword(req.Password)
			if hashErr != nil {
				apierrors.WriteError(w, http.StatusInternalServerError, apierrors.InternalPasswordProcessing, "failed to process password")
				return
			}

			user, err = authSvc.Register(r.Context(), strings.TrimSpace(req.Username), inv.Email, hash)
			if err != nil {
				handleAuthError(w, err)
				return
			}
		}

		if err := orgSvc.AcceptInvitation(r.Context(), token, user.ID, user.Email); err != nil {
			handleOrgError(w, err)
			return
		}

		accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)
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
		json.NewEncoder(w).Encode(completeInvitationResponse{ //nolint:errcheck // response write
			AccessToken:        accessToken,
			RefreshToken:       refreshToken,
			ExpiresIn:          int(cfg.Auth.AccessTokenExpiry.Seconds()),
			Message:            "invitation accepted",
			OrgID:              inv.OrgID,
			Role:               inv.Role,
			MustChangePassword: user.MustChangePassword,
		})
	}
}
