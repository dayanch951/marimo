package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrBadRequest          ErrorCode = "BAD_REQUEST"
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrNotFound            ErrorCode = "NOT_FOUND"
	ErrConflict            ErrorCode = "CONFLICT"
	ErrValidation          ErrorCode = "VALIDATION_ERROR"
	ErrRateLimitExceeded   ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrTooManyRequests     ErrorCode = "TOO_MANY_REQUESTS"

	// Server errors (5xx)
	ErrInternal            ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrTimeout             ErrorCode = "TIMEOUT"
	ErrDatabaseError       ErrorCode = "DATABASE_ERROR"
	ErrExternalService     ErrorCode = "EXTERNAL_SERVICE_ERROR"

	// Business logic errors
	ErrInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	ErrTokenExpired        ErrorCode = "TOKEN_EXPIRED"
	ErrTokenInvalid        ErrorCode = "TOKEN_INVALID"
	ErrDuplicateResource   ErrorCode = "DUPLICATE_RESOURCE"
	ErrTenantNotFound      ErrorCode = "TENANT_NOT_FOUND"
	ErrSubscriptionExpired ErrorCode = "SUBSCRIPTION_EXPIRED"
	ErrFeatureNotAvailable ErrorCode = "FEATURE_NOT_AVAILABLE"
	ErrInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Err        error                  `json:"-"` // Original error for logging
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getStatusCode(code),
	}
}

// Wrap wraps an existing error into AppError
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getStatusCode(code),
		Err:        err,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithDetail adds a single detail field
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// getStatusCode maps error codes to HTTP status codes
func getStatusCode(code ErrorCode) int {
	switch code {
	case ErrBadRequest, ErrValidation:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrInvalidCredentials, ErrTokenExpired, ErrTokenInvalid:
		return http.StatusUnauthorized
	case ErrForbidden, ErrInsufficientPermissions, ErrFeatureNotAvailable:
		return http.StatusForbidden
	case ErrNotFound, ErrTenantNotFound:
		return http.StatusNotFound
	case ErrConflict, ErrDuplicateResource:
		return http.StatusConflict
	case ErrRateLimitExceeded, ErrTooManyRequests:
		return http.StatusTooManyRequests
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors
func BadRequest(message string) *AppError {
	return New(ErrBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return New(ErrUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(ErrForbidden, message)
}

func NotFound(message string) *AppError {
	return New(ErrNotFound, message)
}

func Conflict(message string) *AppError {
	return New(ErrConflict, message)
}

func Internal(message string) *AppError {
	return New(ErrInternal, message)
}

func ValidationError(message string, details map[string]interface{}) *AppError {
	return New(ErrValidation, message).WithDetails(details)
}

func DatabaseError(err error) *AppError {
	return Wrap(err, ErrDatabaseError, "Database operation failed")
}

func ExternalServiceError(service string, err error) *AppError {
	return Wrap(err, ErrExternalService, fmt.Sprintf("External service %s failed", service))
}

// IsAppError checks if error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from error
func GetAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// ValidationErrors represents field validation errors
type ValidationErrors map[string]string

// NewValidationError creates a validation error with field details
func NewValidationError(fields ValidationErrors) *AppError {
	details := make(map[string]interface{})
	for k, v := range fields {
		details[k] = v
	}
	return ValidationError("Validation failed", details)
}

// ErrorResponse represents the JSON error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    ErrorCode              `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
	TraceID string                 `json:"trace_id,omitempty"`
}

// ToResponse converts AppError to ErrorResponse
func (e *AppError) ToResponse(traceID string) ErrorResponse {
	return ErrorResponse{
		Error:   e.Message,
		Code:    e.Code,
		Details: e.Details,
		TraceID: traceID,
	}
}

// Common validation errors
var (
	ErrEmailRequired    = "Email is required"
	ErrEmailInvalid     = "Email format is invalid"
	ErrPasswordRequired = "Password is required"
	ErrPasswordTooShort = "Password must be at least 8 characters"
	ErrNameRequired     = "Name is required"
	ErrIDInvalid        = "Invalid ID format"
	ErrFieldRequired    = "%s is required"
	ErrFieldInvalid     = "%s is invalid"
)
