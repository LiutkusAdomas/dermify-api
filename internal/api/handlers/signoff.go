package handlers

import (
	"encoding/json"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/service"
)

// HandleGetSignOffReadiness checks whether a session has all required components for sign-off.
func HandleGetSignOffReadiness(svc *service.SignoffService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		result, err := svc.ValidateForSignoff(r.Context(), id)
		if err != nil {
			handleSignOffError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result) //nolint:errcheck // response write
	}
}

// HandleSignOffSession validates completeness and transitions a session to signed status.
func HandleSignOffSession(svc *service.SignoffService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		if err := svc.SignOff(r.Context(), id, claims.UserID); err != nil {
			handleSignOffError(w, err)
			return
		}

		m.IncrementSessionSignedCount()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{Message: "session signed off"}) //nolint:errcheck // response write
	}
}

// HandleLockSession transitions a signed session to the locked state.
func HandleLockSession(svc *service.SignoffService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		if err := svc.LockSession(r.Context(), id, claims.UserID); err != nil {
			handleSignOffError(w, err)
			return
		}

		m.IncrementSessionLockedCount()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{Message: "session locked"}) //nolint:errcheck // response write
	}
}
