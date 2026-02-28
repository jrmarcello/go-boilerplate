package apperror

import (
	"fmt"
	"net/http"
)

// Common error codes
const (
	CodeInternalError   = "INTERNAL_ERROR"
	CodeInvalidRequest  = "INVALID_REQUEST"
	CodeValidationError = "VALIDATION_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
)

// AppError is the base application error.
// It implements the error interface and supports unwrapping.
type AppError struct {
	Code       string         // Unique code (e.g., "INVALID_EMAIL")
	Message    string         // User-friendly message
	HTTPStatus int            // Suggested HTTP status (400, 404, 500...)
	Details    map[string]any // Extra details (field, invalid value, etc.)
	Err        error          // Original error (for wrapping)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows using errors.Is() and errors.As() with the original error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// IsDomainError returns true if this is a client error (4xx).
func (e *AppError) IsDomainError() bool {
	return e.HTTPStatus >= 400 && e.HTTPStatus < 500
}

// WithDetails returns a copy with additional details.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	newErr := *e
	newErr.Details = details
	return &newErr
}

// WithError returns a copy with the original error wrapped.
func (e *AppError) WithError(err error) *AppError {
	newErr := *e
	newErr.Err = err
	return &newErr
}

// =============================================================================
// Constructors
// =============================================================================

// New creates a new AppError.
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// BadRequest creates a 400 Bad Request error.
func BadRequest(code, message string) *AppError {
	return New(code, message, http.StatusBadRequest)
}

// NotFound creates a 404 Not Found error.
func NotFound(code, message string) *AppError {
	return New(code, message, http.StatusNotFound)
}

// Conflict creates a 409 Conflict error.
func Conflict(code, message string) *AppError {
	return New(code, message, http.StatusConflict)
}

// Internal creates a 500 Internal Server Error.
func Internal(code, message string) *AppError {
	return New(code, message, http.StatusInternalServerError)
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(code, message string) *AppError {
	return New(code, message, http.StatusUnauthorized)
}

// Forbidden creates a 403 Forbidden error.
func Forbidden(code, message string) *AppError {
	return New(code, message, http.StatusForbidden)
}

// Wrap wraps an existing error into an AppError.
func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}
