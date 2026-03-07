package domain

import "time"

// Device represents an energy-based treatment device in the registry.
type Device struct {
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	Manufacturer string      `json:"manufacturer"`
	Model        string      `json:"model"`
	DeviceType   string      `json:"device_type"`
	Active       bool        `json:"active"`
	CreatedAt    time.Time   `json:"created_at"`
	Handpieces   []Handpiece `json:"handpieces,omitempty"`
}

// Handpiece represents an interchangeable handpiece attached to a device.
type Handpiece struct {
	ID        int64     `json:"id"`
	DeviceID  int64     `json:"device_id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}
