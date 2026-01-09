package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceWrap_NilError(t *testing.T) {
	result := TraceWrap(nil, "message")
	assert.Nil(t, result)
}

func TestTraceWrap_CapturesCallerLocation(t *testing.T) {
	baseErr := errors.New("base error")
	traced := TraceWrap(baseErr, "wrapped here")

	te, ok := traced.(*TracedError)
	assert.True(t, ok, "should be TracedError type")

	// Check trace point
	tp := te.TracePoint()
	assert.Equal(t, "wrapped here", tp.Message)
	assert.Contains(t, tp.File, "trace_test.go")
	assert.Greater(t, tp.Line, 0)
	assert.NotEmpty(t, tp.Function)
}

func TestTraceWrap_ErrorMethod(t *testing.T) {
	baseErr := errors.New("original error")
	traced := TraceWrap(baseErr, "context added")

	// Error message should include both
	assert.Equal(t, "context added: original error", traced.Error())
}

func TestTraceWrap_Unwrap(t *testing.T) {
	baseErr := errors.New("base")
	traced := TraceWrap(baseErr, "wrapper")

	// Should unwrap to base error
	unwrapped := errors.Unwrap(traced)
	assert.Equal(t, baseErr, unwrapped)
}

func TestTraceWrapf_FormatsMessage(t *testing.T) {
	baseErr := errors.New("base error")
	traced := TraceWrapf(baseErr, "failed to process %s with id %d", "file.txt", 42)

	assert.Equal(t, "failed to process file.txt with id 42: base error", traced.Error())
}

func TestTraceWrapf_NilError(t *testing.T) {
	result := TraceWrapf(nil, "message %d", 1)
	assert.Nil(t, result)
}

func TestExtractTracePoints_NoTracedErrors(t *testing.T) {
	// Regular error chain with no TracedErrors
	err1 := errors.New("base")
	err2 := fmt.Errorf("wrapper: %w", err1)

	points := ExtractTracePoints(err2)
	assert.Len(t, points, 0)
}

func TestExtractTracePoints_SingleTracedError(t *testing.T) {
	baseErr := errors.New("base")
	traced := TraceWrap(baseErr, "single trace")

	points := ExtractTracePoints(traced)
	assert.Len(t, points, 1)
	assert.Equal(t, "single trace", points[0].Message)
}

func TestExtractTracePoints_MultipleTracedErrors(t *testing.T) {
	baseErr := errors.New("base")
	trace1 := TraceWrap(baseErr, "first")
	trace2 := TraceWrap(trace1, "second")
	trace3 := TraceWrap(trace2, "third")

	points := ExtractTracePoints(trace3)
	assert.Len(t, points, 3)

	// Should be in order: outermost first
	assert.Equal(t, "third", points[0].Message)
	assert.Equal(t, "second", points[1].Message)
	assert.Equal(t, "first", points[2].Message)
}

func TestExtractTracePoints_MixedChain(t *testing.T) {
	// Mix of traced and regular errors
	baseErr := errors.New("base")
	trace1 := TraceWrap(baseErr, "traced 1")
	regular := fmt.Errorf("regular wrapper: %w", trace1)
	trace2 := TraceWrap(regular, "traced 2")

	points := ExtractTracePoints(trace2)
	assert.Len(t, points, 2)
	assert.Equal(t, "traced 2", points[0].Message)
	assert.Equal(t, "traced 1", points[1].Message)
}

func TestHasTracePoints_True(t *testing.T) {
	baseErr := errors.New("base")
	traced := TraceWrap(baseErr, "traced")

	assert.True(t, HasTracePoints(traced))
}

func TestHasTracePoints_False(t *testing.T) {
	baseErr := errors.New("base")
	wrapped := fmt.Errorf("wrapped: %w", baseErr)

	assert.False(t, HasTracePoints(wrapped))
}

func TestHasTracePoints_DeepInChain(t *testing.T) {
	baseErr := errors.New("base")
	traced := TraceWrap(baseErr, "traced")
	// Wrap the traced error in regular errors
	wrapped1 := fmt.Errorf("level 1: %w", traced)
	wrapped2 := fmt.Errorf("level 2: %w", wrapped1)

	assert.True(t, HasTracePoints(wrapped2))
}

func TestTracedError_WorksWithErrorsIs(t *testing.T) {
	sentinel := errors.New("sentinel error")
	traced := TraceWrap(sentinel, "wrapped")

	// errors.Is should work through TracedError
	assert.True(t, errors.Is(traced, sentinel))
}

type customErrForTest struct {
	code int
}

func (e *customErrForTest) Error() string {
	return fmt.Sprintf("custom error with code %d", e.code)
}

func TestTracedError_WorksWithErrorsAs(t *testing.T) {
	customError := &customErrForTest{code: 42}

	traced := TraceWrap(customError, "wrapped custom")

	var target *customErrForTest
	assert.True(t, errors.As(traced, &target))
	assert.Equal(t, 42, target.code)
}

func TestTraceWrap_PreservesLocationAcrossFunctions(t *testing.T) {
	// Helper function that wraps an error
	wrapInHelper := func(err error) error {
		return TraceWrap(err, "from helper")
	}

	baseErr := errors.New("base")
	traced := wrapInHelper(baseErr)

	te, ok := traced.(*TracedError)
	assert.True(t, ok)

	// Location should be in this test file (from the helper)
	assert.Contains(t, te.trace.File, "trace_test.go")
}

func TestTracePoint_EmptyMessage(t *testing.T) {
	// TracedError with empty message
	te := &TracedError{
		err: errors.New("base"),
		trace: TracePoint{
			File:    "test.go",
			Line:    10,
			Message: "",
		},
	}

	// Error() should just return the wrapped error's message
	assert.Equal(t, "base", te.Error())
}
