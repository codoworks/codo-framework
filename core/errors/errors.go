package errors

import (
	"fmt"
	"net/http"
)

// Error represents a framework error with HTTP status mapping.
type Error struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	HTTPStatus int            `json:"-"`
	Cause      error          `json:"-"`
	Details    map[string]any `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for errors.Unwrap support.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is for error comparison by code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithCause adds a cause to the error and returns the error for chaining.
func (e *Error) WithCause(err error) *Error {
	e.Cause = err
	return e
}

// WithDetails adds details to the error and returns the error for chaining.
func (e *Error) WithDetails(details map[string]any) *Error {
	e.Details = details
	return e
}

// WithDetail adds a single detail to the error and returns the error for chaining.
func (e *Error) WithDetail(key string, value any) *Error {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// New creates a new error with custom code, message, and HTTP status.
func New(code, message string, httpStatus int) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Internal creates an internal server error.
func Internal(msg string) *Error {
	return &Error{
		Code:       CodeInternal,
		Message:    msg,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NotFound creates a not found error.
func NotFound(msg string) *Error {
	return &Error{
		Code:       CodeNotFound,
		Message:    msg,
		HTTPStatus: http.StatusNotFound,
	}
}

// BadRequest creates a bad request error.
func BadRequest(msg string) *Error {
	return &Error{
		Code:       CodeBadRequest,
		Message:    msg,
		HTTPStatus: http.StatusBadRequest,
	}
}

// Unauthorized creates an unauthorized error.
func Unauthorized(msg string) *Error {
	return &Error{
		Code:       CodeUnauthorized,
		Message:    msg,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// Forbidden creates a forbidden error.
func Forbidden(msg string) *Error {
	return &Error{
		Code:       CodeForbidden,
		Message:    msg,
		HTTPStatus: http.StatusForbidden,
	}
}

// Conflict creates a conflict error.
func Conflict(msg string) *Error {
	return &Error{
		Code:       CodeConflict,
		Message:    msg,
		HTTPStatus: http.StatusConflict,
	}
}

// Validation creates a validation error with a list of validation errors.
func Validation(msg string, errs []string) *Error {
	return &Error{
		Code:       CodeValidation,
		Message:    msg,
		HTTPStatus: http.StatusUnprocessableEntity,
		Details:    map[string]any{"errors": errs},
	}
}

// Timeout creates a timeout error.
func Timeout(msg string) *Error {
	return &Error{
		Code:       CodeTimeout,
		Message:    msg,
		HTTPStatus: http.StatusRequestTimeout,
	}
}

// Unavailable creates a service unavailable error.
func Unavailable(msg string) *Error {
	return &Error{
		Code:       CodeUnavailable,
		Message:    msg,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// Wrap wraps an existing error with a framework error.
func Wrap(err error, code, message string, httpStatus int) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      err,
	}
}

// WrapInternal wraps an error as an internal error.
func WrapInternal(err error, msg string) *Error {
	return Internal(msg).WithCause(err)
}

// WrapNotFound wraps an error as a not found error.
func WrapNotFound(err error, msg string) *Error {
	return NotFound(msg).WithCause(err)
}

// WrapBadRequest wraps an error as a bad request error.
func WrapBadRequest(err error, msg string) *Error {
	return BadRequest(msg).WithCause(err)
}

// IsError checks if an error is a framework Error with a specific code.
func IsError(err error, code string) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.Code == code
}

// IsInternal checks if an error is an internal error.
func IsInternal(err error) bool {
	return IsError(err, CodeInternal)
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return IsError(err, CodeNotFound)
}

// IsBadRequest checks if an error is a bad request error.
func IsBadRequest(err error) bool {
	return IsError(err, CodeBadRequest)
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return IsError(err, CodeUnauthorized)
}

// IsForbidden checks if an error is a forbidden error.
func IsForbidden(err error) bool {
	return IsError(err, CodeForbidden)
}

// IsConflict checks if an error is a conflict error.
func IsConflict(err error) bool {
	return IsError(err, CodeConflict)
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return IsError(err, CodeValidation)
}

// GetHTTPStatus returns the HTTP status code for an error.
// Returns 500 for non-framework errors.
func GetHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	e, ok := err.(*Error)
	if !ok {
		return http.StatusInternalServerError
	}
	return e.HTTPStatus
}

// GetCode returns the error code for an error.
// Returns CodeInternal for non-framework errors.
func GetCode(err error) string {
	if err == nil {
		return ""
	}
	e, ok := err.(*Error)
	if !ok {
		return CodeInternal
	}
	return e.Code
}
