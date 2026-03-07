package apierrors

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a structured API error with a machine-readable code.
type ErrorResponse struct {
	Code  string `json:"code" example:"AUTH_INVALID_CREDENTIALS"`
	Error string `json:"error" example:"invalid credentials"`
}

// Error codes for validation failures.
const (
	ValidationInvalidRequestBody = "VALIDATION_INVALID_REQUEST_BODY"
	ValidationRequiredFields     = "VALIDATION_REQUIRED_FIELDS"
)

// Error codes for authentication/authorization failures.
const (
	AuthInvalidCredentials  = "AUTH_INVALID_CREDENTIALS"
	AuthNotAuthenticated    = "AUTH_NOT_AUTHENTICATED"
	AuthMissingHeader       = "AUTH_MISSING_HEADER"
	AuthInvalidHeaderFormat = "AUTH_INVALID_HEADER_FORMAT"
	AuthInvalidToken        = "AUTH_INVALID_TOKEN"
	AuthInvalidRefreshToken = "AUTH_INVALID_REFRESH_TOKEN"
	AuthRefreshTokenRequired = "AUTH_REFRESH_TOKEN_REQUIRED"
)

// Error codes for authorization failures.
const (
	AuthInsufficientRole = "AUTH_INSUFFICIENT_ROLE"
)

// Error codes for user-related failures.
const (
	UserNotFound      = "USER_NOT_FOUND"
	UserAlreadyExists = "USER_ALREADY_EXISTS"
)

// Error codes for role-related failures.
const (
	RoleInvalidRole      = "ROLE_INVALID_ROLE"
	RoleAssignmentFailed = "ROLE_ASSIGNMENT_FAILED"
	RoleUserNotFound     = "ROLE_USER_NOT_FOUND"
)

// Error codes for internal server errors.
const (
	InternalPasswordProcessing      = "INTERNAL_PASSWORD_PROCESSING"
	InternalTokenGeneration         = "INTERNAL_TOKEN_GENERATION"
	InternalRefreshTokenGeneration  = "INTERNAL_REFRESH_TOKEN_GENERATION"
	InternalRefreshTokenStorage     = "INTERNAL_REFRESH_TOKEN_STORAGE"
	InternalUserLookup              = "INTERNAL_USER_LOOKUP"
)

// WriteError writes a structured JSON error response with the given status code,
// error code, and human-readable message.
func WriteError(w http.ResponseWriter, statusCode int, code string, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{ //nolint:errcheck // response write
		Code:  code,
		Error: message,
	})
}
