package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "basic message",
			err:      &Error{Code: CodeInternal, Message: "something went wrong"},
			expected: "INTERNAL_ERROR: something went wrong",
		},
		{
			name:     "empty message",
			err:      &Error{Code: CodeNotFound, Message: ""},
			expected: "NOT_FOUND: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestError_Error_WithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := &Error{
		Code:    CodeInternal,
		Message: "something went wrong",
		Cause:   cause,
	}

	got := err.Error()
	expected := "INTERNAL_ERROR: something went wrong: underlying error"
	assert.Equal(t, expected, got)
}

func TestError_WithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := Internal("test").WithCause(cause)

	assert.Equal(t, cause, err.Cause)
	assert.Contains(t, err.Error(), "underlying error")
}

func TestError_WithDetails(t *testing.T) {
	details := map[string]any{"field": "email", "reason": "invalid format"}
	err := BadRequest("validation failed").WithDetails(details)

	assert.Equal(t, details, err.Details)
}

func TestError_WithDetail(t *testing.T) {
	t.Run("adds single detail", func(t *testing.T) {
		err := BadRequest("validation failed").WithDetail("field", "email")

		assert.Equal(t, "email", err.Details["field"])
	})

	t.Run("initializes details map if nil", func(t *testing.T) {
		err := &Error{Code: CodeBadRequest, Message: "test"}
		err.WithDetail("key", "value")

		assert.NotNil(t, err.Details)
		assert.Equal(t, "value", err.Details["key"])
	})

	t.Run("appends to existing details", func(t *testing.T) {
		err := BadRequest("test").WithDetails(map[string]any{"existing": "value"})
		err.WithDetail("new", "value2")

		assert.Equal(t, "value", err.Details["existing"])
		assert.Equal(t, "value2", err.Details["new"])
	})
}

func TestError_Unwrap(t *testing.T) {
	t.Run("returns cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := Internal("test").WithCause(cause)

		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("returns nil when no cause", func(t *testing.T) {
		err := Internal("test")

		unwrapped := err.Unwrap()
		assert.Nil(t, unwrapped)
	})

	t.Run("works with errors.Unwrap", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := Internal("test").WithCause(cause)

		unwrapped := errors.Unwrap(err)
		assert.Equal(t, cause, unwrapped)
	})
}

func TestError_Is(t *testing.T) {
	t.Run("same code matches", func(t *testing.T) {
		err1 := Internal("error 1")
		err2 := Internal("error 2")

		assert.True(t, err1.Is(err2))
	})

	t.Run("works with errors.Is", func(t *testing.T) {
		err := Internal("test")
		target := &Error{Code: CodeInternal}

		assert.True(t, errors.Is(err, target))
	})
}

func TestError_Is_DifferentCode(t *testing.T) {
	err1 := Internal("error 1")
	err2 := NotFound("error 2")

	assert.False(t, err1.Is(err2))
}

func TestError_Is_NotError(t *testing.T) {
	err := Internal("test")
	otherErr := errors.New("standard error")

	assert.False(t, err.Is(otherErr))
}

func TestNew(t *testing.T) {
	err := New("CUSTOM_CODE", "custom message", http.StatusTeapot)

	assert.Equal(t, "CUSTOM_CODE", err.Code)
	assert.Equal(t, "custom message", err.Message)
	assert.Equal(t, http.StatusTeapot, err.HTTPStatus)
}

func TestInternal(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		wantCode string
		wantHTTP int
	}{
		{
			name:     "basic message",
			msg:      "something went wrong",
			wantCode: CodeInternal,
			wantHTTP: http.StatusInternalServerError,
		},
		{
			name:     "empty message",
			msg:      "",
			wantCode: CodeInternal,
			wantHTTP: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Internal(tt.msg)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.msg, err.Message)
			assert.Equal(t, tt.wantHTTP, err.HTTPStatus)
		})
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		wantCode string
		wantHTTP int
	}{
		{
			name:     "resource not found",
			msg:      "user not found",
			wantCode: CodeNotFound,
			wantHTTP: http.StatusNotFound,
		},
		{
			name:     "empty message",
			msg:      "",
			wantCode: CodeNotFound,
			wantHTTP: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotFound(tt.msg)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.msg, err.Message)
			assert.Equal(t, tt.wantHTTP, err.HTTPStatus)
		})
	}
}

func TestBadRequest(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		wantCode string
		wantHTTP int
	}{
		{
			name:     "invalid input",
			msg:      "invalid email format",
			wantCode: CodeBadRequest,
			wantHTTP: http.StatusBadRequest,
		},
		{
			name:     "empty message",
			msg:      "",
			wantCode: CodeBadRequest,
			wantHTTP: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BadRequest(tt.msg)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.msg, err.Message)
			assert.Equal(t, tt.wantHTTP, err.HTTPStatus)
		})
	}
}

func TestUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		wantCode string
		wantHTTP int
	}{
		{
			name:     "missing token",
			msg:      "authentication required",
			wantCode: CodeUnauthorized,
			wantHTTP: http.StatusUnauthorized,
		},
		{
			name:     "empty message",
			msg:      "",
			wantCode: CodeUnauthorized,
			wantHTTP: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unauthorized(tt.msg)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.msg, err.Message)
			assert.Equal(t, tt.wantHTTP, err.HTTPStatus)
		})
	}
}

func TestForbidden(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		wantCode string
		wantHTTP int
	}{
		{
			name:     "permission denied",
			msg:      "access denied",
			wantCode: CodeForbidden,
			wantHTTP: http.StatusForbidden,
		},
		{
			name:     "empty message",
			msg:      "",
			wantCode: CodeForbidden,
			wantHTTP: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Forbidden(tt.msg)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.msg, err.Message)
			assert.Equal(t, tt.wantHTTP, err.HTTPStatus)
		})
	}
}

