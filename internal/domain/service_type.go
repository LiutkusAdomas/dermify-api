package domain

import "time"

// ServiceType represents a type of service offered by the clinic (e.g. "Botox consultation").
type ServiceType struct {
	ID              int64     `json:"id"`
	OrgID           int64     `json:"org_id"`
	Name            string    `json:"name"`
	DefaultDuration int       `json:"default_duration_minutes"`
	Description     string    `json:"description"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
