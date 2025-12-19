package http

import (
	"net/http"

	"github.com/codoworks/codo-framework/core/errors"
)

// Response is the standard API response structure
type Response struct {
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Errors     []string `json:"errors,omitempty"`
	Payload    any      `json:"payload,omitempty"`
	HTTPStatus int      `json:"-"`
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

// ErrorResponse creates an error response from any error
func ErrorResponse(err error) *Response {
	// Check if it's our custom API error type
	if e, ok := err.(*APIError); ok {
		return &Response{
			Code:       e.Code,
			Message:    e.Message,
			HTTPStatus: e.HTTPStatus,
		}
	}

	// Check for binding errors
	if _, ok := err.(*BindError); ok {
		return &Response{
			Code:       errors.CodeBadRequest,
			Message:    "Invalid request body",
			HTTPStatus: http.StatusBadRequest,
		}
	}

	// Check for param errors
	if e, ok := err.(*ParamError); ok {
		return &Response{
			Code:       errors.CodeBadRequest,
			Message:    e.Error(),
			HTTPStatus: http.StatusBadRequest,
		}
	}

	// Check for validation error list
	if e, ok := err.(*ValidationErrorList); ok {
		return &Response{
			Code:       errors.CodeValidation,
			Message:    "Validation failed",
			Errors:     e.Errors,
			HTTPStatus: http.StatusUnprocessableEntity,
		}
	}

	// Default to internal error
	return &Response{
		Code:       errors.CodeInternal,
		Message:    "An unexpected error occurred",
		HTTPStatus: http.StatusInternalServerError,
	}
}

// ValidationError creates a validation error response
func ValidationError(errs []string) *Response {
	return &Response{
		Code:       errors.CodeValidation,
		Message:    "Validation failed",
		Errors:     errs,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
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
