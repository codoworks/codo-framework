package errors

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Custom error types for testing
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

type anotherCustomError struct {
	code string
}

func (e *anotherCustomError) Error() string {
	return e.code
}

func TestMapError_NilError(t *testing.T) {
	result := MapError(nil)
	assert.Nil(t, result)
}

func TestMapError_AlreadyFrameworkError(t *testing.T) {
	originalErr := BadRequest("test error")
	result := MapError(originalErr)

	// Should return the same error instance
	assert.Same(t, originalErr, result)
	assert.Equal(t, CodeBadRequest, result.Code)
	assert.Equal(t, 400, result.HTTPStatus)
}

func TestMapError_SentinelMatch(t *testing.T) {
	// Test default sentinel mappings
	t.Run("sql.ErrNoRows", func(t *testing.T) {
		result := MapError(sql.ErrNoRows)
		assert.Equal(t, CodeNotFound, result.Code)
		assert.Equal(t, 404, result.HTTPStatus)
		assert.Equal(t, "Resource not found", result.Message)
	})

	t.Run("context.DeadlineExceeded", func(t *testing.T) {
		result := MapError(context.DeadlineExceeded)
		assert.Equal(t, CodeTimeout, result.Code)
		assert.Equal(t, 408, result.HTTPStatus)
	})

	t.Run("context.Canceled", func(t *testing.T) {
		result := MapError(context.Canceled)
		assert.Equal(t, CodeInternal, result.Code)
		assert.Equal(t, 500, result.HTTPStatus)
	})
}

func TestMapError_WrappedSentinel(t *testing.T) {
	// Wrap sql.ErrNoRows in another error
	wrappedErr := fmt.Errorf("failed to find user: %w", sql.ErrNoRows)

	result := MapError(wrappedErr)
	assert.Equal(t, CodeNotFound, result.Code)
	assert.Equal(t, 404, result.HTTPStatus)
}

func TestMapError_CustomSentinelRegistration(t *testing.T) {
	mapper := NewErrorMapper()
	customErr := fmt.Errorf("my custom error")

	mapper.RegisterSentinel(customErr, MappingSpec{
		Code:       "CUSTOM_ERROR",
		HTTPStatus: 422,
		LogLevel:   LogLevelWarn,
		Message:    "Custom error occurred",
	})

	result := mapper.MapError(customErr)
	assert.Equal(t, "CUSTOM_ERROR", result.Code)
	assert.Equal(t, 422, result.HTTPStatus)
	assert.Equal(t, "Custom error occurred", result.Message)
}

func TestMapError_TypeMatch(t *testing.T) {
	mapper := NewErrorMapper()

	// Register type mapping
	mapper.RegisterType((*customError)(nil), MappingSpec{
		Code:       "CUSTOM_TYPE_ERROR",
		HTTPStatus: 400,
		Message:    "Custom type error",
	})

	// Test with instance of that type
	err := &customError{msg: "specific message"}
	result := mapper.MapError(err)

	assert.Equal(t, "CUSTOM_TYPE_ERROR", result.Code)
	assert.Equal(t, 400, result.HTTPStatus)
	assert.Equal(t, "Custom type error", result.Message)
}

func TestMapError_TypeMatchWithEmptyMessage(t *testing.T) {
	mapper := NewErrorMapper()

	// Register type mapping with empty message
	mapper.RegisterType((*customError)(nil), MappingSpec{
		Code:       "CUSTOM_TYPE_ERROR",
		HTTPStatus: 400,
		Message:    "", // Empty - should use error's message
	})

	err := &customError{msg: "my specific message"}
	result := mapper.MapError(err)

	assert.Equal(t, "CUSTOM_TYPE_ERROR", result.Code)
	assert.Equal(t, "my specific message", result.Message) // Uses error's message
}

func TestMapError_PredicateMatch(t *testing.T) {
	mapper := NewErrorMapper()

	// Register predicate that matches errors containing "permission"
	mapper.RegisterPredicate(
		func(err error) bool {
			return err != nil && len(err.Error()) > 0 && err.Error()[0:10] == "permission"
		},
		MappingSpec{
			Code:       CodeForbidden,
			HTTPStatus: 403,
			Message:    "Permission denied",
		},
	)

	err := fmt.Errorf("permission denied for resource X")
	result := mapper.MapError(err)

	assert.Equal(t, CodeForbidden, result.Code)
	assert.Equal(t, 403, result.HTTPStatus)
}

