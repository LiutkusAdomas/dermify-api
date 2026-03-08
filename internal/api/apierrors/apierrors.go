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

// Error codes for patient-related failures.
const (
	PatientNotFound        = "PATIENT_NOT_FOUND"
	PatientVersionConflict = "PATIENT_VERSION_CONFLICT"
	PatientInvalidData     = "PATIENT_INVALID_DATA"
	PatientCreationFailed  = "PATIENT_CREATION_FAILED"
	PatientUpdateFailed    = "PATIENT_UPDATE_FAILED"
	PatientLookupFailed    = "PATIENT_LOOKUP_FAILED"
)

// Error codes for registry-related failures.
const (
	RegistryDeviceNotFound  = "REGISTRY_DEVICE_NOT_FOUND"
	RegistryProductNotFound = "REGISTRY_PRODUCT_NOT_FOUND"
	RegistryLookupFailed    = "REGISTRY_LOOKUP_FAILED"
)

// Error codes for session-related failures.
const (
	SessionNotFound          = "SESSION_NOT_FOUND"
	SessionVersionConflict   = "SESSION_VERSION_CONFLICT"
	SessionInvalidData       = "SESSION_INVALID_DATA"
	SessionInvalidTransition = "SESSION_INVALID_TRANSITION"
	SessionNotEditable       = "SESSION_NOT_EDITABLE"
	SessionCreationFailed    = "SESSION_CREATION_FAILED"
	SessionUpdateFailed      = "SESSION_UPDATE_FAILED"
	SessionLookupFailed      = "SESSION_LOOKUP_FAILED"
)

// Error codes for consent-related failures.
const (
	ConsentNotFound       = "CONSENT_NOT_FOUND"
	ConsentRequired       = "CONSENT_REQUIRED"
	ConsentAlreadyExists  = "CONSENT_ALREADY_EXISTS"
	ConsentInvalidData    = "CONSENT_INVALID_DATA"
	ConsentCreationFailed = "CONSENT_CREATION_FAILED"
)

// Error codes for screening-related failures.
const (
	ScreeningNotFound       = "SCREENING_NOT_FOUND"
	ScreeningAlreadyExists  = "SCREENING_ALREADY_EXISTS"
	ScreeningInvalidData    = "SCREENING_INVALID_DATA"
	ScreeningCreationFailed = "SCREENING_CREATION_FAILED"
)

// Error codes for module-related failures.
const (
	ModuleNotFound       = "MODULE_NOT_FOUND"
	ModuleInvalidData    = "MODULE_INVALID_DATA"
	ModuleCreationFailed = "MODULE_CREATION_FAILED"
	ModuleRemovalFailed  = "MODULE_REMOVAL_FAILED"
)

// Error codes for energy module detail failures.
const (
	ModuleDetailNotFound        = "MODULE_DETAIL_NOT_FOUND"
	ModuleDetailVersionConflict = "MODULE_DETAIL_VERSION_CONFLICT"
	ModuleDeviceTypeMismatch    = "MODULE_DEVICE_TYPE_MISMATCH"
	ModuleHandpieceMismatch     = "MODULE_HANDPIECE_MISMATCH"
	ModuleDetailUpdateFailed    = "MODULE_DETAIL_UPDATE_FAILED"
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
