package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/service"
)

// handlePhotoError maps service photo errors to HTTP responses.
func handlePhotoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPhotoNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.PhotoNotFound, "photo not found")
	case errors.Is(err, service.ErrPhotoConsentRequired):
		apierrors.WriteError(w, http.StatusForbidden,
			apierrors.PhotoConsentRequired, "photo consent required")
	case errors.Is(err, service.ErrPhotoInvalidType):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.PhotoInvalidData, "invalid photo type")
	case errors.Is(err, service.ErrPhotoModuleRequired):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.PhotoInvalidData, "module ID required for label photos")
	case errors.Is(err, service.ErrPhotoInvalidContentType):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.PhotoInvalidData, "invalid content type")
	case errors.Is(err, service.ErrPhotoSessionNotEditable):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.PhotoSessionNotEditable, "session not editable for photo operations")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	default:
		slog.Error("photo operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.PhotoUploadFailed, "photo operation failed")
	}
}
