package domain

import "time"

// Patient represents a patient record with demographics and metadata.
type Patient struct {
	ID                int64     `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	DateOfBirth       time.Time `json:"date_of_birth"`
	Sex               string    `json:"sex"`
	Phone             *string   `json:"phone"`
	Email             *string   `json:"email"`
	ExternalReference *string   `json:"external_reference"`
	Version           int       `json:"version"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         int64     `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         int64     `json:"updated_by"`
}

// SessionSummary represents a minimal view of a patient session.
// This is a placeholder for Phase 2 when sessions are fully implemented.
type SessionSummary struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}
