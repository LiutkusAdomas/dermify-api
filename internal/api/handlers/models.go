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

// IPLModuleDetailResponse represents an IPL module detail in API responses.
type IPLModuleDetailResponse struct {
	ID             int64     `json:"id" example:"1"`
	ModuleID       int64     `json:"module_id" example:"1"`
	DeviceID       int64     `json:"device_id" example:"1"`
	HandpieceID    *int64    `json:"handpiece_id,omitempty" example:"1"`
	FilterBand     *string   `json:"filter_band,omitempty" example:"560nm"`
	LightguideSize *string   `json:"lightguide_size,omitempty" example:"15x35mm"`
	Fluence        *float64  `json:"fluence,omitempty" example:"14.5"`
	PulseDuration  *float64  `json:"pulse_duration,omitempty" example:"20.0"`
	PulseDelay     *float64  `json:"pulse_delay,omitempty" example:"30.0"`
	PulseCount     *int      `json:"pulse_count,omitempty" example:"3"`
	Passes         *int      `json:"passes,omitempty" example:"2"`
	TotalPulses    *int      `json:"total_pulses,omitempty" example:"120"`
	CoolingMode    *string   `json:"cooling_mode,omitempty" example:"contact"`
	Notes          *string   `json:"notes,omitempty"`
	Version        int       `json:"version" example:"1"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      int64     `json:"created_by" example:"1"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      int64     `json:"updated_by" example:"1"`
}

// NdYAGModuleDetailResponse represents an Nd:YAG module detail in API responses.
type NdYAGModuleDetailResponse struct {
	ID             int64     `json:"id" example:"1"`
	ModuleID       int64     `json:"module_id" example:"1"`
	DeviceID       int64     `json:"device_id" example:"1"`
	HandpieceID    *int64    `json:"handpiece_id,omitempty" example:"1"`
	Wavelength     *string   `json:"wavelength,omitempty" example:"1064nm"`
	SpotSize       *string   `json:"spot_size,omitempty" example:"6mm"`
	Fluence        *float64  `json:"fluence,omitempty" example:"35.0"`
	PulseDuration  *float64  `json:"pulse_duration,omitempty" example:"10.0"`
	RepetitionRate *float64  `json:"repetition_rate,omitempty" example:"2.0"`
	CoolingType    *string   `json:"cooling_type,omitempty" example:"air"`
	TotalPulses    *int      `json:"total_pulses,omitempty" example:"200"`
	Notes          *string   `json:"notes,omitempty"`
	Version        int       `json:"version" example:"1"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      int64     `json:"created_by" example:"1"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      int64     `json:"updated_by" example:"1"`
}

// CO2ModuleDetailResponse represents a CO2 module detail in API responses.
type CO2ModuleDetailResponse struct {
	ID              int64     `json:"id" example:"1"`
	ModuleID        int64     `json:"module_id" example:"1"`
	DeviceID        int64     `json:"device_id" example:"1"`
	HandpieceID     *int64    `json:"handpiece_id,omitempty" example:"1"`
	Mode            *string   `json:"mode,omitempty" example:"fractional"`
	ScannerPattern  *string   `json:"scanner_pattern,omitempty" example:"square"`
	Power           *float64  `json:"power,omitempty" example:"20.0"`
	PulseEnergy     *float64  `json:"pulse_energy,omitempty" example:"100.0"`
	PulseDuration   *float64  `json:"pulse_duration,omitempty" example:"0.5"`
	Density         *float64  `json:"density,omitempty" example:"5.0"`
	Pattern         *string   `json:"pattern,omitempty" example:"grid"`
	Passes          *int      `json:"passes,omitempty" example:"2"`
	AnaesthesiaUsed *string   `json:"anaesthesia_used,omitempty" example:"topical"`
	Notes           *string   `json:"notes,omitempty"`
	Version         int       `json:"version" example:"1"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       int64     `json:"created_by" example:"1"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       int64     `json:"updated_by" example:"1"`
}

// RFModuleDetailResponse represents an RF module detail in API responses.
type RFModuleDetailResponse struct {
	ID            int64     `json:"id" example:"1"`
	ModuleID      int64     `json:"module_id" example:"1"`
	DeviceID      int64     `json:"device_id" example:"1"`
	HandpieceID   *int64    `json:"handpiece_id,omitempty" example:"1"`
	RFMode        *string   `json:"rf_mode,omitempty" example:"bipolar"`
	TipType       *string   `json:"tip_type,omitempty" example:"insulated"`
	Depth         *float64  `json:"depth,omitempty" example:"2.0"`
	EnergyLevel   *float64  `json:"energy_level,omitempty" example:"30.0"`
	Overlap       *float64  `json:"overlap,omitempty" example:"10.0"`
	PulsesPerZone *int      `json:"pulses_per_zone,omitempty" example:"3"`
	TotalPulses   *int      `json:"total_pulses,omitempty" example:"150"`
	Notes         *string   `json:"notes,omitempty"`
	Version       int       `json:"version" example:"1"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     int64     `json:"created_by" example:"1"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     int64     `json:"updated_by" example:"1"`
}
