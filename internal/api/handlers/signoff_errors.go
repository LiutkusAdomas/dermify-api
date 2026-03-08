package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/service"
)

// handleSignOffError maps service sign-off errors to HTTP responses.
func handleSignOffError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	case errors.Is(err, service.ErrSessionIncomplete):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SignoffSessionIncomplete, "session is incomplete for sign-off")
	case errors.Is(err, service.ErrSessionNotAwaitingSignoff):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SignoffNotReady, "session is not in awaiting_signoff state")
	case errors.Is(err, service.ErrInvalidStateTransition):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionInvalidTransition, "invalid state transition")
	case errors.Is(err, service.ErrSessionVersionConflict):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionVersionConflict, "session was modified by another user")
	default:
		slog.Error("sign-off operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.SignoffFailed, "sign-off operation failed")
	}
}
