package domain

import "time"

// Notification channel constants.
const (
	NotificationChannelSMS   = "sms"
	NotificationChannelEmail = "email"
)

// Notification type constants.
const (
	NotificationTypeAppointmentConfirmation = "appointment_confirmation"
	NotificationTypeAppointmentReminder     = "appointment_reminder"
	NotificationTypeAppointmentCancellation = "appointment_cancellation"
	NotificationTypeAppointmentReschedule   = "appointment_reschedule"
)

// Notification status constants.
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)

// Notification represents a message sent to a patient.
type Notification struct {
	ID          int64      `json:"id"`
	OrgID       int64      `json:"org_id"`
	PatientID   int64      `json:"patient_id"`
	Channel     string     `json:"channel"`
	Type        string     `json:"type"`
	Recipient   string     `json:"recipient"`
	Subject     string     `json:"subject"`
	Body        string     `json:"body"`
	Status      string     `json:"status"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	Error       string     `json:"error,omitempty"`
	ReferenceID int64      `json:"reference_id"`
	CreatedAt   time.Time  `json:"created_at"`
}