func TestConflict(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		wantCode string
		wantHTTP int
	}{
		{
			name:     "duplicate resource",
			msg:      "email already exists",
			wantCode: CodeConflict,
			wantHTTP: http.StatusConflict,
		},
		{
			name:     "empty message",
			msg:      "",
			wantCode: CodeConflict,
			wantHTTP: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Conflict(tt.msg)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.msg, err.Message)
			assert.Equal(t, tt.wantHTTP, err.HTTPStatus)
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		errs       []string
		wantCode   string
		wantHTTP   int
		wantErrors []string
	}{
		{
			name:       "single validation error",
			msg:        "validation failed",
			errs:       []string{"email is required"},
			wantCode:   CodeValidation,
			wantHTTP:   http.StatusUnprocessableEntity,
			wantErrors: []string{"email is required"},
		},
		{
			name:       "multiple validation errors",
			msg:        "validation failed",
			errs:       []string{"email is required", "name too short"},
			wantCode:   CodeValidation,
			wantHTTP:   http.StatusUnprocessableEntity,
			wantErrors: []string{"email is required", "name too short"},
		},
		{
			name:       "empty errors",
			msg:        "validation failed",
			errs:       []string{},
			wantCode:   CodeValidation,
			wantHTTP:   http.StatusUnprocessableEntity,
			wantErrors: []string{},
		},
		{
			name:       "nil errors",
			msg:        "validation failed",
			errs:       nil,
			wantCode:   CodeValidation,
			wantHTTP:   http.StatusUnprocessableEntity,
			wantErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validation(tt.msg, tt.errs)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.msg, err.Message)
			assert.Equal(t, tt.wantHTTP, err.HTTPStatus)
			assert.Equal(t, tt.wantErrors, err.Details["errors"])
		})
	}
}

func TestTimeout(t *testing.T) {
	err := Timeout("request timed out")

	assert.Equal(t, CodeTimeout, err.Code)
	assert.Equal(t, "request timed out", err.Message)
	assert.Equal(t, http.StatusRequestTimeout, err.HTTPStatus)
}

func TestUnavailable(t *testing.T) {
	err := Unavailable("service temporarily unavailable")

	assert.Equal(t, CodeUnavailable, err.Code)
	assert.Equal(t, "service temporarily unavailable", err.Message)
	assert.Equal(t, http.StatusServiceUnavailable, err.HTTPStatus)
}

func TestWrap(t *testing.T) {
	cause := errors.New("database connection failed")
	err := Wrap(cause, CodeInternal, "failed to save user", http.StatusInternalServerError)

	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, "failed to save user", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
	assert.Equal(t, cause, err.Cause)
}

func TestWrapInternal(t *testing.T) {
	cause := errors.New("unexpected error")
	err := WrapInternal(cause, "internal error occurred")

	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, "internal error occurred", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
	assert.Equal(t, cause, err.Cause)
}

func TestWrapNotFound(t *testing.T) {
	cause := errors.New("record not in database")
	err := WrapNotFound(cause, "user not found")

	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, "user not found", err.Message)
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
	assert.Equal(t, cause, err.Cause)
}

func TestWrapBadRequest(t *testing.T) {
	cause := errors.New("json unmarshal failed")
	err := WrapBadRequest(cause, "invalid request body")

	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "invalid request body", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, cause, err.Cause)
}

func TestIsError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		expected bool
	}{
		{
			name:     "matching code",
			err:      Internal("test"),
			code:     CodeInternal,
			expected: true,
		},
		{
			name:     "non-matching code",
			err:      Internal("test"),
			code:     CodeNotFound,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			code:     CodeInternal,
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			code:     CodeInternal,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsError(tt.err, tt.code)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsInternal(t *testing.T) {
	assert.True(t, IsInternal(Internal("test")))
	assert.False(t, IsInternal(NotFound("test")))
	assert.False(t, IsInternal(nil))
	assert.False(t, IsInternal(errors.New("standard error")))
}

func TestIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(NotFound("test")))
	assert.False(t, IsNotFound(Internal("test")))
	assert.False(t, IsNotFound(nil))
}

func TestIsBadRequest(t *testing.T) {
	assert.True(t, IsBadRequest(BadRequest("test")))
	assert.False(t, IsBadRequest(Internal("test")))
	assert.False(t, IsBadRequest(nil))
}

func TestIsUnauthorized(t *testing.T) {
	assert.True(t, IsUnauthorized(Unauthorized("test")))
	assert.False(t, IsUnauthorized(Internal("test")))
	assert.False(t, IsUnauthorized(nil))
}

func TestIsForbidden(t *testing.T) {
	assert.True(t, IsForbidden(Forbidden("test")))
	assert.False(t, IsForbidden(Internal("test")))
	assert.False(t, IsForbidden(nil))
}

func TestIsConflict(t *testing.T) {
	assert.True(t, IsConflict(Conflict("test")))
	assert.False(t, IsConflict(Internal("test")))
	assert.False(t, IsConflict(nil))
}

func TestIsValidation(t *testing.T) {
	assert.True(t, IsValidation(Validation("test", nil)))
	assert.False(t, IsValidation(Internal("test")))
	assert.False(t, IsValidation(nil))
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "internal error",
			err:      Internal("test"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "not found error",
			err:      NotFound("test"),
			expected: http.StatusNotFound,
		},
		{
			name:     "bad request error",
			err:      BadRequest("test"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: http.StatusOK,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHTTPStatus(tt.err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "internal error",
			err:      Internal("test"),
			expected: CodeInternal,
		},
		{
			name:     "not found error",
			err:      NotFound("test"),
			expected: CodeNotFound,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCode(tt.err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
