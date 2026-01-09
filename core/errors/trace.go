package errors

import (
	"fmt"
	"runtime"
)

// TracePoint represents a point where an error was observed/wrapped.
// Unlike a full stack trace, trace points only capture intentional wrap points,
// making it easier to see the error's journey through the application layers.
type TracePoint struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Message  string `json:"message,omitempty"`
}

// TracedError wraps an error with caller information.
// Use TraceWrap() instead of fmt.Errorf() when you want to preserve the call location.
type TracedError struct {
	err   error
	trace TracePoint
}

// TraceWrap wraps an error with the current caller location and a message.
// Unlike fmt.Errorf, this preserves exactly where the wrap occurred, making it
// easier to trace errors back to their origin.
//
// Example:
//
//	if err := os.MkdirAll(path, 0755); err != nil {
//	    return errors.TraceWrap(err, "failed to create directory")
//	}
//
// When the error reaches the error handler, it will show:
//
//	"failed to create directory: permission denied"
//
// With trace points showing exactly where each wrap occurred.
func TraceWrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		// Fallback to regular wrapping if we can't capture caller info
		return fmt.Errorf("%s: %w", msg, err)
	}

	fn := runtime.FuncForPC(pc)
	fnName := ""
	if fn != nil {
		fnName = extractFunctionName(fn.Name())
	}

	return &TracedError{
		err: err,
		trace: TracePoint{
			File:     toModuleRelativePath(file),
			Line:     line,
			Function: fnName,
			Message:  msg,
		},
	}
}

// TraceWrapf wraps an error with a formatted message and caller location.
//
// Example:
//
//	return errors.TraceWrapf(err, "failed to upload file %s", filename)
func TraceWrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	// Skip 1 for TraceWrapf itself
	return traceWrapWithSkip(err, fmt.Sprintf(format, args...), 2)
}

// traceWrapWithSkip is the internal implementation with configurable skip depth.
func traceWrapWithSkip(err error, msg string, skip int) error {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return fmt.Errorf("%s: %w", msg, err)
	}

	fn := runtime.FuncForPC(pc)
	fnName := ""
	if fn != nil {
		fnName = extractFunctionName(fn.Name())
	}

	return &TracedError{
		err: err,
		trace: TracePoint{
			File:     toModuleRelativePath(file),
			Line:     line,
			Function: fnName,
			Message:  msg,
		},
	}
}

// Error implements the error interface.
func (e *TracedError) Error() string {
	if e.trace.Message != "" {
		return fmt.Sprintf("%s: %v", e.trace.Message, e.err)
	}
	return e.err.Error()
}

// Unwrap returns the wrapped error for errors.Unwrap support.
func (e *TracedError) Unwrap() error {
	return e.err
}

// TracePoint returns the trace information captured at wrap time.
func (e *TracedError) TracePoint() TracePoint {
	return e.trace
}

// ExtractTracePoints walks an error chain and extracts all TracedError trace points.
// The points are returned in order from outermost to innermost (most recent wrap first).
//
// Example:
//
//	err := errors.TraceWrap(osErr, "failed to create dir")
//	err = errors.TraceWrap(err, "failed to upload")
//	err = errors.TraceWrap(err, "failed to process request")
//
//	points := errors.ExtractTracePoints(err)
//	// points[0] = "failed to process request" (outermost)
//	// points[1] = "failed to upload"
//	// points[2] = "failed to create dir" (innermost)
func ExtractTracePoints(err error) []TracePoint {
	var points []TracePoint
	current := err

	for current != nil {
		if te, ok := current.(*TracedError); ok {
			points = append(points, te.trace)
		}
		current = unwrapSingle(current)
	}

	return points
}

// unwrapSingle extracts a single wrapped error.
// This is used internally and handles both Unwrap() error and Unwrap() []error.
func unwrapSingle(err error) error {
	switch e := err.(type) {
	case interface{ Unwrap() error }:
		return e.Unwrap()
	case interface{ Unwrap() []error }:
		errs := e.Unwrap()
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	default:
		return nil
	}
}

// HasTracePoints checks if an error chain contains any TracedErrors.
func HasTracePoints(err error) bool {
	current := err
	for current != nil {
		if _, ok := current.(*TracedError); ok {
			return true
		}
		current = unwrapSingle(current)
	}
	return false
}
