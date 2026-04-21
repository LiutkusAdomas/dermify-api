package domain

import "time"

// User represents a registered user in the system.
type User struct {
	ID                 int64     `json:"id"`
	Username           string    `json:"username"`
	Email              string    `json:"email"`
	PasswordHash       string    `json:"-"`
	Bio                *string   `json:"bio,omitempty"`
	Role               string    `json:"role"`
	Language           string    `json:"language"`
	Timezone           string    `json:"timezone"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
