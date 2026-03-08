package domain

import "time"

// IPLModuleDetail represents the type-specific parameters for an IPL
// (Intense Pulsed Light) procedure module. Each detail row is linked to a
// session_modules row via ModuleID and references the device used.
type IPLModuleDetail struct {
	ID             int64      `json:"id"`
	ModuleID       int64      `json:"module_id"`
	DeviceID       int64      `json:"device_id"`
	HandpieceID    *int64     `json:"handpiece_id,omitempty"`
	FilterBand     *string    `json:"filter_band,omitempty"`
	LightguideSize *string    `json:"lightguide_size,omitempty"`
	Fluence        *float64   `json:"fluence,omitempty"`
	PulseDuration  *float64   `json:"pulse_duration,omitempty"`
	PulseDelay     *float64   `json:"pulse_delay,omitempty"`
	PulseCount     *int       `json:"pulse_count,omitempty"`
	Passes         *int       `json:"passes,omitempty"`
	TotalPulses    *int       `json:"total_pulses,omitempty"`
	CoolingMode    *string    `json:"cooling_mode,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	Version        int        `json:"version"`
	CreatedAt      time.Time  `json:"created_at"`
	CreatedBy      int64      `json:"created_by"`
	UpdatedAt      time.Time  `json:"updated_at"`
	UpdatedBy      int64      `json:"updated_by"`
}
