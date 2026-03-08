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

// SessionResponse represents a session in API responses.
type SessionResponse struct {
	ID               int64      `json:"id" example:"1"`
	PatientID        int64      `json:"patient_id" example:"1"`
	ClinicianID      int64      `json:"clinician_id" example:"1"`
	Status           string     `json:"status" example:"draft"`
	ScheduledAt      *time.Time `json:"scheduled_at"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	PatientGoal      *string    `json:"patient_goal"`
	FitzpatrickType  *int       `json:"fitzpatrick_type"`
	IsTanned         bool       `json:"is_tanned"`
	IsPregnant       bool       `json:"is_pregnant"`
	OnAnticoagulants bool       `json:"on_anticoagulants"`
	PhotoConsent     *string    `json:"photo_consent"`
	Notes            *string    `json:"notes"`
	IndicationCodes  []int64    `json:"indication_code_ids,omitempty"`
	Version          int        `json:"version" example:"1"`
	CreatedAt        time.Time  `json:"created_at"`
	CreatedBy        int64      `json:"created_by" example:"1"`
	UpdatedAt        time.Time  `json:"updated_at"`
	UpdatedBy        int64      `json:"updated_by" example:"1"`
}

// ConsentResponse represents a consent record in API responses.
type ConsentResponse struct {
	ID             int64     `json:"id" example:"1"`
	SessionID      int64     `json:"session_id" example:"1"`
	ConsentType    string    `json:"consent_type" example:"informed"`
	ConsentMethod  string    `json:"consent_method" example:"verbal"`
	ObtainedAt     time.Time `json:"obtained_at"`
	RisksDiscussed bool      `json:"risks_discussed" example:"true"`
	Notes          *string   `json:"notes"`
	Version        int       `json:"version" example:"1"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      int64     `json:"created_by" example:"1"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      int64     `json:"updated_by" example:"1"`
}

// ScreeningResponse represents a contraindication screening in API responses.
type ScreeningResponse struct {
	ID                 int64     `json:"id" example:"1"`
	SessionID          int64     `json:"session_id" example:"1"`
	Pregnant           bool      `json:"pregnant"`
	Breastfeeding      bool      `json:"breastfeeding"`
	ActiveInfection    bool      `json:"active_infection"`
	ActiveColdSores    bool      `json:"active_cold_sores"`
	Isotretinoin       bool      `json:"isotretinoin"`
	Photosensitivity   bool      `json:"photosensitivity"`
	AutoimmuneDisorder bool      `json:"autoimmune_disorder"`
	KeloidHistory      bool      `json:"keloid_history"`
	Anticoagulants     bool      `json:"anticoagulants"`
	RecentTan          bool      `json:"recent_tan"`
	HasFlags           bool      `json:"has_flags"`
	MitigationNotes    *string   `json:"mitigation_notes"`
	Notes              *string   `json:"notes"`
	Version            int       `json:"version" example:"1"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          int64     `json:"created_by" example:"1"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          int64     `json:"updated_by" example:"1"`
}

// ModuleResponse represents a session module in API responses.
type ModuleResponse struct {
	ID         int64     `json:"id" example:"1"`
	SessionID  int64     `json:"session_id" example:"1"`
	ModuleType string    `json:"module_type" example:"ipl"`
	SortOrder  int       `json:"sort_order" example:"1"`
	Version    int       `json:"version" example:"1"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  int64     `json:"created_by" example:"1"`
	UpdatedAt  time.Time `json:"updated_at"`
	UpdatedBy  int64     `json:"updated_by" example:"1"`
}
