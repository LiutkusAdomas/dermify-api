package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type createAddendumRequest struct {
	Reason  string `json:"reason"`
	Content string `json:"content"`
}

// HandleCreateAddendum creates a new addendum on a locked session.
//
//	@Summary		Create addendum
//	@Description	Creates a new addendum note on a locked session.
//	@Tags			addendums
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int						true	"Session ID"
//	@Param			request	body	createAddendumRequest	true	"Addendum details"
//	@Success		201		{object}	domain.Addendum
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/addendums [post]
func HandleCreateAddendum(svc *service.AddendumService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req createAddendumRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		addendum := domain.Addendum{
			SessionID: id,
			AuthorID:  claims.UserID,
			Reason:    req.Reason,
			Content:   req.Content,
		}

		if err := svc.CreateAddendum(r.Context(), &addendum); err != nil {
			handleAddendumCreateError(w, err)
			return
		}

		m.IncrementAddendumCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(addendum) //nolint:errcheck // response write
	}
}

// HandleListAddendums returns all addendums for a session.
//
//	@Summary		List addendums
//	@Description	Returns all addendum notes for a session.
//	@Tags			addendums
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Session ID"
//	@Success		200	{array}		domain.Addendum
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		500	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/addendums [get]
func HandleListAddendums(svc *service.AddendumService, _ *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		addendums, err := svc.ListBySession(r.Context(), id)
		if err != nil {
			slog.Error("failed to list addendums", "session_id", id, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.AddendumCreationFailed, "failed to list addendums")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(addendums) //nolint:errcheck // response write
	}
}

// HandleGetAddendum returns a single addendum by ID.
//
//	@Summary		Get addendum
//	@Description	Returns a single addendum by ID.
//	@Tags			addendums
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int	true	"Session ID"
//	@Param			addendumId	path	int	true	"Addendum ID"
//	@Success		200	{object}	domain.Addendum
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/addendums/{addendumId} [get]
func HandleGetAddendum(svc *service.AddendumService, _ *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		addendumIDStr := chi.URLParam(r, "addendumId")

		addendumID, err := strconv.ParseInt(addendumIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.AddendumInvalidData, "invalid addendum ID")
			return
		}

		addendum, err := svc.GetByID(r.Context(), addendumID)
		if err != nil {
			handleAddendumGetError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(addendum) //nolint:errcheck // response write
	}
}

// handleAddendumCreateError maps service addendum creation errors to HTTP responses.
func handleAddendumCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	case errors.Is(err, service.ErrSessionNotLocked):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.AddendumSessionNotLocked, "addendums only allowed on locked sessions")
	case errors.Is(err, service.ErrInvalidAddendumData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.AddendumInvalidData, "invalid addendum data")
	default:
		slog.Error("failed to create addendum", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.AddendumCreationFailed, "failed to create addendum")
	}
}

// handleAddendumGetError maps service addendum lookup errors to HTTP responses.
func handleAddendumGetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAddendumNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.AddendumNotFound, "addendum not found")
	default:
		slog.Error("failed to get addendum", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.AddendumCreationFailed, "failed to get addendum")
	}
}
