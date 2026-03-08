package domain

import "time"

// CO2ModuleDetail represents the type-specific parameters for a CO2
// fractional laser procedure module. Each detail row is linked to a
// session_modules row via ModuleID and references the device used.
type CO2ModuleDetail struct {
	ID              int64      `json:"id"`
	ModuleID        int64      `json:"module_id"`
	DeviceID        int64      `json:"device_id"`
	HandpieceID     *int64     `json:"handpiece_id,omitempty"`
	Mode            *string    `json:"mode,omitempty"`
	ScannerPattern  *string    `json:"scanner_pattern,omitempty"`
	Power           *float64   `json:"power,omitempty"`
	PulseEnergy     *float64   `json:"pulse_energy,omitempty"`
	PulseDuration   *float64   `json:"pulse_duration,omitempty"`
	Density         *float64   `json:"density,omitempty"`
	Pattern         *string    `json:"pattern,omitempty"`
	Passes          *int       `json:"passes,omitempty"`
	AnaesthesiaUsed *string    `json:"anaesthesia_used,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	Version         int        `json:"version"`
	CreatedAt       time.Time  `json:"created_at"`
	CreatedBy       int64      `json:"created_by"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UpdatedBy       int64      `json:"updated_by"`
}
