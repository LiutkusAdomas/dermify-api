package domain

import "time"

// Module type constants define the available procedure module types.
const (
	ModuleTypeIPL       = "ipl"
	ModuleTypeNdYAG     = "ndyag"
	ModuleTypeCO2       = "co2"
	ModuleTypeRF        = "rf"
	ModuleTypeFiller    = "filler"
	ModuleTypeBotulinum = "botulinum_toxin"
)

// SessionModule represents a procedure module slot within a treatment session.
// This is the polymorphic base that future phases extend with type-specific
// detail tables.
type SessionModule struct {
	ID         int64     `json:"id"`
	SessionID  int64     `json:"session_id"`
	ModuleType string    `json:"module_type"`
	SortOrder  int       `json:"sort_order"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  int64     `json:"created_by"`
	UpdatedAt  time.Time `json:"updated_at"`
	UpdatedBy  int64     `json:"updated_by"`
}
