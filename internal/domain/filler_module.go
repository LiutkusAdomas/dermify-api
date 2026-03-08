package domain

import "time"

// FillerModuleDetail represents the type-specific parameters for a dermal
// filler procedure module. Each detail row is linked to a session_modules row
// via ModuleID and references the product used.
type FillerModuleDetail struct {
	ID              int64      `json:"id"`
	ModuleID        int64      `json:"module_id"`
	ProductID       int64      `json:"product_id"`
	BatchNumber     *string    `json:"batch_number,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	SyringeVolume   *float64   `json:"syringe_volume,omitempty"`
	TotalVolume     *float64   `json:"total_volume,omitempty"`
	NeedleType      *string    `json:"needle_type,omitempty"`
	InjectionPlane  *string    `json:"injection_plane,omitempty"`
	AnatomicalSites *string    `json:"anatomical_sites,omitempty"`
	Endpoint        *string    `json:"endpoint,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	Version         int        `json:"version"`
	CreatedAt       time.Time  `json:"created_at"`
	CreatedBy       int64      `json:"created_by"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UpdatedBy       int64      `json:"updated_by"`
}
