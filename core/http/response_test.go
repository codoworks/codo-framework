package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	fwkerrors "github.com/codoworks/codo-framework/core/errors"
	"github.com/codoworks/codo-framework/core/forms"
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

func TestResponse_ToStrict(t *testing.T) {
	t.Run("converts nil slices to empty slices", func(t *testing.T) {
		resp := &Response{
			Code:    "OK",
			Message: "Success",
			// Errors and Warnings are nil
		}

		strict := resp.ToStrict()

		assert.NotNil(t, strict.Errors)
		assert.Len(t, strict.Errors, 0)
		assert.NotNil(t, strict.Warnings)
		assert.Len(t, strict.Warnings, 0)
	})

	t.Run("preserves existing values", func(t *testing.T) {
		resp := &Response{
			Code:    "OK",
			Message: "Success",
			Errors: []ValidationError{
				{Field: "email", Message: "required"},
			},
			Warnings: []Warning{
				{Code: "DEPRECATED", Message: "API deprecated"},
			},
			Payload: map[string]string{"key": "value"},
		}

		strict := resp.ToStrict()

		assert.Equal(t, "OK", strict.Code)
		assert.Equal(t, "Success", strict.Message)
		assert.Len(t, strict.Errors, 1)
		assert.Equal(t, "email", strict.Errors[0].Field)
		assert.Len(t, strict.Warnings, 1)
		assert.Equal(t, "DEPRECATED", strict.Warnings[0].Code)
		assert.Equal(t, map[string]string{"key": "value"}, strict.Payload)
	})

	t.Run("preserves Page when set", func(t *testing.T) {
		resp := &Response{
			Code:    "OK",
			Message: "Success",
			Page: &forms.PageMeta{
				Type:    "offset",
				Total:   100,
				Page:    2,
				PerPage: 25,
				Pages:   4,
				HasNext: true,
				HasPrev: true,
			},
		}

		strict := resp.ToStrict()

		assert.NotNil(t, strict.Page)
		assert.Equal(t, int64(100), strict.Page.Total)
		assert.Equal(t, 2, strict.Page.Page)
	})

	t.Run("Page is nil when not set", func(t *testing.T) {
		resp := &Response{
			Code:    "OK",
			Message: "Success",
		}

		strict := resp.ToStrict()

		assert.Nil(t, strict.Page)
	})
}

func TestResponse_Page_Field(t *testing.T) {
	t.Run("Page can be set directly", func(t *testing.T) {
		resp := Success(nil)
		resp.Page = &forms.PageMeta{
			Type:    "offset",
			Total:   50,
			Page:    1,
			PerPage: 20,
			Pages:   3,
			HasNext: true,
			HasPrev: false,
		}

		assert.NotNil(t, resp.Page)
		assert.Equal(t, int64(50), resp.Page.Total)
		assert.Equal(t, 1, resp.Page.Page)
		assert.Equal(t, 20, resp.Page.PerPage)
		assert.True(t, resp.Page.HasNext)
		assert.False(t, resp.Page.HasPrev)
	})

	t.Run("Page defaults to nil", func(t *testing.T) {
		resp := Success(nil)
		assert.Nil(t, resp.Page)
	})
}

func TestHandlerConfig_StrictResponse(t *testing.T) {
	t.Run("default is false", func(t *testing.T) {
		cfg := defaultHandlerConfig()
		assert.False(t, cfg.StrictResponse)
	})

	t.Run("can be set to true", func(t *testing.T) {
		SetHandlerConfig(HandlerConfig{
			StrictResponse: true,
		})

		cfg := GetHandlerConfig()
		assert.True(t, cfg.StrictResponse)

		// Reset to default
		SetHandlerConfig(defaultHandlerConfig())
	})
}

