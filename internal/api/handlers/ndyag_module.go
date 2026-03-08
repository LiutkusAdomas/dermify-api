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

type createNdYAGModuleRequest struct {
	DeviceID       int64    `json:"device_id"`
	HandpieceID    *int64   `json:"handpiece_id,omitempty"`
	Wavelength     *string  `json:"wavelength,omitempty"`
	SpotSize       *string  `json:"spot_size,omitempty"`
	Fluence        *float64 `json:"fluence,omitempty"`
	PulseDuration  *float64 `json:"pulse_duration,omitempty"`
	RepetitionRate *float64 `json:"repetition_rate,omitempty"`
	CoolingType    *string  `json:"cooling_type,omitempty"`
	TotalPulses    *int     `json:"total_pulses,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

type updateNdYAGModuleRequest struct {
	DeviceID       int64    `json:"device_id"`
	HandpieceID    *int64   `json:"handpiece_id,omitempty"`
	Wavelength     *string  `json:"wavelength,omitempty"`
	SpotSize       *string  `json:"spot_size,omitempty"`
	Fluence        *float64 `json:"fluence,omitempty"`
	PulseDuration  *float64 `json:"pulse_duration,omitempty"`
	RepetitionRate *float64 `json:"repetition_rate,omitempty"`
	CoolingType    *string  `json:"cooling_type,omitempty"`
	TotalPulses    *int     `json:"total_pulses,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	Version        int      `json:"version"`
}

// HandleCreateNdYAGModule creates an Nd:YAG module detail for a session.
func HandleCreateNdYAGModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req createNdYAGModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.NdYAGModuleDetail{
			DeviceID:       req.DeviceID,
			HandpieceID:    req.HandpieceID,
			Wavelength:     req.Wavelength,
			SpotSize:       req.SpotSize,
			Fluence:        req.Fluence,
			PulseDuration:  req.PulseDuration,
			RepetitionRate: req.RepetitionRate,
			CoolingType:    req.CoolingType,
			TotalPulses:    req.TotalPulses,
			Notes:          req.Notes,
		}

		result, err := svc.CreateNdYAGModule(r.Context(), sessionID, detail, claims.UserID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		m.IncrementEnergyModuleCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toNdYAGModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleGetNdYAGModule returns an Nd:YAG module detail by module ID.
func HandleGetNdYAGModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		result, err := svc.GetNdYAGModule(r.Context(), moduleID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toNdYAGModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleUpdateNdYAGModule updates an Nd:YAG module detail.
func HandleUpdateNdYAGModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateNdYAGModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.NdYAGModuleDetail{
			ID:             moduleID,
			DeviceID:       req.DeviceID,
			HandpieceID:    req.HandpieceID,
			Wavelength:     req.Wavelength,
			SpotSize:       req.SpotSize,
			Fluence:        req.Fluence,
			PulseDuration:  req.PulseDuration,
			RepetitionRate: req.RepetitionRate,
			CoolingType:    req.CoolingType,
			TotalPulses:    req.TotalPulses,
			Notes:          req.Notes,
			Version:        req.Version,
		}

		if err := svc.UpdateNdYAGModule(r.Context(), detail, claims.UserID); err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetNdYAGModule(r.Context(), moduleID)
		if err != nil {
			slog.Error("failed to fetch updated NdYAG module", "module_id", moduleID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ModuleDetailUpdateFailed, "module updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toNdYAGModuleDetailResponse(updated)) //nolint:errcheck // response write
	}
}

// toNdYAGModuleDetailResponse converts a domain Nd:YAG module detail to an API response.
func toNdYAGModuleDetailResponse(d *domain.NdYAGModuleDetail) NdYAGModuleDetailResponse {
	return NdYAGModuleDetailResponse{
		ID:             d.ID,
		ModuleID:       d.ModuleID,
		DeviceID:       d.DeviceID,
		HandpieceID:    d.HandpieceID,
		Wavelength:     d.Wavelength,
		SpotSize:       d.SpotSize,
		Fluence:        d.Fluence,
		PulseDuration:  d.PulseDuration,
		RepetitionRate: d.RepetitionRate,
		CoolingType:    d.CoolingType,
		TotalPulses:    d.TotalPulses,
		Notes:          d.Notes,
		Version:        d.Version,
		CreatedAt:      d.CreatedAt,
		CreatedBy:      d.CreatedBy,
		UpdatedAt:      d.UpdatedAt,
		UpdatedBy:      d.UpdatedBy,
	}
}
