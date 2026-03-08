package handlers

import (
	"encoding/json"
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

type createRFModuleRequest struct {
	DeviceID      int64    `json:"device_id"`
	HandpieceID   *int64   `json:"handpiece_id,omitempty"`
	RFMode        *string  `json:"rf_mode,omitempty"`
	TipType       *string  `json:"tip_type,omitempty"`
	Depth         *float64 `json:"depth,omitempty"`
	EnergyLevel   *float64 `json:"energy_level,omitempty"`
	Overlap       *float64 `json:"overlap,omitempty"`
	PulsesPerZone *int     `json:"pulses_per_zone,omitempty"`
	TotalPulses   *int     `json:"total_pulses,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
}

type updateRFModuleRequest struct {
	DeviceID      int64    `json:"device_id"`
	HandpieceID   *int64   `json:"handpiece_id,omitempty"`
	RFMode        *string  `json:"rf_mode,omitempty"`
	TipType       *string  `json:"tip_type,omitempty"`
	Depth         *float64 `json:"depth,omitempty"`
	EnergyLevel   *float64 `json:"energy_level,omitempty"`
	Overlap       *float64 `json:"overlap,omitempty"`
	PulsesPerZone *int     `json:"pulses_per_zone,omitempty"`
	TotalPulses   *int     `json:"total_pulses,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
	Version       int      `json:"version"`
}

// HandleCreateRFModule creates an RF module detail for a session.
func HandleCreateRFModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req createRFModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.RFModuleDetail{
			DeviceID:      req.DeviceID,
			HandpieceID:   req.HandpieceID,
			RFMode:        req.RFMode,
			TipType:       req.TipType,
			Depth:         req.Depth,
			EnergyLevel:   req.EnergyLevel,
			Overlap:       req.Overlap,
			PulsesPerZone: req.PulsesPerZone,
			TotalPulses:   req.TotalPulses,
			Notes:         req.Notes,
		}

		result, err := svc.CreateRFModule(r.Context(), sessionID, detail, claims.UserID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		m.IncrementEnergyModuleCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toRFModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleGetRFModule returns an RF module detail by module ID.
func HandleGetRFModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		result, err := svc.GetRFModule(r.Context(), moduleID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toRFModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleUpdateRFModule updates an RF module detail.
func HandleUpdateRFModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateRFModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.RFModuleDetail{
			ID:            moduleID,
			DeviceID:      req.DeviceID,
			HandpieceID:   req.HandpieceID,
			RFMode:        req.RFMode,
			TipType:       req.TipType,
			Depth:         req.Depth,
			EnergyLevel:   req.EnergyLevel,
			Overlap:       req.Overlap,
			PulsesPerZone: req.PulsesPerZone,
			TotalPulses:   req.TotalPulses,
			Notes:         req.Notes,
			Version:       req.Version,
		}

		if err := svc.UpdateRFModule(r.Context(), detail, claims.UserID); err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetRFModule(r.Context(), moduleID)
		if err != nil {
			slog.Error("failed to fetch updated RF module", "module_id", moduleID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ModuleDetailUpdateFailed, "module updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toRFModuleDetailResponse(updated)) //nolint:errcheck // response write
	}
}

// toRFModuleDetailResponse converts a domain RF module detail to an API response.
func toRFModuleDetailResponse(d *domain.RFModuleDetail) RFModuleDetailResponse {
	return RFModuleDetailResponse{
		ID:            d.ID,
		ModuleID:      d.ModuleID,
		DeviceID:      d.DeviceID,
		HandpieceID:   d.HandpieceID,
		RFMode:        d.RFMode,
		TipType:       d.TipType,
		Depth:         d.Depth,
		EnergyLevel:   d.EnergyLevel,
		Overlap:       d.Overlap,
		PulsesPerZone: d.PulsesPerZone,
		TotalPulses:   d.TotalPulses,
		Notes:         d.Notes,
		Version:       d.Version,
		CreatedAt:     d.CreatedAt,
		CreatedBy:     d.CreatedBy,
		UpdatedAt:     d.UpdatedAt,
		UpdatedBy:     d.UpdatedBy,
	}
}
