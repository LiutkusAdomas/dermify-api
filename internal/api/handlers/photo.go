package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

// HandleUploadBeforePhoto uploads a before-type photo to a session.
//
//	@Summary		Upload before photo
//	@Description	Uploads a before-type photo to a session.
//	@Tags			photos
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"Session ID"
//	@Param			photo	formData	file	true	"Photo file"
//	@Success		201		{object}	domain.Photo
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		403		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/photos/before [post]
func HandleUploadBeforePhoto(svc *service.PhotoService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		sessionID, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, domain.MaxPhotoSize)

		file, header, err := r.FormFile("photo")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "invalid or missing photo file")
			return
		}
		defer file.Close()

		// Detect content type from file contents.
		buf := make([]byte, 512) //nolint:mnd // 512 bytes for content type detection
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "unable to read photo file")
			return
		}

		contentType := http.DetectContentType(buf[:n])

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.PhotoUploadFailed, "failed to process photo file")
			return
		}

		photo := &domain.Photo{
			SessionID:    sessionID,
			PhotoType:    domain.PhotoTypeBefore,
			OriginalName: header.Filename,
			ContentType:  contentType,
			SizeBytes:    header.Size,
			CreatedBy:    claims.UserID,
			UpdatedBy:    claims.UserID,
		}

		if err := svc.UploadPhoto(r.Context(), photo, file); err != nil {
			handlePhotoError(w, err)
			return
		}

		m.IncrementPhotoUploadedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(photo) //nolint:errcheck // response write
	}
}

// HandleUploadLabelPhoto uploads a label-type photo to a session module.
//
//	@Summary		Upload label photo
//	@Description	Uploads a label-type photo for a specific treatment module.
//	@Tags			photos
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int		true	"Session ID"
//	@Param			moduleId	path		int		true	"Module ID"
//	@Param			photo		formData	file	true	"Photo file"
//	@Success		201			{object}	domain.Photo
//	@Failure		400			{object}	apierrors.ErrorResponse
//	@Failure		403			{object}	apierrors.ErrorResponse
//	@Failure		409			{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/photos/label/{moduleId} [post]
func HandleUploadLabelPhoto(svc *service.PhotoService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		sessionID, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		moduleID, err := parseModuleIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "invalid module ID")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, domain.MaxPhotoSize)

		file, header, err := r.FormFile("photo")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "invalid or missing photo file")
			return
		}
		defer file.Close()

		// Detect content type from file contents.
		buf := make([]byte, 512) //nolint:mnd // 512 bytes for content type detection
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "unable to read photo file")
			return
		}

		contentType := http.DetectContentType(buf[:n])

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.PhotoUploadFailed, "failed to process photo file")
			return
		}

		photo := &domain.Photo{
			SessionID:    sessionID,
			ModuleID:     &moduleID,
			PhotoType:    domain.PhotoTypeLabel,
			OriginalName: header.Filename,
			ContentType:  contentType,
			SizeBytes:    header.Size,
			CreatedBy:    claims.UserID,
			UpdatedBy:    claims.UserID,
		}

		if err := svc.UploadPhoto(r.Context(), photo, file); err != nil {
			handlePhotoError(w, err)
			return
		}

		m.IncrementPhotoUploadedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(photo) //nolint:errcheck // response write
	}
}

// HandleListSessionPhotos returns all photos for a session.
//
//	@Summary		List session photos
//	@Description	Returns all photos for a session.
//	@Tags			photos
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Session ID"
//	@Success		200	{array}		domain.Photo
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/photos [get]
func HandleListSessionPhotos(svc *service.PhotoService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = m // metrics client available for future use

		sessionID, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		photos, err := svc.ListBySession(r.Context(), sessionID)
		if err != nil {
			handlePhotoError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(photos) //nolint:errcheck // response write
	}
}

// HandleGetPhoto returns metadata for a single photo.
//
//	@Summary		Get photo metadata
//	@Description	Returns metadata for a single photo.
//	@Tags			photos
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int	true	"Session ID"
//	@Param			photoId	path	int	true	"Photo ID"
//	@Success		200	{object}	domain.Photo
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/photos/{photoId} [get]
func HandleGetPhoto(svc *service.PhotoService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = m // metrics client available for future use

		photoID, err := parsePhotoIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "invalid photo ID")
			return
		}

		photo, err := svc.GetByID(r.Context(), photoID)
		if err != nil {
			handlePhotoError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(photo) //nolint:errcheck // response write
	}
}

// HandleServePhotoFile serves the actual photo file from the filesystem.
// The file path is derived from the database record (server-controlled), preventing path traversal.
//
//	@Summary		Serve photo file
//	@Description	Serves the actual photo file binary from the filesystem.
//	@Tags			photos
//	@Produce		image/jpeg,image/png
//	@Security		BearerAuth
//	@Param			id		path	int	true	"Session ID"
//	@Param			photoId	path	int	true	"Photo ID"
//	@Success		200
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/photos/{photoId}/file [get]
func HandleServePhotoFile(svc *service.PhotoService, basePath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		photoID, err := parsePhotoIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "invalid photo ID")
			return
		}

		photo, err := svc.GetByID(r.Context(), photoID)
		if err != nil {
			handlePhotoError(w, err)
			return
		}

		fullPath := filepath.Join(basePath, filepath.FromSlash(photo.FilePath))

		w.Header().Set("Content-Type", photo.ContentType)
		http.ServeFile(w, r, fullPath)
	}
}

// HandleDeletePhoto removes a photo from a session.
//
//	@Summary		Delete photo
//	@Description	Removes a photo from a session. Not allowed on signed/locked sessions.
//	@Tags			photos
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int	true	"Session ID"
//	@Param			photoId	path	int	true	"Photo ID"
//	@Success		204
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Failure		409	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/photos/{photoId} [delete]
func HandleDeletePhoto(svc *service.PhotoService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = m // metrics client available for future use

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		photoID, err := parsePhotoIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.PhotoInvalidData, "invalid photo ID")
			return
		}

		if err := svc.DeletePhoto(r.Context(), photoID, claims.UserID); err != nil {
			handlePhotoError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// parsePhotoIDParam extracts the photo ID from the URL path parameter.
func parsePhotoIDParam(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "photoId")
	return strconv.ParseInt(idStr, 10, 64)
}

