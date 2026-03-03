package handlers

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
