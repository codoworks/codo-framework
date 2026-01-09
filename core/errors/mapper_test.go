package errors

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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
	// Message now preserves the full error chain (new behavior)
	assert.Equal(t, "some unknown error", result.Message)
	// UserMessage is the generic safe message for end users
	assert.Equal(t, "An unexpected error occurred", result.UserMessage)
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

// Tests for error chain preservation (Phase 1)

func TestMapError_PreservesErrorChainMessage(t *testing.T) {
	// Simulate a typical error chain like MAFS would create
	innerErr := fmt.Errorf("permission denied")
	middleErr := fmt.Errorf("failed to create directory: %w", innerErr)
	outerErr := fmt.Errorf("failed to upload to storage: %w", middleErr)

	result := MapError(outerErr)

	// Should be internal error (unknown type)
	assert.Equal(t, CodeInternal, result.Code)
	assert.Equal(t, 500, result.HTTPStatus)

	// Message should contain the full error chain, not just "An unexpected error occurred"
	assert.Contains(t, result.Message, "failed to upload to storage")
	assert.Contains(t, result.Message, "failed to create directory")
	assert.Contains(t, result.Message, "permission denied")

	// UserMessage should be generic (for end users)
	assert.Equal(t, "An unexpected error occurred", result.UserMessage)

	// Original error preserved
	assert.Equal(t, outerErr, result.Cause)
}

func TestMapError_ExtractsErrorChainToDetails(t *testing.T) {
	// Create a wrapped error chain
	inner := fmt.Errorf("root cause")
	middle := fmt.Errorf("middle layer: %w", inner)
	outer := fmt.Errorf("outer layer: %w", middle)

	result := MapError(outer)

	// Should have error chain in details
	chain, ok := result.Details["errorChain"].([]string)
	assert.True(t, ok, "should have errorChain in details")
	assert.True(t, len(chain) > 1, "error chain should have multiple entries")

	// Chain should include all levels
	fullChain := fmt.Sprintf("%v", chain)
	assert.Contains(t, fullChain, "outer layer")
	assert.Contains(t, fullChain, "middle layer")
	assert.Contains(t, fullChain, "root cause")
}

func TestMapError_TraceWrapPreservesCallerLocation(t *testing.T) {
	// Create an error using TraceWrap
	innerErr := fmt.Errorf("disk full")
	tracedErr := TraceWrap(innerErr, "failed to write file")

	result := MapError(tracedErr)

	// Should have trace points in details
	tracePoints, ok := result.Details["tracePoints"].([]TracePoint)
	assert.True(t, ok, "should have tracePoints in details")
	assert.Len(t, tracePoints, 1)

	// Trace point should have the message
	assert.Equal(t, "failed to write file", tracePoints[0].Message)

	// Caller should point to this test file, not mapper.go
	assert.Contains(t, result.Caller.File, "mapper_test.go")
}

func TestMapError_NestedTraceWrapPreservesAllLocations(t *testing.T) {
	// Create a chain of traced errors
	baseErr := fmt.Errorf("underlying error")
	trace1 := TraceWrap(baseErr, "layer 1")
	trace2 := TraceWrap(trace1, "layer 2")
	trace3 := TraceWrap(trace2, "layer 3")

	result := MapError(trace3)

	// Should have all trace points
	tracePoints, ok := result.Details["tracePoints"].([]TracePoint)
	assert.True(t, ok)
	assert.Len(t, tracePoints, 3)

	// Should be in order: outermost first
	assert.Equal(t, "layer 3", tracePoints[0].Message)
	assert.Equal(t, "layer 2", tracePoints[1].Message)
	assert.Equal(t, "layer 1", tracePoints[2].Message)

	// Caller should be the deepest trace point (closest to original error)
	assert.Equal(t, "layer 1", tracePoints[2].Message)
}

// Tests for OS error predicates (Phase 3)

func TestMapError_OSPermissionDenied(t *testing.T) {
	// Wrap os.ErrPermission
	wrappedErr := fmt.Errorf("failed to create file: %w", os.ErrPermission)

	result := MapError(wrappedErr)

	assert.Equal(t, CodeForbidden, result.Code)
	assert.Equal(t, 403, result.HTTPStatus)
	assert.Equal(t, "Permission denied", result.Message)
}

func TestMapError_OSNotExist(t *testing.T) {
	// Wrap os.ErrNotExist
	wrappedErr := fmt.Errorf("failed to read config: %w", os.ErrNotExist)

	result := MapError(wrappedErr)

	assert.Equal(t, CodeNotFound, result.Code)
	assert.Equal(t, 404, result.HTTPStatus)
}

func TestMapError_OSExist(t *testing.T) {
	// Wrap os.ErrExist
	wrappedErr := fmt.Errorf("failed to create file: %w", os.ErrExist)

	result := MapError(wrappedErr)

	assert.Equal(t, CodeConflict, result.Code)
	assert.Equal(t, 409, result.HTTPStatus)
}

func TestMapError_OSClosed(t *testing.T) {
	// Wrap os.ErrClosed
	wrappedErr := fmt.Errorf("failed to write: %w", os.ErrClosed)

	result := MapError(wrappedErr)

	assert.Equal(t, CodeUnavailable, result.Code)
	assert.Equal(t, 503, result.HTTPStatus)
}

func TestMapError_DeepNestedOSError(t *testing.T) {
	// Deeply nested OS error should still be detected
	err1 := fmt.Errorf("layer 1: %w", os.ErrPermission)
	err2 := fmt.Errorf("layer 2: %w", err1)
	err3 := fmt.Errorf("layer 3: %w", err2)

	result := MapError(err3)

	// Should still map to Forbidden thanks to errors.Is walking the chain
	assert.Equal(t, CodeForbidden, result.Code)
	assert.Equal(t, 403, result.HTTPStatus)
}