func TestErrorResponse_ExposeDetails_IncludesStackTrace(t *testing.T) {
	// Enable stack traces for 5xx errors (default behavior)
	fwkerrors.SetCaptureConfig(fwkerrors.CaptureConfig{
		StackTraceOn5xx: true,
		StackTraceDepth: 10,
	})

	t.Run("stack trace in details when ExposeDetails is true", func(t *testing.T) {
		// Enable ExposeDetails
		SetHandlerConfig(HandlerConfig{
			ExposeDetails:     true,
			ExposeStackTraces: false, // Only ExposeDetails, not ExposeStackTraces
		})
		defer SetHandlerConfig(defaultHandlerConfig())

		// Create a framework error with stack trace (5xx errors get stack traces)
		err := fwkerrors.Internal("test error")

		resp := ErrorResponse(err)

		// Should have details with stack trace
		assert.NotNil(t, resp.Details)
		stackTrace, ok := resp.Details["stackTrace"]
		assert.True(t, ok, "details should contain stackTrace")
		assert.NotNil(t, stackTrace, "stackTrace should not be nil")

		// The top-level StackTrace should NOT be set (only ExposeDetails is true)
		assert.Nil(t, resp.StackTrace)
	})

	t.Run("no details when ExposeDetails is false", func(t *testing.T) {
		// Disable ExposeDetails
		SetHandlerConfig(HandlerConfig{
			ExposeDetails:     false,
			ExposeStackTraces: false,
		})
		defer SetHandlerConfig(defaultHandlerConfig())

		err := fwkerrors.Internal("test error")

		resp := ErrorResponse(err)

		// Should have no details
		assert.Nil(t, resp.Details)
		assert.Nil(t, resp.StackTrace)
	})

	t.Run("no duplicate stackTrace when both enabled", func(t *testing.T) {
		// Enable both - stack trace should only appear in details
		SetHandlerConfig(HandlerConfig{
			ExposeDetails:     true,
			ExposeStackTraces: true,
		})
		defer SetHandlerConfig(defaultHandlerConfig())

		err := fwkerrors.Internal("test error")

		resp := ErrorResponse(err)

		// Should have stack trace in details
		assert.NotNil(t, resp.Details)
		_, hasStackInDetails := resp.Details["stackTrace"]
		assert.True(t, hasStackInDetails, "details should contain stackTrace")
		// Top-level StackTrace should NOT be set (to avoid duplication)
		assert.Nil(t, resp.StackTrace, "top-level StackTrace should NOT be set when ExposeDetails is true")
	})

	t.Run("top-level stackTrace when only ExposeStackTraces enabled", func(t *testing.T) {
		// Only ExposeStackTraces - no details, so stack trace goes to top-level field
		SetHandlerConfig(HandlerConfig{
			ExposeDetails:     false,
			ExposeStackTraces: true,
		})
		defer SetHandlerConfig(defaultHandlerConfig())

		err := fwkerrors.Internal("test error")

		resp := ErrorResponse(err)

		// Should have top-level stack trace
		assert.NotNil(t, resp.StackTrace, "top-level StackTrace should be set when ExposeDetails is false")
		// Should NOT have details
		assert.Nil(t, resp.Details, "details should be nil when ExposeDetails is false")
	})
}

func TestErrorResponse_ExposeDetails_IncludesAllDebugInfo(t *testing.T) {
	// Enable stack traces
	fwkerrors.SetCaptureConfig(fwkerrors.CaptureConfig{
		StackTraceOn5xx: true,
		StackTraceDepth: 10,
	})

	SetHandlerConfig(HandlerConfig{
		ExposeDetails: true,
	})
	defer SetHandlerConfig(defaultHandlerConfig())

	// Create an error with cause
	cause := errors.New("underlying cause")
	err := fwkerrors.WrapInternal(cause, "wrapped error")

	resp := ErrorResponse(err)

	assert.NotNil(t, resp.Details)

	// Should have causeMessage
	causeMsg, hasCause := resp.Details["causeMessage"]
	assert.True(t, hasCause, "should have causeMessage")
	assert.Contains(t, causeMsg, "underlying cause")

	// Should have location
	location, hasLocation := resp.Details["location"]
	assert.True(t, hasLocation, "should have location")
	assert.NotNil(t, location)

	// Should have stackTrace
	_, hasStack := resp.Details["stackTrace"]
	assert.True(t, hasStack, "should have stackTrace")
}
