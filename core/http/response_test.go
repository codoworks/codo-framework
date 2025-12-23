package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	payload := map[string]string{"key": "value"}
	resp := Success(payload)

	assert.Equal(t, "OK", resp.Code)
	assert.Equal(t, "Success", resp.Message)
	assert.Equal(t, payload, resp.Payload)
	assert.Equal(t, http.StatusOK, resp.HTTPStatus)
}

func TestSuccessWithMessage(t *testing.T) {
	payload := map[string]string{"key": "value"}
	resp := SuccessWithMessage("Custom message", payload)

	assert.Equal(t, "OK", resp.Code)
	assert.Equal(t, "Custom message", resp.Message)
	assert.Equal(t, payload, resp.Payload)
	assert.Equal(t, http.StatusOK, resp.HTTPStatus)
}

func TestCreated(t *testing.T) {
	payload := map[string]string{"id": "123"}
	resp := Created(payload)

	assert.Equal(t, "CREATED", resp.Code)
	assert.Equal(t, "Resource created", resp.Message)
	assert.Equal(t, payload, resp.Payload)
	assert.Equal(t, http.StatusCreated, resp.HTTPStatus)
}

func TestCreatedWithMessage(t *testing.T) {
	payload := map[string]string{"id": "123"}
	resp := CreatedWithMessage("User created", payload)

	assert.Equal(t, "CREATED", resp.Code)
	assert.Equal(t, "User created", resp.Message)
	assert.Equal(t, payload, resp.Payload)
	assert.Equal(t, http.StatusCreated, resp.HTTPStatus)
}

func TestNoContentResponse(t *testing.T) {
	resp := NoContentResponse()

	assert.Equal(t, "NO_CONTENT", resp.Code)
	assert.Equal(t, "No content", resp.Message)
	assert.Nil(t, resp.Payload)
	assert.Equal(t, http.StatusNoContent, resp.HTTPStatus)
}

func TestErrorResponse_APIError(t *testing.T) {
	err := &APIError{
		Code:       "CUSTOM_ERROR",
		Message:    "Custom error message",
		HTTPStatus: http.StatusTeapot,
	}

	resp := ErrorResponse(err)

	assert.Equal(t, "CUSTOM_ERROR", resp.Code)
	assert.Equal(t, "Custom error message", resp.Message)
	assert.Equal(t, http.StatusTeapot, resp.HTTPStatus)
}

func TestErrorResponse_BindError(t *testing.T) {
	err := &BindError{Cause: errors.New("invalid json")}

	resp := ErrorResponse(err)

	assert.Equal(t, "BAD_REQUEST", resp.Code)
	assert.Equal(t, "Invalid request body", resp.Message)
	assert.Equal(t, http.StatusBadRequest, resp.HTTPStatus)
}

func TestErrorResponse_ParamError(t *testing.T) {
	err := &ParamError{Param: "id", Message: "required"}

	resp := ErrorResponse(err)

	assert.Equal(t, "BAD_REQUEST", resp.Code)
	assert.Contains(t, resp.Message, "parameter id")
	assert.Equal(t, http.StatusBadRequest, resp.HTTPStatus)
}

func TestErrorResponse_ValidationErrorList(t *testing.T) {
	err := &ValidationErrorList{Errors: []ValidationError{
		{Field: "email", Message: "is required"},
		{Field: "name", Message: "is too short"},
	}}

	resp := ErrorResponse(err)

	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Len(t, resp.Errors, 2)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.HTTPStatus)
}

func TestErrorResponse_UnknownError(t *testing.T) {
	err := errors.New("unknown error")

	resp := ErrorResponse(err)

	assert.Equal(t, "INTERNAL_ERROR", resp.Code)
	assert.Equal(t, "An unexpected error occurred", resp.Message)
	assert.Equal(t, http.StatusInternalServerError, resp.HTTPStatus)
}

func TestNotFound(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		resp := NotFound("User not found")
		assert.Equal(t, "NOT_FOUND", resp.Code)
		assert.Equal(t, "User not found", resp.Message)
		assert.Equal(t, http.StatusNotFound, resp.HTTPStatus)
	})

	t.Run("empty message uses default", func(t *testing.T) {
		resp := NotFound("")
		assert.Equal(t, "Resource not found", resp.Message)
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		resp := BadRequest("Invalid input")
		assert.Equal(t, "BAD_REQUEST", resp.Code)
		assert.Equal(t, "Invalid input", resp.Message)
		assert.Equal(t, http.StatusBadRequest, resp.HTTPStatus)
	})

	t.Run("empty message uses default", func(t *testing.T) {
		resp := BadRequest("")
		assert.Equal(t, "Bad request", resp.Message)
	})
}

func TestUnauthorized(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		resp := Unauthorized("Invalid token")
		assert.Equal(t, "UNAUTHORIZED", resp.Code)
		assert.Equal(t, "Invalid token", resp.Message)
		assert.Equal(t, http.StatusUnauthorized, resp.HTTPStatus)
	})

	t.Run("empty message uses default", func(t *testing.T) {
		resp := Unauthorized("")
		assert.Equal(t, "Unauthorized", resp.Message)
	})
}

func TestForbidden(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		resp := Forbidden("Access denied")
		assert.Equal(t, "FORBIDDEN", resp.Code)
		assert.Equal(t, "Access denied", resp.Message)
		assert.Equal(t, http.StatusForbidden, resp.HTTPStatus)
	})

	t.Run("empty message uses default", func(t *testing.T) {
		resp := Forbidden("")
		assert.Equal(t, "Forbidden", resp.Message)
	})
}

func TestConflict(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		resp := Conflict("Resource already exists")
		assert.Equal(t, "CONFLICT", resp.Code)
		assert.Equal(t, "Resource already exists", resp.Message)
		assert.Equal(t, http.StatusConflict, resp.HTTPStatus)
	})

	t.Run("empty message uses default", func(t *testing.T) {
		resp := Conflict("")
		assert.Equal(t, "Resource conflict", resp.Message)
	})
}

func TestServiceUnavailable(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		resp := ServiceUnavailable("Database down")
		assert.Equal(t, "SERVICE_UNAVAILABLE", resp.Code)
		assert.Equal(t, "Database down", resp.Message)
		assert.Equal(t, http.StatusServiceUnavailable, resp.HTTPStatus)
	})

	t.Run("empty message uses default", func(t *testing.T) {
		resp := ServiceUnavailable("")
		assert.Equal(t, "Service temporarily unavailable", resp.Message)
	})
}

func TestInternalError(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		resp := InternalError("Something went wrong")
		assert.Equal(t, "INTERNAL_ERROR", resp.Code)
		assert.Equal(t, "Something went wrong", resp.Message)
		assert.Equal(t, http.StatusInternalServerError, resp.HTTPStatus)
	})

	t.Run("empty message uses default", func(t *testing.T) {
		resp := InternalError("")
		assert.Equal(t, "An unexpected error occurred", resp.Message)
	})
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError("CUSTOM_CODE", "Custom message", http.StatusTeapot)

	assert.Equal(t, "CUSTOM_CODE", err.Code)
	assert.Equal(t, "Custom message", err.Message)
	assert.Equal(t, http.StatusTeapot, err.HTTPStatus)
	assert.Equal(t, "Custom message", err.Error())
}
