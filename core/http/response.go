package http

import (
	"net/http"
	"sync"

	"github.com/codoworks/codo-framework/core/errors"
	"github.com/codoworks/codo-framework/core/forms"
)

// HandlerConfig holds settings for HTTP error response behavior
// Set via SetHandlerConfig during initialization
type HandlerConfig struct {
	ExposeDetails     bool // Include error details in response
	ExposeStackTraces bool // Include stack traces in response (NEVER in production)
	StrictResponse    bool // Include all fields in responses (no omitempty behavior)
}

// defaultHandlerConfig returns safe defaults (production-ready)
func defaultHandlerConfig() HandlerConfig {
	return HandlerConfig{
		ExposeDetails:     false,
		ExposeStackTraces: false,
	}
}

var (
	handlerConfig   = defaultHandlerConfig()
	handlerConfigMu sync.RWMutex
)

// SetHandlerConfig sets the global handler configuration
// Should be called during bootstrap before handling requests
func SetHandlerConfig(cfg HandlerConfig) {
	handlerConfigMu.Lock()
	defer handlerConfigMu.Unlock()
	handlerConfig = cfg
}

// GetHandlerConfig returns the current handler configuration
func GetHandlerConfig() HandlerConfig {
	handlerConfigMu.RLock()
	defer handlerConfigMu.RUnlock()
	return handlerConfig
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`   // JSON field name (e.g., "email", "address.street")
	Message string `json:"message"` // Constraint message (e.g., "is required", "must be valid")
}

// Response is the standard API response structure
type Response struct {
	Code       string             `json:"code"`
	Message    string             `json:"message"`
	Errors     []ValidationError  `json:"errors,omitempty"`   // Structured validation errors
	Warnings   []Warning          `json:"warnings,omitempty"` // Non-fatal issues (partial failures, deprecations)
	Payload    any                `json:"payload,omitempty"`
	Page       *forms.PageMeta    `json:"page,omitempty"`     // Pagination metadata (set via pagination.SetMeta or SetCursorMeta)
	HTTPStatus int                `json:"-"`

	// Debug fields (only included when config.ExposeDetails/ExposeStackTraces is true)
	Details    map[string]any       `json:"details,omitempty"`    // Error details (dev mode only)
	StackTrace []errors.StackFrame  `json:"stackTrace,omitempty"` // Stack trace (dev mode only, NEVER in production)
}

// StrictResponse is a wrapper for Response that always includes all fields
// Used when response.strict config is true
type StrictResponse struct {
	Code       string              `json:"code"`
	Message    string              `json:"message"`
	Errors     []ValidationError   `json:"errors"`   // No omitempty - always present
	Warnings   []Warning           `json:"warnings"` // No omitempty - always present
	Payload    any                 `json:"payload"`  // No omitempty - null if empty
	Page       *forms.PageMeta     `json:"page"`     // No omitempty - null if empty
	Details    map[string]any      `json:"details,omitempty"`    // Still conditional on config
	StackTrace []errors.StackFrame `json:"stackTrace,omitempty"` // Still conditional on config
}

// ToStrict converts Response to StrictResponse (ensures empty arrays not nil)
func (r *Response) ToStrict() StrictResponse {
	errs := r.Errors
	if errs == nil {
		errs = []ValidationError{}
	}
	warnings := r.Warnings
	if warnings == nil {
		warnings = []Warning{}
	}
	return StrictResponse{
		Code:       r.Code,
		Message:    r.Message,
		Errors:     errs,
		Warnings:   warnings,
		Payload:    r.Payload,
		Page:       r.Page,
		Details:    r.Details,
		StackTrace: r.StackTrace,
	}
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
	cfg := GetHandlerConfig()

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

	// Optionally expose details (dev mode only)
	if cfg.ExposeDetails {
		details := make(map[string]any)

		// Include error details (filtering out validationErrors)
		for k, v := range fwkErr.Details {
			if k != "validationErrors" {
				details[k] = v
			}
		}

		// Include full cause message for debugging
		// This shows the complete error chain message when available
		if fwkErr.Cause != nil {
			details["causeMessage"] = fwkErr.Cause.Error()
		}

		// Include caller location if available
		if fwkErr.Caller != nil {
			details["location"] = map[string]any{
				"file":     fwkErr.Caller.File,
				"line":     fwkErr.Caller.Line,
				"function": fwkErr.Caller.Function,
			}
		}

		// Include stack trace in details for complete debugging info
		if len(fwkErr.StackTrace) > 0 {
			details["stackTrace"] = fwkErr.StackTrace
		}

		if len(details) > 0 {
			resp.Details = details
		}
	}

	// Optionally expose stack trace as top-level field (dev mode only, NEVER in production)
	// Only set top-level StackTrace when ExposeDetails is false (to avoid duplication)
	// When ExposeDetails is true, the stack trace is already included in details["stackTrace"]
	if cfg.ExposeStackTraces && !cfg.ExposeDetails && len(fwkErr.StackTrace) > 0 {
		resp.StackTrace = fwkErr.StackTrace
	}

	return resp
}

// NotFound creates a not found response
//
// Deprecated: Use ErrorResponse(errors.NotFound("message")) instead.
// The errors package provides richer error context including caller info,
// stack traces, and error chaining.
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
//
// Deprecated: Use ErrorResponse(errors.BadRequest("message")) instead.
// The errors package provides richer error context including caller info,
// stack traces, and error chaining.
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
//
// Deprecated: Use ErrorResponse(errors.Unauthorized("message")) instead.
// The errors package provides richer error context including caller info,
// stack traces, and error chaining.
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
//
// Deprecated: Use ErrorResponse(errors.Forbidden("message")) instead.
// The errors package provides richer error context including caller info,
// stack traces, and error chaining.
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
//
// Deprecated: Use ErrorResponse(errors.Conflict("message")) instead.
// The errors package provides richer error context including caller info,
// stack traces, and error chaining.
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
//
// Deprecated: Use ErrorResponse(errors.Unavailable("message")) instead.
// The errors package provides richer error context including caller info,
// stack traces, and error chaining.
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
//
// Deprecated: Use ErrorResponse(errors.Internal("message")) instead.
// The errors package provides richer error context including caller info,
// stack traces, and error chaining.
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
