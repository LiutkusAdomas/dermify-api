package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/service"
)

// handleEnergyModuleError maps service energy module errors to HTTP responses.
// This is shared across all four energy module handler files (IPL, NdYAG, CO2, RF).
func handleEnergyModuleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrConsentRequired):
		apierrors.WriteError(w, http.StatusUnprocessableEntity,
			apierrors.ConsentRequired, "consent required before adding modules")
	case errors.Is(err, service.ErrSessionNotEditable):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.SessionNotEditable, "session is not editable in current state")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	case errors.Is(err, service.ErrDeviceNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.RegistryDeviceNotFound, "device not found")
	case errors.Is(err, service.ErrDeviceTypeMismatch):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleDeviceTypeMismatch, "device type does not match module type")
	case errors.Is(err, service.ErrHandpieceMismatch):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleHandpieceMismatch, "handpiece does not belong to device")
	case errors.Is(err, service.ErrModuleDetailNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.ModuleDetailNotFound, "module detail not found")
	case errors.Is(err, service.ErrModuleDetailVersionConflict):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.ModuleDetailVersionConflict, "module detail was modified by another user")
	case errors.Is(err, service.ErrInvalidModuleData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleInvalidData, "invalid module data")
	case errors.Is(err, service.ErrInvalidSessionData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ModuleInvalidData, "invalid module data")
	default:
		slog.Error("energy module operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ModuleCreationFailed, "energy module operation failed")
	}
}
