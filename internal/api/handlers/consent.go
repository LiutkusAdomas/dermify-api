package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

type recordConsentRequest struct {
	ConsentType    string    `json:"consent_type"`
	ConsentMethod  string    `json:"consent_method"`
	ObtainedAt     time.Time `json:"obtained_at"`
	RisksDiscussed bool      `json:"risks_discussed"`
	Notes          *string   `json:"notes"`
}

type updateConsentRequest struct {
	ConsentType    string    `json:"consent_type"`
	ConsentMethod  string    `json:"consent_method"`
	ObtainedAt     time.Time `json:"obtained_at"`
	RisksDiscussed bool      `json:"risks_discussed"`
	Notes          *string   `json:"notes"`
	Version        int       `json:"version"`
}

// HandleRecordConsent records a consent for a session.
//
//	@Summary		Record consent
//	@Description	Records informed consent for a session (one per session).
//	@Tags			consent
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int						true	"Session ID"
//	@Param			request	body	recordConsentRequest	true	"Consent details"
//	@Success		201		{object}	ConsentResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/consent [post]
func HandleRecordConsent(svc *service.ConsentService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req recordConsentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		consent := &domain.Consent{
			SessionID:      sessionID,
			ConsentType:    req.ConsentType,
			ConsentMethod:  req.ConsentMethod,
			ObtainedAt:     req.ObtainedAt,
			RisksDiscussed: req.RisksDiscussed,
			Notes:          req.Notes,
			CreatedBy:      claims.UserID,
			UpdatedBy:      claims.UserID,
		}

		if err := svc.RecordConsent(r.Context(), consent); err != nil {
			handleConsentCreateError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toConsentResponse(consent)) //nolint:errcheck // response write
	}
}

// HandleGetConsent returns the consent record for a session.
//
//	@Summary		Get consent
//	@Description	Returns the consent record for a session.
//	@Tags			consent
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Session ID"
//	@Success		200	{object}	ConsentResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/consent [get]
func HandleGetConsent(svc *service.ConsentService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sessionID, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		consent, err := svc.GetBySessionID(r.Context(), sessionID)
		if err != nil {
			handleConsentLookupError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toConsentResponse(consent)) //nolint:errcheck // response write
	}
}

// HandleUpdateConsent updates a consent record for a session.
//
//	@Summary		Update consent
//	@Description	Updates the consent record for a session. Requires version for optimistic locking.
//	@Tags			consent
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int						true	"Session ID"
//	@Param			request	body	updateConsentRequest	true	"Updated consent details"
//	@Success		200		{object}	ConsentResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/consent [put]
func HandleUpdateConsent(svc *service.ConsentService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateConsentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		consent := &domain.Consent{
			SessionID:      sessionID,
			ConsentType:    req.ConsentType,
			ConsentMethod:  req.ConsentMethod,
			ObtainedAt:     req.ObtainedAt,
			RisksDiscussed: req.RisksDiscussed,
			Notes:          req.Notes,
			Version:        req.Version,
			UpdatedBy:      claims.UserID,
		}

		if err := svc.UpdateConsent(r.Context(), consent); err != nil {
			handleConsentUpdateError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetBySessionID(r.Context(), sessionID)
		if err != nil {
			slog.Error("failed to fetch updated consent", "session_id", sessionID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ConsentCreationFailed, "consent updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toConsentResponse(updated)) //nolint:errcheck // response write
	}
}

// toConsentResponse converts a domain consent to an API response.
func toConsentResponse(c *domain.Consent) ConsentResponse {
	return ConsentResponse{
		ID:             c.ID,
		SessionID:      c.SessionID,
		ConsentType:    c.ConsentType,
		ConsentMethod:  c.ConsentMethod,
		ObtainedAt:     c.ObtainedAt,
		RisksDiscussed: c.RisksDiscussed,
		Notes:          c.Notes,
		Version:        c.Version,
		CreatedAt:      c.CreatedAt,
		CreatedBy:      c.CreatedBy,
		UpdatedAt:      c.UpdatedAt,
		UpdatedBy:      c.UpdatedBy,
	}
}

// handleConsentCreateError maps service create errors to HTTP responses.
func handleConsentCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrConsentAlreadyExists):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.ConsentAlreadyExists, "consent already exists for this session")
	case errors.Is(err, service.ErrInvalidConsentData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ConsentInvalidData, "invalid consent data: check required fields")
	default:
		slog.Error("failed to record consent", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ConsentCreationFailed, "failed to record consent")
	}
}

// handleConsentLookupError maps service lookup errors to HTTP responses.
func handleConsentLookupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrConsentNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.ConsentNotFound, "consent not found")
	default:
		slog.Error("failed to get consent", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ConsentCreationFailed, "failed to look up consent")
	}
}

// handleConsentUpdateError maps service update errors to HTTP responses.
func handleConsentUpdateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidConsentData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ConsentInvalidData, "invalid consent data: check required fields")
	case errors.Is(err, service.ErrConsentNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.ConsentNotFound, "consent not found")
	default:
		slog.Error("failed to update consent", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ConsentCreationFailed, "failed to update consent")
	}
}
