package http

import (
	"net/http"

	"github.com/codoworks/codo-framework/core/errors"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`   // JSON field name (e.g., "email", "address.street")
	Message string `json:"message"` // Constraint message (e.g., "is required", "must be valid")
}

// Response is the standard API response structure
type Response struct {
	Code       string             `json:"code"`
	Message    string             `json:"message"`
	Errors     []ValidationError  `json:"errors,omitempty"`   // BREAKING CHANGE: was []string, now structured
	Warnings   []Warning          `json:"warnings,omitempty"` // Non-fatal issues (partial failures, deprecations)
	Payload    any                `json:"payload,omitempty"`
	HTTPStatus int                `json:"-"`
}

// APIError is an error that carries HTTP status and code information
type APIError struct {
	Code       string
	Message    string
	HTTPStatus int
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// Success creates a success response
func Success(payload any) *Response {
	return &Response{
		Code:       "OK",
		Message:    "Success",
		Payload:    payload,
		HTTPStatus: http.StatusOK,
	}
}

// SuccessWithMessage creates a success response with a custom message
func SuccessWithMessage(message string, payload any) *Response {
	return &Response{
		Code:       "OK",
		Message:    message,
		Payload:    payload,
		HTTPStatus: http.StatusOK,
	}
}

// Created creates a 201 response
func Created(payload any) *Response {
	return &Response{
		Code:       "CREATED",
		Message:    "Resource created",
		Payload:    payload,
		HTTPStatus: http.StatusCreated,
	}
}

// CreatedWithMessage creates a 201 response with a custom message
func CreatedWithMessage(message string, payload any) *Response {
	return &Response{
		Code:       "CREATED",
		Message:    message,
		Payload:    payload,
		HTTPStatus: http.StatusCreated,
	}
}

// NoContentResponse creates a 204 response
func NoContentResponse() *Response {
	return &Response{
		Code:       "NO_CONTENT",
		Message:    "No content",
		HTTPStatus: http.StatusNoContent,
	}
}

// ErrorResponse creates an error response from any error using the error mapper
func ErrorResponse(err error) *Response {
	// Map any error to framework Error using the global mapper
	// This handles all registered error types (BindError, ParamError, ValidationErrorList, etc.)
	fwkErr := errors.MapError(err)

	resp := &Response{
		Code:       fwkErr.Code,
		Message:    fwkErr.Message,
		HTTPStatus: fwkErr.HTTPStatus,
	}

	// Use user message if available (for custom user-facing messages)
	if fwkErr.UserMessage != "" {
		resp.Message = fwkErr.UserMessage
	}

	// Add validation errors if present
	// ValidationErrorList stores errors in Details["validationErrors"]
	if validationErrs, ok := fwkErr.Details["validationErrors"].([]ValidationError); ok {
		resp.Errors = validationErrs
	}

	return resp
}

// NotFound creates a not found response
func NotFound(message string) *Response {
	if message == "" {
		message = "Resource not found"
	}
	return &Response{
		Code:       errors.CodeNotFound,
		Message:    message,
		HTTPStatus: http.StatusNotFound,
	}
}

// BadRequest creates a bad request response
func BadRequest(message string) *Response {
	if message == "" {
		message = "Bad request"
	}
	return &Response{
		Code:       errors.CodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// Unauthorized creates an unauthorized response
func Unauthorized(message string) *Response {
	if message == "" {
		message = "Unauthorized"
	}
	return &Response{
		Code:       errors.CodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// Forbidden creates a forbidden response
func Forbidden(message string) *Response {
	if message == "" {
		message = "Forbidden"
	}
	return &Response{
		Code:       errors.CodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// Conflict creates a conflict response
func Conflict(message string) *Response {
	if message == "" {
		message = "Resource conflict"
	}
	return &Response{
		Code:       errors.CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// ServiceUnavailable creates a service unavailable response
func ServiceUnavailable(message string) *Response {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return &Response{
		Code:       errors.CodeUnavailable,
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// InternalError creates an internal error response
func InternalError(message string) *Response {
	if message == "" {
		message = "An unexpected error occurred"
	}
	return &Response{
		Code:       errors.CodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewAPIError creates a new APIError
func NewAPIError(code string, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}
