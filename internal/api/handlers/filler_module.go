package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"dermify-api/internal/api/apierrors"
	"dermify-api/internal/api/metrics"
	"dermify-api/internal/api/middleware"
	"dermify-api/internal/domain"
	"dermify-api/internal/service"

	"github.com/go-chi/chi/v5"
)

type createFillerModuleRequest struct {
	ProductID       int64      `json:"product_id"`
	BatchNumber     *string    `json:"batch_number,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	SyringeVolume   *float64   `json:"syringe_volume,omitempty"`
	TotalVolume     *float64   `json:"total_volume,omitempty"`
	NeedleType      *string    `json:"needle_type,omitempty"`
	InjectionPlane  *string    `json:"injection_plane,omitempty"`
	AnatomicalSites *string    `json:"anatomical_sites,omitempty"`
	Endpoint        *string    `json:"endpoint,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
}

type updateFillerModuleRequest struct {
	ProductID       int64      `json:"product_id"`
	BatchNumber     *string    `json:"batch_number,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	SyringeVolume   *float64   `json:"syringe_volume,omitempty"`
	TotalVolume     *float64   `json:"total_volume,omitempty"`
	NeedleType      *string    `json:"needle_type,omitempty"`
	InjectionPlane  *string    `json:"injection_plane,omitempty"`
	AnatomicalSites *string    `json:"anatomical_sites,omitempty"`
	Endpoint        *string    `json:"endpoint,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	Version         int        `json:"version"`
}

// HandleCreateFillerModule creates a filler module detail for a session.
//
//	@Summary		Create filler module
//	@Description	Creates a dermal filler module detail for a session.
//	@Tags			modules-filler
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int							true	"Session ID"
//	@Param			request	body	createFillerModuleRequest	true	"Filler module details"
//	@Success		201		{object}	FillerModuleDetailResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules/filler [post]
func HandleCreateFillerModule(svc *service.InjectableModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req createFillerModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.FillerModuleDetail{
			ProductID:       req.ProductID,
			BatchNumber:     req.BatchNumber,
			ExpiryDate:      req.ExpiryDate,
			SyringeVolume:   req.SyringeVolume,
			TotalVolume:     req.TotalVolume,
			NeedleType:      req.NeedleType,
			InjectionPlane:  req.InjectionPlane,
			AnatomicalSites: req.AnatomicalSites,
			Endpoint:        req.Endpoint,
			Notes:           req.Notes,
		}

		result, err := svc.CreateFillerModule(r.Context(), sessionID, detail, claims.UserID)
		if err != nil {
			handleInjectableModuleError(w, err)
			return
		}

		m.IncrementInjectableModuleCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toFillerModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleGetFillerModule returns a filler module detail by module ID.
//
//	@Summary		Get filler module
//	@Description	Returns a filler module detail by module ID.
//	@Tags			modules-filler
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int	true	"Session ID"
//	@Param			moduleId	path	int	true	"Module ID"
//	@Success		200	{object}	FillerModuleDetailResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules/filler/{moduleId} [get]
func HandleGetFillerModule(svc *service.InjectableModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		result, err := svc.GetFillerModule(r.Context(), moduleID)
		if err != nil {
			handleInjectableModuleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toFillerModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleUpdateFillerModule updates a filler module detail.
//
//	@Summary		Update filler module
//	@Description	Updates a filler module detail. Requires version for optimistic locking.
//	@Tags			modules-filler
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int							true	"Session ID"
//	@Param			moduleId	path	int							true	"Module ID"
//	@Param			request		body	updateFillerModuleRequest	true	"Updated filler details"
//	@Success		200			{object}	FillerModuleDetailResponse
//	@Failure		400			{object}	apierrors.ErrorResponse
//	@Failure		404			{object}	apierrors.ErrorResponse
//	@Failure		409			{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules/filler/{moduleId} [put]
func HandleUpdateFillerModule(svc *service.InjectableModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		if claims == nil {
			apierrors.WriteError(w, http.StatusUnauthorized,
				apierrors.AuthNotAuthenticated, "not authenticated")
			return
		}

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		var req updateFillerModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.FillerModuleDetail{
			ID:              moduleID,
			ProductID:       req.ProductID,
			BatchNumber:     req.BatchNumber,
			ExpiryDate:      req.ExpiryDate,
			SyringeVolume:   req.SyringeVolume,
			TotalVolume:     req.TotalVolume,
			NeedleType:      req.NeedleType,
			InjectionPlane:  req.InjectionPlane,
			AnatomicalSites: req.AnatomicalSites,
			Endpoint:        req.Endpoint,
			Notes:           req.Notes,
			Version:         req.Version,
		}

		if err := svc.UpdateFillerModule(r.Context(), detail, claims.UserID); err != nil {
			handleInjectableModuleError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetFillerModule(r.Context(), moduleID)
		if err != nil {
			slog.Error("failed to fetch updated filler module", "module_id", moduleID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ModuleDetailUpdateFailed, "module updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toFillerModuleDetailResponse(updated)) //nolint:errcheck // response write
	}
}

// toFillerModuleDetailResponse converts a domain filler module detail to an API response.
func toFillerModuleDetailResponse(d *domain.FillerModuleDetail) FillerModuleDetailResponse {
	return FillerModuleDetailResponse{
		ID:              d.ID,
		ModuleID:        d.ModuleID,
		ProductID:       d.ProductID,
		BatchNumber:     d.BatchNumber,
		ExpiryDate:      d.ExpiryDate,
		SyringeVolume:   d.SyringeVolume,
		TotalVolume:     d.TotalVolume,
		NeedleType:      d.NeedleType,
		InjectionPlane:  d.InjectionPlane,
		AnatomicalSites: d.AnatomicalSites,
		Endpoint:        d.Endpoint,
		Notes:           d.Notes,
		Version:         d.Version,
		CreatedAt:       d.CreatedAt,
		CreatedBy:       d.CreatedBy,
		UpdatedAt:       d.UpdatedAt,
		UpdatedBy:       d.UpdatedBy,
	}
}
