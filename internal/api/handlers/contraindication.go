package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"
)

type recordScreeningRequest struct {
	Pregnant           bool    `json:"pregnant"`
	Breastfeeding      bool    `json:"breastfeeding"`
	ActiveInfection    bool    `json:"active_infection"`
	ActiveColdSores    bool    `json:"active_cold_sores"`
	Isotretinoin       bool    `json:"isotretinoin"`
	Photosensitivity   bool    `json:"photosensitivity"`
	AutoimmuneDisorder bool    `json:"autoimmune_disorder"`
	KeloidHistory      bool    `json:"keloid_history"`
	Anticoagulants     bool    `json:"anticoagulants"`
	RecentTan          bool    `json:"recent_tan"`
	MitigationNotes    *string `json:"mitigation_notes"`
	Notes              *string `json:"notes"`
}

type updateScreeningRequest struct {
	Pregnant           bool    `json:"pregnant"`
	Breastfeeding      bool    `json:"breastfeeding"`
	ActiveInfection    bool    `json:"active_infection"`
	ActiveColdSores    bool    `json:"active_cold_sores"`
	Isotretinoin       bool    `json:"isotretinoin"`
	Photosensitivity   bool    `json:"photosensitivity"`
	AutoimmuneDisorder bool    `json:"autoimmune_disorder"`
	KeloidHistory      bool    `json:"keloid_history"`
	Anticoagulants     bool    `json:"anticoagulants"`
	RecentTan          bool    `json:"recent_tan"`
	MitigationNotes    *string `json:"mitigation_notes"`
	Notes              *string `json:"notes"`
	Version            int     `json:"version"`
}

// HandleRecordScreening records a contraindication screening for a session.
//
//	@Summary		Record screening
//	@Description	Records a contraindication screening for a session (one per session).
//	@Tags			screening
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int						true	"Session ID"
//	@Param			request	body	recordScreeningRequest	true	"Screening details"
//	@Success		201		{object}	ScreeningResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Failure		500		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/screening [post]
func HandleRecordScreening(svc *service.ContraindicationService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req recordScreeningRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		screening := &domain.ContraindicationScreening{
			SessionID:          sessionID,
			Pregnant:           req.Pregnant,
			Breastfeeding:      req.Breastfeeding,
			ActiveInfection:    req.ActiveInfection,
			ActiveColdSores:    req.ActiveColdSores,
			Isotretinoin:       req.Isotretinoin,
			Photosensitivity:   req.Photosensitivity,
			AutoimmuneDisorder: req.AutoimmuneDisorder,
			KeloidHistory:      req.KeloidHistory,
			Anticoagulants:     req.Anticoagulants,
			RecentTan:          req.RecentTan,
			MitigationNotes:    req.MitigationNotes,
			Notes:              req.Notes,
			CreatedBy:          claims.UserID,
			UpdatedBy:          claims.UserID,
		}

		if err := svc.RecordScreening(r.Context(), screening); err != nil {
			handleScreeningCreateError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toScreeningResponse(screening)) //nolint:errcheck // response write
	}
}

// HandleGetScreening returns the screening record for a session.
//
//	@Summary		Get screening
//	@Description	Returns the contraindication screening for a session.
//	@Tags			screening
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Session ID"
//	@Success		200	{object}	ScreeningResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/screening [get]
func HandleGetScreening(svc *service.ContraindicationService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sessionID, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.SessionInvalidData, "invalid session ID")
			return
		}

		screening, err := svc.GetBySessionID(r.Context(), sessionID)
		if err != nil {
			handleScreeningLookupError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toScreeningResponse(screening)) //nolint:errcheck // response write
	}
}

// HandleUpdateScreening updates a screening record for a session.
//
//	@Summary		Update screening
//	@Description	Updates the contraindication screening for a session. Requires version for optimistic locking.
//	@Tags			screening
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int							true	"Session ID"
//	@Param			request	body	updateScreeningRequest	true	"Updated screening details"
//	@Success		200		{object}	ScreeningResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/screening [put]
func HandleUpdateScreening(svc *service.ContraindicationService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateScreeningRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		screening := &domain.ContraindicationScreening{
			SessionID:          sessionID,
			Pregnant:           req.Pregnant,
			Breastfeeding:      req.Breastfeeding,
			ActiveInfection:    req.ActiveInfection,
			ActiveColdSores:    req.ActiveColdSores,
			Isotretinoin:       req.Isotretinoin,
			Photosensitivity:   req.Photosensitivity,
			AutoimmuneDisorder: req.AutoimmuneDisorder,
			KeloidHistory:      req.KeloidHistory,
			Anticoagulants:     req.Anticoagulants,
			RecentTan:          req.RecentTan,
			MitigationNotes:    req.MitigationNotes,
			Notes:              req.Notes,
			Version:            req.Version,
			UpdatedBy:          claims.UserID,
		}

		if err := svc.UpdateScreening(r.Context(), screening); err != nil {
			handleScreeningUpdateError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetBySessionID(r.Context(), sessionID)
		if err != nil {
			slog.Error("failed to fetch updated screening", "session_id", sessionID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ScreeningCreationFailed, "screening updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toScreeningResponse(updated)) //nolint:errcheck // response write
	}
}

// toScreeningResponse converts a domain screening to an API response.
func toScreeningResponse(s *domain.ContraindicationScreening) ScreeningResponse {
	return ScreeningResponse{
		ID:                 s.ID,
		SessionID:          s.SessionID,
		Pregnant:           s.Pregnant,
		Breastfeeding:      s.Breastfeeding,
		ActiveInfection:    s.ActiveInfection,
		ActiveColdSores:    s.ActiveColdSores,
		Isotretinoin:       s.Isotretinoin,
		Photosensitivity:   s.Photosensitivity,
		AutoimmuneDisorder: s.AutoimmuneDisorder,
		KeloidHistory:      s.KeloidHistory,
		Anticoagulants:     s.Anticoagulants,
		RecentTan:          s.RecentTan,
		HasFlags:           s.HasFlags,
		MitigationNotes:    s.MitigationNotes,
		Notes:              s.Notes,
		Version:            s.Version,
		CreatedAt:          s.CreatedAt,
		CreatedBy:          s.CreatedBy,
		UpdatedAt:          s.UpdatedAt,
		UpdatedBy:          s.UpdatedBy,
	}
}

// handleScreeningCreateError maps service create errors to HTTP responses.
func handleScreeningCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrScreeningAlreadyExists):
		apierrors.WriteError(w, http.StatusConflict,
			apierrors.ScreeningAlreadyExists, "screening already exists for this session")
	case errors.Is(err, service.ErrInvalidScreeningData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ScreeningInvalidData, "invalid screening data: check required fields")
	default:
		slog.Error("failed to record screening", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ScreeningCreationFailed, "failed to record screening")
	}
}

// handleScreeningLookupError maps service lookup errors to HTTP responses.
func handleScreeningLookupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrScreeningNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.ScreeningNotFound, "screening not found")
	default:
		slog.Error("failed to get screening", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ScreeningCreationFailed, "failed to look up screening")
	}
}

// handleScreeningUpdateError maps service update errors to HTTP responses.
func handleScreeningUpdateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidScreeningData):
		apierrors.WriteError(w, http.StatusBadRequest,
			apierrors.ScreeningInvalidData, "invalid screening data: check required fields")
	case errors.Is(err, service.ErrScreeningNotFound):
		apierrors.WriteError(w, http.StatusNotFound,
			apierrors.ScreeningNotFound, "screening not found")
	default:
		slog.Error("failed to update screening", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError,
			apierrors.ScreeningCreationFailed, "failed to update screening")
	}
}
