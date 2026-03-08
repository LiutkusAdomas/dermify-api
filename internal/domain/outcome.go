package domain

import "time"

// Outcome status constants define the possible results of a treatment session.
const (
	OutcomeStatusCompleted = "completed"
	OutcomeStatusPartial   = "partial"
	OutcomeStatusAborted   = "aborted"
)

// SessionOutcome represents the recorded outcome of a treatment session,
// including the overall status, aftercare instructions, and any red flags.
// There is at most one outcome per session (singleton pattern like Consent).
type SessionOutcome struct {
	ID             int64      `json:"id"`
	SessionID      int64      `json:"session_id"`
	OutcomeStatus  string     `json:"outcome_status"`
	EndpointIDs    []int64    `json:"endpoint_ids,omitempty"`
	AftercareNotes *string    `json:"aftercare_notes,omitempty"`
	RedFlagsText   *string    `json:"red_flags_text,omitempty"`
	ContactInfo    *string    `json:"contact_info,omitempty"`
	FollowUpAt     *time.Time `json:"follow_up_at,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	Version        int        `json:"version"`
	CreatedAt      time.Time  `json:"created_at"`
	CreatedBy      int64      `json:"created_by"`
	UpdatedAt      time.Time  `json:"updated_at"`
	UpdatedBy      int64      `json:"updated_by"`
}
