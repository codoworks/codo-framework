package http

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/errors"
)

// Test that HTTP error types are properly registered with the error mapper
// These registrations happen in init() when the package is imported

func TestBindErrorMapping(t *testing.T) {
	err := &BindError{Cause: fmt.Errorf("invalid JSON")}
	fwkErr := errors.MapError(err)

	assert.Equal(t, errors.CodeBadRequest, fwkErr.Code)
	assert.Equal(t, 400, fwkErr.HTTPStatus)
	assert.Equal(t, "Invalid request body", fwkErr.Message)
}

func TestParamErrorMapping(t *testing.T) {
	err := &ParamError{Param: "user_id", Message: "must be a valid UUID"}
	fwkErr := errors.MapError(err)

	assert.Equal(t, errors.CodeBadRequest, fwkErr.Code)
	assert.Equal(t, 400, fwkErr.HTTPStatus)
	// Message should use the error's own message since spec.Message is empty
	assert.Contains(t, fwkErr.Message, "user_id")
}

func TestValidationErrorListMapping(t *testing.T) {
	err := &ValidationErrorList{
		Errors: []ValidationError{
			{Field: "email", Message: "is required"},
			{Field: "name", Message: "is too short"},
		},
	}
	fwkErr := errors.MapError(err)

	assert.Equal(t, errors.CodeValidation, fwkErr.Code)
	assert.Equal(t, 422, fwkErr.HTTPStatus)
	assert.Equal(t, "Validation failed", fwkErr.Message)

	// Check that validation errors are in Details
	validationErrs, ok := fwkErr.Details["validationErrors"].([]ValidationError)
	assert.True(t, ok)
	assert.Len(t, validationErrs, 2)
	assert.Equal(t, "email", validationErrs[0].Field)
}

func TestAPIErrorMapping(t *testing.T) {
	err := &APIError{
		Code:       "CUSTOM_CODE",
		Message:    "Custom message",
		HTTPStatus: 418, // I'm a teapot
	}
	fwkErr := errors.MapError(err)

	assert.Equal(t, "CUSTOM_CODE", fwkErr.Code)
	assert.Equal(t, 418, fwkErr.HTTPStatus)
	assert.Equal(t, "Custom message", fwkErr.Message)
}

func TestAPIErrorMapping_PreservesValues(t *testing.T) {
	testCases := []struct {
		name       string
		code       string
		message    string
		httpStatus int
	}{
		{"teapot", "TEAPOT", "I'm a teapot", 418},
		{"forbidden", "FORBIDDEN", "No access", 403},
		{"conflict", "RESOURCE_CONFLICT", "Already exists", 409},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &APIError{
				Code:       tc.code,
				Message:    tc.message,
				HTTPStatus: tc.httpStatus,
			}
			fwkErr := errors.MapError(err)

			assert.Equal(t, tc.code, fwkErr.Code)
			assert.Equal(t, tc.message, fwkErr.Message)
			assert.Equal(t, tc.httpStatus, fwkErr.HTTPStatus)
		})
	}
}

func TestErrorResponse_UsesMapper(t *testing.T) {
	// Test that ErrorResponse properly uses the mapper

	t.Run("BindError", func(t *testing.T) {
		err := &BindError{Cause: fmt.Errorf("bad json")}
		resp := ErrorResponse(err)

		assert.Equal(t, errors.CodeBadRequest, resp.Code)
		assert.Equal(t, 400, resp.HTTPStatus)
	})

	t.Run("ValidationErrorList", func(t *testing.T) {
		err := &ValidationErrorList{
			Errors: []ValidationError{
				{Field: "email", Message: "required"},
			},
		}
		resp := ErrorResponse(err)

		assert.Equal(t, errors.CodeValidation, resp.Code)
		assert.Equal(t, 422, resp.HTTPStatus)
		assert.Len(t, resp.Errors, 1)
	})

	t.Run("APIError", func(t *testing.T) {
		err := &APIError{Code: "MY_CODE", Message: "My message", HTTPStatus: 422}
		resp := ErrorResponse(err)

		assert.Equal(t, "MY_CODE", resp.Code)
		assert.Equal(t, 422, resp.HTTPStatus)
	})

	t.Run("Unknown error", func(t *testing.T) {
		err := fmt.Errorf("unknown error")
		resp := ErrorResponse(err)

		assert.Equal(t, errors.CodeInternal, resp.Code)
		assert.Equal(t, 500, resp.HTTPStatus)
	})
}
