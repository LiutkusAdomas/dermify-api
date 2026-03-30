package domain

import "time"

// AppointmentStatus constants define the lifecycle of an appointment.
const (
	AppointmentStatusScheduled  = "scheduled"
	AppointmentStatusConfirmed  = "confirmed"
	AppointmentStatusCheckedIn  = "checked_in"
	AppointmentStatusInProgress = "in_progress"
	AppointmentStatusCompleted  = "completed"
	AppointmentStatusCancelled  = "cancelled"
	AppointmentStatusNoShow     = "no_show"
)

// ValidAppointmentStatus checks if a status is valid.
func ValidAppointmentStatus(s string) bool {
	switch s {
	case AppointmentStatusScheduled, AppointmentStatusConfirmed, AppointmentStatusCheckedIn,
		AppointmentStatusInProgress, AppointmentStatusCompleted, AppointmentStatusCancelled,
		AppointmentStatusNoShow:
		return true
	}
	return false
}

// Appointment represents a scheduled visit at the clinic.
type Appointment struct {
	ID                 int64     `json:"id"`
	OrgID              int64     `json:"org_id"`
	PatientID          int64     `json:"patient_id"`
	DoctorID           int64     `json:"doctor_id"`
	ServiceTypeID      int64     `json:"service_type_id"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	Status             string    `json:"status"`
	Notes              string    `json:"notes"`
	CancellationReason string    `json:"cancellation_reason,omitempty"`
	SessionID          *int64    `json:"session_id,omitempty"`
	CreatedBy          int64     `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Version            int       `json:"version"`

	PatientFirstName string `json:"patient_first_name,omitempty"`
	PatientLastName  string `json:"patient_last_name,omitempty"`
	DoctorName       string `json:"doctor_name,omitempty"`
	ServiceTypeName  string `json:"service_type_name,omitempty"`
}
