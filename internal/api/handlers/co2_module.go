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

type createCO2ModuleRequest struct {
	DeviceID        int64    `json:"device_id"`
	HandpieceID     *int64   `json:"handpiece_id,omitempty"`
	Mode            *string  `json:"mode,omitempty"`
	ScannerPattern  *string  `json:"scanner_pattern,omitempty"`
	Power           *float64 `json:"power,omitempty"`
	PulseEnergy     *float64 `json:"pulse_energy,omitempty"`
	PulseDuration   *float64 `json:"pulse_duration,omitempty"`
	Density         *float64 `json:"density,omitempty"`
	Pattern         *string  `json:"pattern,omitempty"`
	Passes          *int     `json:"passes,omitempty"`
	AnaesthesiaUsed *string  `json:"anaesthesia_used,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
}

type updateCO2ModuleRequest struct {
	DeviceID        int64    `json:"device_id"`
	HandpieceID     *int64   `json:"handpiece_id,omitempty"`
	Mode            *string  `json:"mode,omitempty"`
	ScannerPattern  *string  `json:"scanner_pattern,omitempty"`
	Power           *float64 `json:"power,omitempty"`
	PulseEnergy     *float64 `json:"pulse_energy,omitempty"`
	PulseDuration   *float64 `json:"pulse_duration,omitempty"`
	Density         *float64 `json:"density,omitempty"`
	Pattern         *string  `json:"pattern,omitempty"`
	Passes          *int     `json:"passes,omitempty"`
	AnaesthesiaUsed *string  `json:"anaesthesia_used,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
	Version         int      `json:"version"`
}

// HandleCreateCO2Module creates a CO2 module detail for a session.
func HandleCreateCO2Module(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req createCO2ModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.CO2ModuleDetail{
			DeviceID:        req.DeviceID,
			HandpieceID:     req.HandpieceID,
			Mode:            req.Mode,
			ScannerPattern:  req.ScannerPattern,
			Power:           req.Power,
			PulseEnergy:     req.PulseEnergy,
			PulseDuration:   req.PulseDuration,
			Density:         req.Density,
			Pattern:         req.Pattern,
			Passes:          req.Passes,
			AnaesthesiaUsed: req.AnaesthesiaUsed,
			Notes:           req.Notes,
		}

		result, err := svc.CreateCO2Module(r.Context(), sessionID, detail, claims.UserID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		m.IncrementEnergyModuleCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toCO2ModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleGetCO2Module returns a CO2 module detail by module ID.
func HandleGetCO2Module(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		result, err := svc.GetCO2Module(r.Context(), moduleID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toCO2ModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleUpdateCO2Module updates a CO2 module detail.
func HandleUpdateCO2Module(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateCO2ModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.CO2ModuleDetail{
			ID:              moduleID,
			DeviceID:        req.DeviceID,
			HandpieceID:     req.HandpieceID,
			Mode:            req.Mode,
			ScannerPattern:  req.ScannerPattern,
			Power:           req.Power,
			PulseEnergy:     req.PulseEnergy,
			PulseDuration:   req.PulseDuration,
			Density:         req.Density,
			Pattern:         req.Pattern,
			Passes:          req.Passes,
			AnaesthesiaUsed: req.AnaesthesiaUsed,
			Notes:           req.Notes,
			Version:         req.Version,
		}

		if err := svc.UpdateCO2Module(r.Context(), detail, claims.UserID); err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetCO2Module(r.Context(), moduleID)
		if err != nil {
			slog.Error("failed to fetch updated CO2 module", "module_id", moduleID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ModuleDetailUpdateFailed, "module updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toCO2ModuleDetailResponse(updated)) //nolint:errcheck // response write
	}
}

// toCO2ModuleDetailResponse converts a domain CO2 module detail to an API response.
func toCO2ModuleDetailResponse(d *domain.CO2ModuleDetail) CO2ModuleDetailResponse {
	return CO2ModuleDetailResponse{
		ID:              d.ID,
		ModuleID:        d.ModuleID,
		DeviceID:        d.DeviceID,
		HandpieceID:     d.HandpieceID,
		Mode:            d.Mode,
		ScannerPattern:  d.ScannerPattern,
		Power:           d.Power,
		PulseEnergy:     d.PulseEnergy,
		PulseDuration:   d.PulseDuration,
		Density:         d.Density,
		Pattern:         d.Pattern,
		Passes:          d.Passes,
		AnaesthesiaUsed: d.AnaesthesiaUsed,
		Notes:           d.Notes,
		Version:         d.Version,
		CreatedAt:       d.CreatedAt,
		CreatedBy:       d.CreatedBy,
		UpdatedAt:       d.UpdatedAt,
		UpdatedBy:       d.UpdatedBy,
	}
}
