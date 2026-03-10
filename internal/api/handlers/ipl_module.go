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

type createIPLModuleRequest struct {
	DeviceID       int64    `json:"device_id"`
	HandpieceID    *int64   `json:"handpiece_id,omitempty"`
	FilterBand     *string  `json:"filter_band,omitempty"`
	LightguideSize *string  `json:"lightguide_size,omitempty"`
	Fluence        *float64 `json:"fluence,omitempty"`
	PulseDuration  *float64 `json:"pulse_duration,omitempty"`
	PulseDelay     *float64 `json:"pulse_delay,omitempty"`
	PulseCount     *int     `json:"pulse_count,omitempty"`
	Passes         *int     `json:"passes,omitempty"`
	TotalPulses    *int     `json:"total_pulses,omitempty"`
	CoolingMode    *string  `json:"cooling_mode,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

type updateIPLModuleRequest struct {
	DeviceID       int64    `json:"device_id"`
	HandpieceID    *int64   `json:"handpiece_id,omitempty"`
	FilterBand     *string  `json:"filter_band,omitempty"`
	LightguideSize *string  `json:"lightguide_size,omitempty"`
	Fluence        *float64 `json:"fluence,omitempty"`
	PulseDuration  *float64 `json:"pulse_duration,omitempty"`
	PulseDelay     *float64 `json:"pulse_delay,omitempty"`
	PulseCount     *int     `json:"pulse_count,omitempty"`
	Passes         *int     `json:"passes,omitempty"`
	TotalPulses    *int     `json:"total_pulses,omitempty"`
	CoolingMode    *string  `json:"cooling_mode,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	Version        int      `json:"version"`
}

// HandleCreateIPLModule creates an IPL module detail for a session.
//
//	@Summary		Create IPL module
//	@Description	Creates an IPL (Intense Pulsed Light) module detail for a session.
//	@Tags			modules-ipl
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int						true	"Session ID"
//	@Param			request	body	createIPLModuleRequest	true	"IPL module details"
//	@Success		201		{object}	IPLModuleDetailResponse
//	@Failure		400		{object}	apierrors.ErrorResponse
//	@Failure		404		{object}	apierrors.ErrorResponse
//	@Failure		409		{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules/ipl [post]
func HandleCreateIPLModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req createIPLModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.IPLModuleDetail{
			DeviceID:       req.DeviceID,
			HandpieceID:    req.HandpieceID,
			FilterBand:     req.FilterBand,
			LightguideSize: req.LightguideSize,
			Fluence:        req.Fluence,
			PulseDuration:  req.PulseDuration,
			PulseDelay:     req.PulseDelay,
			PulseCount:     req.PulseCount,
			Passes:         req.Passes,
			TotalPulses:    req.TotalPulses,
			CoolingMode:    req.CoolingMode,
			Notes:          req.Notes,
		}

		result, err := svc.CreateIPLModule(r.Context(), sessionID, detail, claims.UserID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		m.IncrementEnergyModuleCreatedCount()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toIPLModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleGetIPLModule returns an IPL module detail by module ID.
//
//	@Summary		Get IPL module
//	@Description	Returns an IPL module detail by module ID.
//	@Tags			modules-ipl
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int	true	"Session ID"
//	@Param			moduleId	path	int	true	"Module ID"
//	@Success		200	{object}	IPLModuleDetailResponse
//	@Failure		400	{object}	apierrors.ErrorResponse
//	@Failure		404	{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules/ipl/{moduleId} [get]
func HandleGetIPLModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		moduleIDStr := chi.URLParam(r, "moduleId")
		moduleID, err := strconv.ParseInt(moduleIDStr, 10, 64)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ModuleInvalidData, "invalid module ID")
			return
		}

		result, err := svc.GetIPLModule(r.Context(), moduleID)
		if err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toIPLModuleDetailResponse(result)) //nolint:errcheck // response write
	}
}

// HandleUpdateIPLModule updates an IPL module detail.
//
//	@Summary		Update IPL module
//	@Description	Updates an IPL module detail. Requires version for optimistic locking.
//	@Tags			modules-ipl
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int						true	"Session ID"
//	@Param			moduleId	path	int						true	"Module ID"
//	@Param			request		body	updateIPLModuleRequest	true	"Updated IPL details"
//	@Success		200			{object}	IPLModuleDetailResponse
//	@Failure		400			{object}	apierrors.ErrorResponse
//	@Failure		404			{object}	apierrors.ErrorResponse
//	@Failure		409			{object}	apierrors.ErrorResponse
//	@Router			/sessions/{id}/modules/ipl/{moduleId} [put]
func HandleUpdateIPLModule(svc *service.EnergyModuleService, m *metrics.Client) func(http.ResponseWriter, *http.Request) {
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

		var req updateIPLModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest,
				apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		detail := &domain.IPLModuleDetail{
			ID:             moduleID,
			DeviceID:       req.DeviceID,
			HandpieceID:    req.HandpieceID,
			FilterBand:     req.FilterBand,
			LightguideSize: req.LightguideSize,
			Fluence:        req.Fluence,
			PulseDuration:  req.PulseDuration,
			PulseDelay:     req.PulseDelay,
			PulseCount:     req.PulseCount,
			Passes:         req.Passes,
			TotalPulses:    req.TotalPulses,
			CoolingMode:    req.CoolingMode,
			Notes:          req.Notes,
			Version:        req.Version,
		}

		if err := svc.UpdateIPLModule(r.Context(), detail, claims.UserID); err != nil {
			handleEnergyModuleError(w, err)
			return
		}

		// Re-fetch to get the full updated record.
		updated, err := svc.GetIPLModule(r.Context(), moduleID)
		if err != nil {
			slog.Error("failed to fetch updated IPL module", "module_id", moduleID, "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError,
				apierrors.ModuleDetailUpdateFailed, "module updated but failed to retrieve")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(toIPLModuleDetailResponse(updated)) //nolint:errcheck // response write
	}
}

// toIPLModuleDetailResponse converts a domain IPL module detail to an API response.
func toIPLModuleDetailResponse(d *domain.IPLModuleDetail) IPLModuleDetailResponse {
	return IPLModuleDetailResponse{
		ID:             d.ID,
		ModuleID:       d.ModuleID,
		DeviceID:       d.DeviceID,
		HandpieceID:    d.HandpieceID,
		FilterBand:     d.FilterBand,
		LightguideSize: d.LightguideSize,
		Fluence:        d.Fluence,
		PulseDuration:  d.PulseDuration,
		PulseDelay:     d.PulseDelay,
		PulseCount:     d.PulseCount,
		Passes:         d.Passes,
		TotalPulses:    d.TotalPulses,
		CoolingMode:    d.CoolingMode,
		Notes:          d.Notes,
		Version:        d.Version,
		CreatedAt:      d.CreatedAt,
		CreatedBy:      d.CreatedBy,
		UpdatedAt:      d.UpdatedAt,
		UpdatedBy:      d.UpdatedBy,
	}
}
