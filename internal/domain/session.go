package domain

import "time"

// Session status constants define the lifecycle states of a treatment session.
const (
	SessionStatusDraft           = "draft"
	SessionStatusInProgress      = "in_progress"
	SessionStatusAwaitingSignoff = "awaiting_signoff"
	SessionStatusSigned          = "signed"
	SessionStatusLocked          = "locked"
)

// Fitzpatrick skin type range constants.
const (
	FitzpatrickMin = 1
	FitzpatrickMax = 6
)

// Photo consent constants.
const (
	PhotoConsentYes     = "yes"
	PhotoConsentNo      = "no"
	PhotoConsentLimited = "limited"
)

// Session represents a treatment session with header fields, lifecycle status,
// and audit metadata.
type Session struct {
	ID               int64      `json:"id"`
	PatientID        int64      `json:"patient_id"`
	ClinicianID      int64      `json:"clinician_id"`
	Status           string     `json:"status"`
	ScheduledAt      *time.Time `json:"scheduled_at"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	PatientGoal      *string    `json:"patient_goal"`
	FitzpatrickType  *int       `json:"fitzpatrick_type"`
	IsTanned         bool       `json:"is_tanned"`
	IsPregnant       bool       `json:"is_pregnant"`
	OnAnticoagulants bool       `json:"on_anticoagulants"`
	PhotoConsent     *string    `json:"photo_consent"`
	Notes            *string    `json:"notes"`
	IndicationCodes  []int64    `json:"indication_code_ids,omitempty"`
	Version          int        `json:"version"`
	CreatedAt        time.Time  `json:"created_at"`
	CreatedBy        int64      `json:"created_by"`
	UpdatedAt        time.Time  `json:"updated_at"`
	UpdatedBy        int64      `json:"updated_by"`
}