func TestMapError_ConverterPriority(t *testing.T) {
	mapper := NewErrorMapper()

	// Register a type mapping
	mapper.RegisterType((*customError)(nil), MappingSpec{
		Code:       "TYPE_MATCH",
		HTTPStatus: 400,
	})

	// Register a converter that takes priority
	mapper.RegisterConverter(func(err error) *Error {
		if _, ok := err.(*customError); ok {
			return New("CONVERTER_MATCH", "Converter handled this", 422)
		}
		return nil
	})

	err := &customError{msg: "test"}
	result := mapper.MapError(err)

	// Converter should win
	assert.Equal(t, "CONVERTER_MATCH", result.Code)
	assert.Equal(t, 422, result.HTTPStatus)
}

func TestMapError_ConverterReturnsNil(t *testing.T) {
	mapper := NewErrorMapper()

	// Register converter that returns nil for our error
	mapper.RegisterConverter(func(err error) *Error {
		return nil // Always returns nil
	})

	// Register type mapping as fallback
	mapper.RegisterType((*customError)(nil), MappingSpec{
		Code:       "TYPE_MATCH",
		HTTPStatus: 400,
	})

	err := &customError{msg: "test"}
	result := mapper.MapError(err)

	// Should fall through to type mapping
	assert.Equal(t, "TYPE_MATCH", result.Code)
}

func TestMapError_DefaultInternalError(t *testing.T) {
	// Unknown error type should map to internal error
	err := fmt.Errorf("some unknown error")
	result := MapError(err)

	assert.Equal(t, CodeInternal, result.Code)
	assert.Equal(t, 500, result.HTTPStatus)
	assert.Equal(t, "An unexpected error occurred", result.Message)
	assert.Equal(t, err, result.Cause) // Original error preserved
}

func TestMapError_MultipleTypeRegistrations(t *testing.T) {
	mapper := NewErrorMapper()

	mapper.RegisterType((*customError)(nil), MappingSpec{
		Code:       "CUSTOM_ERROR",
		HTTPStatus: 400,
	})

	mapper.RegisterType((*anotherCustomError)(nil), MappingSpec{
		Code:       "ANOTHER_ERROR",
		HTTPStatus: 422,
	})

	t.Run("first type", func(t *testing.T) {
		result := mapper.MapError(&customError{msg: "test"})
		assert.Equal(t, "CUSTOM_ERROR", result.Code)
	})

	t.Run("second type", func(t *testing.T) {
		result := mapper.MapError(&anotherCustomError{code: "test"})
		assert.Equal(t, "ANOTHER_ERROR", result.Code)
	})
}

func TestGetMapper_ReturnsSameInstance(t *testing.T) {
	mapper1 := GetMapper()
	mapper2 := GetMapper()
	assert.Same(t, mapper1, mapper2)
}

func TestMapError_CausePreserved(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	result := MapError(originalErr)

	assert.Equal(t, originalErr, result.Cause)
}

func TestMapError_NestedWrappedSentinel(t *testing.T) {
	// Double-wrap sql.ErrNoRows
	wrappedOnce := fmt.Errorf("layer 1: %w", sql.ErrNoRows)
	wrappedTwice := fmt.Errorf("layer 2: %w", wrappedOnce)

	result := MapError(wrappedTwice)
	assert.Equal(t, CodeNotFound, result.Code)
	assert.Equal(t, 404, result.HTTPStatus)
}

func TestRegisterType_NilPointer(t *testing.T) {
	mapper := NewErrorMapper()

	// This should not panic
	assert.NotPanics(t, func() {
		mapper.RegisterType((*customError)(nil), MappingSpec{
			Code:       "TEST",
			HTTPStatus: 400,
		})
	})

	// And should work correctly
	result := mapper.MapError(&customError{msg: "test"})
	assert.Equal(t, "TEST", result.Code)
}
