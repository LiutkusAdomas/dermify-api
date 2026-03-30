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

type createAppointmentRequest struct {
	PatientID     int64  `json:"patient_id"`
	DoctorID      int64  `json:"doctor_id"`
	ServiceTypeID int64  `json:"service_type_id"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Notes         string `json:"notes"`
}

type updateAppointmentRequest struct {
	PatientID     int64  `json:"patient_id"`
	DoctorID      int64  `json:"doctor_id"`
	ServiceTypeID int64  `json:"service_type_id"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Notes         string `json:"notes"`
}

type updateStatusRequest struct {
	Status             string `json:"status"`
	CancellationReason string `json:"cancellation_reason"`
}

// HandleCreateAppointment creates a new appointment.
func HandleCreateAppointment(svc *service.AppointmentService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		membership := middleware.GetOrgMembership(r.Context())
		if claims == nil || membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		var req createAppointmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentInvalidData, "invalid start_time format, use RFC3339")
			return
		}
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentInvalidData, "invalid end_time format, use RFC3339")
			return
		}

		appt := &domain.Appointment{
			OrgID:         membership.OrgID,
			PatientID:     req.PatientID,
			DoctorID:      req.DoctorID,
			ServiceTypeID: req.ServiceTypeID,
			StartTime:     startTime,
			EndTime:       endTime,
			Notes:         req.Notes,
			CreatedBy:     claims.UserID,
		}

		if err := svc.Create(r.Context(), appt); err != nil {
			handleAppointmentError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(appt) //nolint:errcheck // response write
	}
}

// HandleListAppointments lists appointments with filters.
func HandleListAppointments(svc *service.AppointmentService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		filter := service.AppointmentFilter{
			OrgID:     membership.OrgID,
			DoctorID:  parseInt64Param(r, "doctor_id", 0),
			PatientID: parseInt64Param(r, "patient_id", 0),
			Status:    r.URL.Query().Get("status"),
			Page:      parseIntParam(r, "page", 1),
			PerPage:   parseIntParam(r, "per_page", 100),
		}

		if startStr := r.URL.Query().Get("start"); startStr != "" {
			if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				filter.Start = t
			}
		}
		if endStr := r.URL.Query().Get("end"); endStr != "" {
			if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				filter.End = t
			}
		}

		result, err := svc.List(r.Context(), filter)
		if err != nil {
			slog.Error("failed to list appointments", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.AppointmentLookupFailed, "failed to query appointments")
			return
		}

		if result.Appointments == nil {
			result.Appointments = []domain.Appointment{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result) //nolint:errcheck // response write
	}
}

// HandleGetAppointment returns a single appointment.
func HandleGetAppointment(svc *service.AppointmentService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid appointment id")
			return
		}

		appt, err := svc.GetByID(r.Context(), id, membership.OrgID)
		if err != nil {
			handleAppointmentError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(appt) //nolint:errcheck // response write
	}
}

// HandleUpdateAppointment reschedules an appointment.
func HandleUpdateAppointment(svc *service.AppointmentService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid appointment id")
			return
		}

		var req updateAppointmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentInvalidData, "invalid start_time format")
			return
		}
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentInvalidData, "invalid end_time format")
			return
		}

		appt := &domain.Appointment{
			ID:            id,
			OrgID:         membership.OrgID,
			PatientID:     req.PatientID,
			DoctorID:      req.DoctorID,
			ServiceTypeID: req.ServiceTypeID,
			StartTime:     startTime,
			EndTime:       endTime,
			Notes:         req.Notes,
		}

		if err := svc.Update(r.Context(), appt); err != nil {
			handleAppointmentError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(appt) //nolint:errcheck // response write
	}
}

// HandleUpdateAppointmentStatus changes an appointment's status.
func HandleUpdateAppointmentStatus(svc *service.AppointmentService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid appointment id")
			return
		}

		var req updateStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		if !domain.ValidAppointmentStatus(req.Status) {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentInvalidData, "invalid status")
			return
		}

		if err := svc.UpdateStatus(r.Context(), id, membership.OrgID, req.Status, req.CancellationReason); err != nil {
			handleAppointmentError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "status updated"}) //nolint:errcheck // response write
	}
}

// HandleStartSessionFromAppointment creates a session from an appointment.
func HandleStartSessionFromAppointment(svc *service.AppointmentService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := middleware.GetUserClaims(r.Context())
		membership := middleware.GetOrgMembership(r.Context())
		if claims == nil || membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid appointment id")
			return
		}

		session, err := svc.StartSession(r.Context(), id, membership.OrgID, claims.UserID)
		if err != nil {
			handleAppointmentError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(session) //nolint:errcheck // response write
	}
}

func handleAppointmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAppointmentNotFound):
		apierrors.WriteError(w, http.StatusNotFound, apierrors.AppointmentNotFound, "appointment not found")
	case errors.Is(err, service.ErrAppointmentVersionConflict):
		apierrors.WriteError(w, http.StatusConflict, apierrors.AppointmentVersionConflict, "appointment was modified by another user")
	case errors.Is(err, service.ErrInvalidAppointmentData):
		apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentInvalidData, "invalid appointment data")
	case errors.Is(err, service.ErrAppointmentOverlap):
		apierrors.WriteError(w, http.StatusConflict, apierrors.AppointmentOverlap, "appointment overlaps with an existing one")
	case errors.Is(err, service.ErrAppointmentInvalidTransition):
		apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentInvalidTransition, "invalid status transition")
	case errors.Is(err, service.ErrAppointmentOutsideHours):
		apierrors.WriteError(w, http.StatusBadRequest, apierrors.AppointmentOutsideHours, "appointment is outside working hours")
	default:
		slog.Error("appointment operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError, apierrors.AppointmentLookupFailed, "internal error")
	}
}
