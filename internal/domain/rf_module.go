package domain

import "time"

// RFModuleDetail represents the type-specific parameters for a
// radiofrequency (RF) microneedling procedure module. Each detail row is
// linked to a session_modules row via ModuleID and references the device used.
type RFModuleDetail struct {
	ID           int64      `json:"id"`
	ModuleID     int64      `json:"module_id"`
	DeviceID     int64      `json:"device_id"`
	HandpieceID  *int64     `json:"handpiece_id,omitempty"`
	RFMode       *string    `json:"rf_mode,omitempty"`
	TipType      *string    `json:"tip_type,omitempty"`
	Depth        *float64   `json:"depth,omitempty"`
	EnergyLevel  *float64   `json:"energy_level,omitempty"`
	Overlap      *float64   `json:"overlap,omitempty"`
	PulsesPerZone *int      `json:"pulses_per_zone,omitempty"`
	TotalPulses  *int       `json:"total_pulses,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	Version      int        `json:"version"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    int64      `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    int64      `json:"updated_by"`
}
