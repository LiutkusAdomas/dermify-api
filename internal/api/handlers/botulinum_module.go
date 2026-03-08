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

type createBotulinumModuleRequest struct {
	ProductID              int64            `json:"product_id"`
	BatchNumber            *string          `json:"batch_number,omitempty"`
	ExpiryDate             *time.Time       `json:"expiry_date,omitempty"`
	Diluent                *string          `json:"diluent,omitempty"`
	DilutionVolume         *float64         `json:"dilution_volume,omitempty"`
	ResultingConcentration *string          `json:"resulting_concentration,omitempty"`
	TotalUnits             *float64         `json:"total_units,omitempty"`
	InjectionSites         json.RawMessage  `json:"injection_sites,omitempty"`
	Notes                  *string          `json:"notes,omitempty"`
}

type updateBotulinumModuleRequest struct {
	ProductID              int64            `json:"product_id"`
	BatchNumber            *string          `json:"batch_number,omitempty"`
	ExpiryDate             *time.Time       `json:"expiry_date,omitempty"`
	Diluent                *string          `json:"diluent,omitempty"`
	DilutionVolume         *float64         `json:"dilution_volume,omitempty"`
	ResultingConcentration *string          `json:"resulting_concentration,omitempty"`
	TotalUnits             *float64         `json:"total_units,omitempty"`
	InjectionSites         json.RawMessage  `json:"injection_sites,omitempty"`
	Notes                  *string          `json:"notes,omitempty"`
	Version                int              `json:"version"`
}

// HandleCreateBotulinumModule creates a botulinum module detail for a session.
func HandleCreateBotulinumModule(svc *service.InjectableModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req createBotulinumModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.BotulinumModuleDetail{
			ProductID:              req.ProductID,
			BatchNumber:            req.BatchNumber,
			ExpiryDate:             req.ExpiryDate,
			Diluent:                req.Diluent,
			DilutionVolume:         req.DilutionVolume,
			ResultingConcentration: req.ResultingConcentration,
			TotalUnits:             req.TotalUnits,
			InjectionSites:         req.InjectionSites,
			Notes:                  req.Notes,
		}

		result, err := svc.CreateBotulinumModule(r.Context(), sessionID, detail, claims.UserID)
		if err != nil {
			handleInjectableModuleError(w, err)
			return
		}

		m.IncrementInjectableModuleCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toBotulinumModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleGetBotulinumModule returns a botulinum module detail by module ID.
func HandleGetBotulinumModule(svc *service.InjectableModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		result, err := svc.GetBotulinumModule(r.Context(), moduleID)
		if err != nil {
			handleInjectableModuleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toBotulinumModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleUpdateBotulinumModule updates a botulinum module detail.
func HandleUpdateBotulinumModule(svc *service.InjectableModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateBotulinumModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.BotulinumModuleDetail{
			ID:                     moduleID,
			ProductID:              req.ProductID,
			BatchNumber:            req.BatchNumber,
			ExpiryDate:             req.ExpiryDate,
			Diluent:                req.Diluent,
			DilutionVolume:         req.DilutionVolume,
			ResultingConcentration: req.ResultingConcentration,
			TotalUnits:             req.TotalUnits,
			InjectionSites:         req.InjectionSites,
			Notes:                  req.Notes,
			Version:                req.Version,
		}

		if err := svc.UpdateBotulinumModule(r.Context(), detail, claims.UserID); err != nil {
			handleInjectableModuleError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetBotulinumModule(r.Context(), moduleID)
		if err != nil {
			slog.Error("failed to fetch updated botulinum module", "module_id", moduleID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ModuleDetailUpdateFailed, "module updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toBotulinumModuleDetailResponse(updated)) //nolint:errcheck // response write
	}
}

// toBotulinumModuleDetailResponse converts a domain botulinum module detail to an API response.
func toBotulinumModuleDetailResponse(d *domain.BotulinumModuleDetail) BotulinumModuleDetailResponse {
	return BotulinumModuleDetailResponse{
		ID:                     d.ID,
		ModuleID:               d.ModuleID,
		ProductID:              d.ProductID,
		BatchNumber:            d.BatchNumber,
		ExpiryDate:             d.ExpiryDate,
		Diluent:                d.Diluent,
		DilutionVolume:         d.DilutionVolume,
		ResultingConcentration: d.ResultingConcentration,
		TotalUnits:             d.TotalUnits,
		InjectionSites:         d.InjectionSites,
		Notes:                  d.Notes,
		Version:                d.Version,
		CreatedAt:              d.CreatedAt,
		CreatedBy:              d.CreatedBy,
		UpdatedAt:              d.UpdatedAt,
		UpdatedBy:              d.UpdatedBy,
	}
}
