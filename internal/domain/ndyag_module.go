package domain

import "time"

// NdYAGModuleDetail represents the type-specific parameters for an Nd:YAG
// laser procedure module. Each detail row is linked to a session_modules row
// via ModuleID and references the device used.
type NdYAGModuleDetail struct {
	ID             int64      `json:"id"`
	ModuleID       int64      `json:"module_id"`
	DeviceID       int64      `json:"device_id"`
	HandpieceID    *int64     `json:"handpiece_id,omitempty"`
	Wavelength     *string    `json:"wavelength,omitempty"`
	SpotSize       *string    `json:"spot_size,omitempty"`
	Fluence        *float64   `json:"fluence,omitempty"`
	PulseDuration  *float64   `json:"pulse_duration,omitempty"`
	RepetitionRate *float64   `json:"repetition_rate,omitempty"`
	CoolingType    *string    `json:"cooling_type,omitempty"`
	TotalPulses    *int       `json:"total_pulses,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	Version        int        `json:"version"`
	CreatedAt      time.Time  `json:"created_at"`
	CreatedBy      int64      `json:"created_by"`
	UpdatedAt      time.Time  `json:"updated_at"`
	UpdatedBy      int64      `json:"updated_by"`
}
