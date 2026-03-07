package domain

import "time"

// Consent represents a consent record attached to a treatment session.
type Consent struct {
	ID             int64     `json:"id"`
	SessionID      int64     `json:"session_id"`
	ConsentType    string    `json:"consent_type"`
	ConsentMethod  string    `json:"consent_method"`
	ObtainedAt     time.Time `json:"obtained_at"`
	RisksDiscussed bool      `json:"risks_discussed"`
	Notes          *string   `json:"notes"`
	Version        int       `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      int64     `json:"created_by"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      int64     `json:"updated_by"`
}
