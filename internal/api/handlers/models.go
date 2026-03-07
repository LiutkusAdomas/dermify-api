package handlers

import "time"

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message" example:"operation successful"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string `json:"status" example:"healthy"`
	Timestamp string `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Version   string `json:"version" example:"1.0.0"`
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"johndoe"`
	Email    string `json:"email" example:"johndoe@example.com"`
}

// CreateUserResponse represents the response after creating a user.
type CreateUserResponse struct {
	ID       int    `json:"id" example:"3"`
	Username string `json:"username" example:"newuser"`
	Email    string `json:"email" example:"newuser@example.com"`
	Message  string `json:"message" example:"User created successfully"`
}

// UpdateUserResponse represents the response after updating a user.
type UpdateUserResponse struct {
	ID       string `json:"id" example:"1"`
	Username string `json:"username" example:"updateduser"`
	Email    string `json:"email" example:"updated@example.com"`
	Message  string `json:"message" example:"User updated successfully"`
}

// PatientResponse represents a patient in API responses.
type PatientResponse struct {
	ID                int64      `json:"id" example:"1"`
	FirstName         string     `json:"first_name" example:"Jane"`
	LastName          string     `json:"last_name" example:"Doe"`
	DateOfBirth       string     `json:"date_of_birth" example:"1990-01-15"`
	Sex               string     `json:"sex" example:"female"`
	Phone             *string    `json:"phone,omitempty" example:"+1234567890"`
	Email             *string    `json:"email,omitempty" example:"jane@example.com"`
	ExternalReference *string    `json:"external_reference,omitempty" example:"EXT-001"`
	Version           int        `json:"version" example:"1"`
	SessionCount      int        `json:"session_count" example:"0"`
	LastSessionDate   *time.Time `json:"last_session_date"`
	CreatedAt         time.Time  `json:"created_at"`
	CreatedBy         int64      `json:"created_by" example:"1"`
	UpdatedAt         time.Time  `json:"updated_at"`
	UpdatedBy         int64      `json:"updated_by" example:"1"`
}

// PaginatedResponse wraps a list response with pagination metadata.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total" example:"50"`
	Page       int         `json:"page" example:"1"`
	PerPage    int         `json:"per_page" example:"20"`
	TotalPages int         `json:"total_pages" example:"3"`
}
