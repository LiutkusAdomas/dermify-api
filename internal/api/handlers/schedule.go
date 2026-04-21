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

type workingHoursEntry struct {
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type setWorkingHoursRequest struct {
	Hours []workingHoursEntry `json:"hours"`
}

type createOverrideRequest struct {
	Date      string  `json:"date"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Reason    string  `json:"reason"`
}

// HandleSetWorkingHours bulk-upserts a doctor's weekly working hours.
func HandleSetWorkingHours(svc *service.ScheduleService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		doctorID, err := parseNamedIDParam(r, "doctorId")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid doctor id")
			return
		}

		var req setWorkingHoursRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		hours := make([]domain.WorkingHours, len(req.Hours))
		for i, h := range req.Hours {
			hours[i] = domain.WorkingHours{
				DayOfWeek: h.DayOfWeek,
				StartTime: h.StartTime,
				EndTime:   h.EndTime,
			}
		}

		if err := svc.SetWorkingHours(r.Context(), membership.OrgID, doctorID, hours); err != nil {
			handleScheduleError(w, err)
			return
		}

		result, err := svc.GetWorkingHours(r.Context(), membership.OrgID, doctorID)
		if err != nil {
			handleScheduleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result) //nolint:errcheck // response write
	}
}

// HandleGetWorkingHours returns a doctor's weekly schedule.
func HandleGetWorkingHours(svc *service.ScheduleService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		doctorID, err := parseNamedIDParam(r, "doctorId")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid doctor id")
			return
		}

		hours, err := svc.GetWorkingHours(r.Context(), membership.OrgID, doctorID)
		if err != nil {
			handleScheduleError(w, err)
			return
		}

		if hours == nil {
			hours = []domain.WorkingHours{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(hours) //nolint:errcheck // response write
	}
}

// HandleCreateOverride creates a schedule override.
func HandleCreateOverride(svc *service.ScheduleService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		doctorID, err := parseNamedIDParam(r, "doctorId")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid doctor id")
			return
		}

		var req createOverrideRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid request body")
			return
		}

		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ScheduleInvalidData, "invalid date format, use YYYY-MM-DD")
			return
		}

		override := &domain.ScheduleOverride{
			OrgID:     membership.OrgID,
			DoctorID:  doctorID,
			Date:      date,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			Reason:    req.Reason,
		}

		if err := svc.CreateOverride(r.Context(), override); err != nil {
			handleScheduleError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(override) //nolint:errcheck // response write
	}
}

// HandleDeleteOverride removes a schedule override.
func HandleDeleteOverride(svc *service.ScheduleService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		id, err := parseIDParam(r)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid override id")
			return
		}

		if err := svc.DeleteOverride(r.Context(), id, membership.OrgID); err != nil {
			handleScheduleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "override deleted"}) //nolint:errcheck // response write
	}
}

// HandleListOverrides lists schedule overrides for a doctor.
func HandleListOverrides(svc *service.ScheduleService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		doctorID, err := parseNamedIDParam(r, "doctorId")
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationInvalidRequestBody, "invalid doctor id")
			return
		}

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")

		from := time.Now()
		to := from.AddDate(0, 3, 0)

		if fromStr != "" {
			if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
				from = parsed
			}
		}
		if toStr != "" {
			if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
				to = parsed
			}
		}

		overrides, err := svc.ListOverrides(r.Context(), membership.OrgID, doctorID, from, to)
		if err != nil {
			handleScheduleError(w, err)
			return
		}

		if overrides == nil {
			overrides = []*domain.ScheduleOverride{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(overrides) //nolint:errcheck // response write
	}
}

// HandleGetAvailability returns available time slots for a doctor on a date.
// The `date` query parameter is interpreted in the `timezone` query parameter if supplied,
// otherwise in the organization's configured timezone (falling back to UTC).
func HandleGetAvailability(scheduleSvc *service.ScheduleService, appointmentSvc *service.AppointmentService, orgSvc *service.OrganizationService, m *metrics.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		membership := middleware.GetOrgMembership(r.Context())
		if membership == nil {
			apierrors.WriteError(w, http.StatusForbidden, apierrors.OrgNotMember, "not a member")
			return
		}

		doctorID := parseInt64Param(r, "doctor_id", 0)
		if doctorID <= 0 {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "doctor_id is required")
			return
		}

		dateStr := r.URL.Query().Get("date")
		if dateStr == "" {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ValidationRequiredFields, "date is required")
			return
		}
		tzQuery := r.URL.Query().Get("timezone")
		if tzQuery == "" {
			if org, err := orgSvc.GetByID(r.Context(), membership.OrgID); err == nil && org != nil {
				tzQuery = org.Timezone
			}
		}
		loc, err := parseTimezoneOrUTC(tzQuery)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ScheduleInvalidData, "invalid timezone")
			return
		}
		date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			apierrors.WriteError(w, http.StatusBadRequest, apierrors.ScheduleInvalidData, "invalid date format")
			return
		}

		duration := parseIntParam(r, "duration", 30)

		existing, err := appointmentSvc.GetTimeSlotsForDate(r.Context(), membership.OrgID, doctorID, date)
		if err != nil {
			slog.Error("failed to get existing appointments", "error", err)
			apierrors.WriteError(w, http.StatusInternalServerError, apierrors.AppointmentLookupFailed, "failed to check existing appointments")
			return
		}

		slots, err := scheduleSvc.GetAvailableSlots(r.Context(), membership.OrgID, doctorID, date, duration, existing)
		if err != nil {
			handleScheduleError(w, err)
			return
		}

		if slots == nil {
			slots = []domain.TimeSlot{}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(slots) //nolint:errcheck // response write
	}
}

func handleScheduleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrScheduleInvalidData):
		apierrors.WriteError(w, http.StatusBadRequest, apierrors.ScheduleInvalidData, "invalid schedule data")
	case errors.Is(err, service.ErrScheduleOverrideNotFound):
		apierrors.WriteError(w, http.StatusNotFound, apierrors.ScheduleNotFound, "schedule override not found")
	default:
		slog.Error("schedule operation failed", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError, apierrors.ScheduleLookupFailed, "internal error")
	}
}

func parseTimezoneOrUTC(tz string) (*time.Location, error) {
	if tz == "" {
		return time.UTC, nil
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	return loc, nil
}
