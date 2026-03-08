package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/service"
)

// handleInjectableModuleError maps service injectable module errors to HTTP responses.
// This is shared across filler and botulinum module handler files.
func handleInjectableModuleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrModuleDetailNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.ModuleDetailNotFound, "module detail not found")
	case errors.Is(err, service.ErrModuleDetailVersionConflict):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.ModuleDetailVersionConflict, "module detail was modified by another user")
	case errors.Is(err, service.ErrProductTypeMismatch):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleProductTypeMismatch, "product type does not match module type")
	case errors.Is(err, service.ErrInvalidInjectionSites):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleInvalidInjectionSites, "invalid injection sites data")
	case errors.Is(err, service.ErrInvalidModuleData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleInvalidData, "invalid module data")
	case errors.Is(err, service.ErrProductNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.RegistryProductNotFound, "product not found")
	case errors.Is(err, service.ErrConsentRequired):
		apierrors.WriteError(w, http.StatusForbidden,
			apierrors.ConsentRequired, "consent required before adding modules")
	case errors.Is(err, service.ErrSessionNotEditable):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionNotEditable, "session is not editable in current state")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	default:
		slog.Error("injectable module operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.InjectableModuleCreationFailed, "injectable module operation failed")
	}
}
