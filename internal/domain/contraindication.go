package domain

import "time"

// ContraindicationScreening represents the contraindication checklist
// completed for a treatment session.
type ContraindicationScreening struct {
	ID                 int64     `json:"id"`
	SessionID          int64     `json:"session_id"`
	Pregnant           bool      `json:"pregnant"`
	Breastfeeding      bool      `json:"breastfeeding"`
	ActiveInfection    bool      `json:"active_infection"`
	ActiveColdSores    bool      `json:"active_cold_sores"`
	Isotretinoin       bool      `json:"isotretinoin"`
	Photosensitivity   bool      `json:"photosensitivity"`
	AutoimmuneDisorder bool      `json:"autoimmune_disorder"`
	KeloidHistory      bool      `json:"keloid_history"`
	Anticoagulants     bool      `json:"anticoagulants"`
	RecentTan          bool      `json:"recent_tan"`
	HasFlags           bool      `json:"has_flags"`
	MitigationNotes    *string   `json:"mitigation_notes"`
	Notes              *string   `json:"notes"`
	Version            int       `json:"version"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          int64     `json:"created_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          int64     `json:"updated_by"`
}
