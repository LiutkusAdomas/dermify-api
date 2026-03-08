package domain

import (
	"encoding/json"
	"time"
)

// InjectionSite represents a single injection point with anatomical location
// and the number of units delivered.
type InjectionSite struct {
	Site  string  `json:"site"`
	Units float64 `json:"units"`
}

// BotulinumModuleDetail represents the type-specific parameters for a
// botulinum toxin procedure module. Each detail row is linked to a
// session_modules row via ModuleID and references the product used.
type BotulinumModuleDetail struct {
	ID                     int64            `json:"id"`
	ModuleID               int64            `json:"module_id"`
	ProductID              int64            `json:"product_id"`
	BatchNumber            *string          `json:"batch_number,omitempty"`
	ExpiryDate             *time.Time       `json:"expiry_date,omitempty"`
	Diluent                *string          `json:"diluent,omitempty"`
	DilutionVolume         *float64         `json:"dilution_volume,omitempty"`
	ResultingConcentration *string          `json:"resulting_concentration,omitempty"`
	TotalUnits             *float64         `json:"total_units,omitempty"`
	InjectionSites         json.RawMessage  `json:"injection_sites,omitempty"`
	Notes                  *string          `json:"notes,omitempty"`
	Version                int              `json:"version"`
	CreatedAt              time.Time        `json:"created_at"`
	CreatedBy              int64            `json:"created_by"`
	UpdatedAt              time.Time        `json:"updated_at"`
	UpdatedBy              int64            `json:"updated_by"`
}
