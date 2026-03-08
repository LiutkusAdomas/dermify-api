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

type recordOutcomeRequest struct {
	OutcomeStatus  string     `json:"outcome_status"`
	EndpointIDs    []int64    `json:"endpoint_ids,omitempty"`
	AftercareNotes *string    `json:"aftercare_notes,omitempty"`
	RedFlagsText   *string    `json:"red_flags_text,omitempty"`
	ContactInfo    *string    `json:"contact_info,omitempty"`
	FollowUpAt     *time.Time `json:"follow_up_at,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
}

type updateOutcomeRequest struct {
	OutcomeStatus  string     `json:"outcome_status"`
	EndpointIDs    []int64    `json:"endpoint_ids,omitempty"`
	AftercareNotes *string    `json:"aftercare_notes,omitempty"`
	RedFlagsText   *string    `json:"red_flags_text,omitempty"`
	ContactInfo    *string    `json:"contact_info,omitempty"`
	FollowUpAt     *time.Time `json:"follow_up_at,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	Version        int        `json:"version"`
}

// HandleRecordOutcome records a session outcome.
func HandleRecordOutcome(svc *service.OutcomeService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req recordOutcomeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		outcome := &domain.SessionOutcome{
			SessionID:      sessionID,
			OutcomeStatus:  req.OutcomeStatus,
			EndpointIDs:    req.EndpointIDs,
			AftercareNotes: req.AftercareNotes,
			RedFlagsText:   req.RedFlagsText,
			ContactInfo:    req.ContactInfo,
			FollowUpAt:     req.FollowUpAt,
			Notes:          req.Notes,
			CreatedBy:      claims.UserID,
			UpdatedBy:      claims.UserID,
		}

		if err := svc.RecordOutcome(r.Context(), outcome); err != nil {
			handleOutcomeError(w, err)
			return
		}

		m.IncrementOutcomeRecordedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toSessionOutcomeResponse(outcome)) //nolint:errcheck // response write
	}
}

// HandleGetOutcome returns the outcome for a session.
func HandleGetOutcome(svc *service.OutcomeService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sessionID, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		outcome, err := svc.GetBySessionID(r.Context(), sessionID)
		if err != nil {
			handleOutcomeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toSessionOutcomeResponse(outcome)) //nolint:errcheck // response write
	}
}

// HandleUpdateOutcome updates a session outcome.
func HandleUpdateOutcome(svc *service.OutcomeService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateOutcomeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		outcome := &domain.SessionOutcome{
			SessionID:      sessionID,
			OutcomeStatus:  req.OutcomeStatus,
			EndpointIDs:    req.EndpointIDs,
			AftercareNotes: req.AftercareNotes,
			RedFlagsText:   req.RedFlagsText,
			ContactInfo:    req.ContactInfo,
			FollowUpAt:     req.FollowUpAt,
			Notes:          req.Notes,
			Version:        req.Version,
			UpdatedBy:      claims.UserID,
		}

		if err := svc.UpdateOutcome(r.Context(), outcome); err != nil {
			handleOutcomeError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetBySessionID(r.Context(), sessionID)
		if err != nil {
			slog.Error("failed to fetch updated outcome", "session_id", sessionID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.OutcomeUpdateFailed, "outcome updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toSessionOutcomeResponse(updated)) //nolint:errcheck // response write
	}
}

// toSessionOutcomeResponse converts a domain session outcome to an API response.
func toSessionOutcomeResponse(o *domain.SessionOutcome) SessionOutcomeResponse {
	return SessionOutcomeResponse{
		ID:             o.ID,
		SessionID:      o.SessionID,
		OutcomeStatus:  o.OutcomeStatus,
		EndpointIDs:    o.EndpointIDs,
		AftercareNotes: o.AftercareNotes,
		RedFlagsText:   o.RedFlagsText,
		ContactInfo:    o.ContactInfo,
		FollowUpAt:     o.FollowUpAt,
		Notes:          o.Notes,
		Version:        o.Version,
		CreatedAt:      o.CreatedAt,
		CreatedBy:      o.CreatedBy,
		UpdatedAt:      o.UpdatedAt,
		UpdatedBy:      o.UpdatedBy,
	}
}

// handleOutcomeError maps service outcome errors to HTTP responses.
func handleOutcomeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrOutcomeAlreadyExists):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.OutcomeAlreadyExists, "outcome already exists for this session")
	case errors.Is(err, service.ErrInvalidOutcomeData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.OutcomeInvalidData, "invalid outcome data")
	case errors.Is(err, service.ErrSessionNotReady):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.OutcomeSessionNotReady, "session not ready for outcome recording")
	case errors.Is(err, service.ErrOutcomeNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.OutcomeNotFound, "outcome not found")
	case errors.Is(err, service.ErrSessionNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.SessionNotFound, "session not found")
	default:
		slog.Error("outcome operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.OutcomeCreationFailed, "outcome operation failed")
	}
}
