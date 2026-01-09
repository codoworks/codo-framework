package errors

import (
	"context"
	"database/sql"
	"log"
	"reflect"
	"sync"
)

// LogLevel represents the severity level for logging
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// MappingSpec defines how an error should be mapped
type MappingSpec struct {
	Code       string
	HTTPStatus int
	LogLevel   LogLevel
	Message    string // Optional: override error message
}

// PredicateMapping allows predicate-based error matching
type PredicateMapping struct {
	Matches func(error) bool
	Spec    MappingSpec
}

// ErrorMapper automatically maps any error to framework Error type
// All methods are thread-safe via internal RWMutex
type ErrorMapper struct {
	mu                sync.RWMutex
	sentinelMappings  map[error]MappingSpec
	typeMappings      map[reflect.Type]MappingSpec
	predicateMappings []PredicateMapping
	converters        []func(error) *Error
}

var globalMapper = NewErrorMapper()

// NewErrorMapper creates a new error mapper with default mappings
func NewErrorMapper() *ErrorMapper {
	m := &ErrorMapper{
		sentinelMappings:  make(map[error]MappingSpec),
		typeMappings:      make(map[reflect.Type]MappingSpec),
		predicateMappings: make([]PredicateMapping, 0),
		converters:        make([]func(error) *Error, 0),
	}
	m.registerDefaults()
	return m
}

// registerDefaults registers framework default error mappings
func (m *ErrorMapper) registerDefaults() {
	// Standard library errors
	m.RegisterSentinel(sql.ErrNoRows, MappingSpec{
		Code:       CodeNotFound,
		HTTPStatus: 404,
		LogLevel:   LogLevelWarn,
		Message:    "Resource not found",
	})

	m.RegisterSentinel(context.DeadlineExceeded, MappingSpec{
		Code:       CodeTimeout,
		HTTPStatus: 408,
		LogLevel:   LogLevelWarn,
		Message:    "Request timeout",
	})

	m.RegisterSentinel(context.Canceled, MappingSpec{
		Code:       CodeInternal,
		HTTPStatus: 500,
		LogLevel:   LogLevelWarn,
		Message:    "Request canceled",
	})

	// OS error predicates - auto-map common file system errors
	// These work with wrapped errors via errors.Is()

	m.RegisterPredicate(isPermissionDenied, MappingSpec{
		Code:       CodeForbidden,
		HTTPStatus: 403,
		LogLevel:   LogLevelError,
		Message:    "Permission denied",
	})

	m.RegisterPredicate(isNotExist, MappingSpec{
		Code:       CodeNotFound,
		HTTPStatus: 404,
		LogLevel:   LogLevelWarn,
		Message:    "Resource not found",
	})

	m.RegisterPredicate(isAlreadyExists, MappingSpec{
		Code:       CodeConflict,
		HTTPStatus: 409,
		LogLevel:   LogLevelWarn,
		Message:    "Resource already exists",
	})

	m.RegisterPredicate(isTimeout, MappingSpec{
		Code:       CodeTimeout,
		HTTPStatus: 408,
		LogLevel:   LogLevelWarn,
		Message:    "Operation timed out",
	})

	m.RegisterPredicate(isClosed, MappingSpec{
		Code:       CodeUnavailable,
		HTTPStatus: 503,
		LogLevel:   LogLevelError,
		Message:    "Resource temporarily unavailable",
	})

	m.RegisterPredicate(isNoSpace, MappingSpec{
		Code:       CodeUnavailable,
		HTTPStatus: 507,
		LogLevel:   LogLevelError,
		Message:    "Insufficient storage",
	})
}

// RegisterSentinel registers a sentinel error mapping
func (m *ErrorMapper) RegisterSentinel(err error, spec MappingSpec) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.sentinelMappings[err]; exists {
		log.Printf("[errors] WARNING: Overwriting existing sentinel mapping for error: %v", err)
	}
	m.sentinelMappings[err] = spec
}

// RegisterType registers a type-based error mapping
func (m *ErrorMapper) RegisterType(errType error, spec MappingSpec) {
	t := reflect.TypeOf(errType)
	if t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t != nil {
		m.mu.Lock()
		defer m.mu.Unlock()
		if _, exists := m.typeMappings[t]; exists {
			log.Printf("[errors] WARNING: Overwriting existing type mapping for: %v", t)
		}
		m.typeMappings[t] = spec
	}
}

// RegisterPredicate registers a predicate-based error mapping
func (m *ErrorMapper) RegisterPredicate(predicate func(error) bool, spec MappingSpec) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.predicateMappings = append(m.predicateMappings, PredicateMapping{
		Matches: predicate,
		Spec:    spec,
	})
}

// RegisterConverter registers a custom error converter function
func (m *ErrorMapper) RegisterConverter(fn func(error) *Error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.converters = append(m.converters, fn)
}

// MapError maps any error to a framework Error
func (m *ErrorMapper) MapError(err error) *Error {
	if err == nil {
		return nil
	}

	// Already our error type? (no lock needed)
	if e, ok := err.(*Error); ok {
		return e
	}

	// Check if error contains a buried framework error (wrapped by fmt.Errorf, etc.)
	// This preserves the original caller info from where the error was created
	if embedded := findEmbeddedError(err); embedded != nil {
		return embedded
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check custom converters first
	for _, converter := range m.converters {
		if converted := converter(err); converted != nil {
			return converted
		}
	}

	// Check sentinel errors (exact match)
	if spec, ok := m.sentinelMappings[err]; ok {
		return m.createFromSpec(err, spec)
	}

	// Check if error wraps a known sentinel (using errors.Is)
	for sentinel, spec := range m.sentinelMappings {
		if isError(err, sentinel) {
			return m.createFromSpec(err, spec)
		}
	}

	// Check type-based mappings
	errType := reflect.TypeOf(err)
	if errType != nil {
		if errType.Kind() == reflect.Ptr {
			errType = errType.Elem()
		}
		if spec, ok := m.typeMappings[errType]; ok {
			return m.createFromSpec(err, spec)
		}
	}

	// Check predicate mappings
	for _, pm := range m.predicateMappings {
		if pm.Matches(err) {
			return m.createFromSpec(err, pm.Spec)
		}
	}

	// Default: internal error with preserved error chain
	return wrapUnknownError(err)
}

// createFromSpec creates an Error from a MappingSpec
func (m *ErrorMapper) createFromSpec(err error, spec MappingSpec) *Error {
	msg := spec.Message
	if msg == "" {
		msg = err.Error()
	}

	e := Wrap(err, spec.Code, msg, spec.HTTPStatus)
	e.LogLevel = spec.LogLevel
	return e
}

// findEmbeddedError walks the error chain to find any embedded framework *Error.
// This is used to find framework errors that have been wrapped by fmt.Errorf or similar.
// Returns nil if no framework error is found in the chain.
func findEmbeddedError(err error) *Error {
	current := err
	for current != nil {
		if e, ok := current.(*Error); ok {
			return e
		}
		// Try to unwrap
		type unwrapper interface {
			Unwrap() error
		}
		if u, ok := current.(unwrapper); ok {
			current = u.Unwrap()
		} else {
			break
		}
	}
	return nil
}

// isError checks if an error matches a target (similar to errors.Is but simpler)
func isError(err, target error) bool {
	if err == target {
		return true
	}

	// Check if error has Unwrap method
	type unwrapper interface {
		Unwrap() error
	}

	if u, ok := err.(unwrapper); ok {
		return isError(u.Unwrap(), target)
	}

	return false
}

// GetMapper returns the global error mapper for custom registration
func GetMapper() *ErrorMapper {
	return globalMapper
}

// MapError is a convenience function that uses the global mapper
func MapError(err error) *Error {
	return globalMapper.MapError(err)
}

// wrapUnknownError creates a framework error from an unknown error type,
// preserving the full error chain message for debugging while keeping
// a generic user-facing message.
func wrapUnknownError(err error) *Error {
	// Use the full error chain as the developer-facing message
	fullMessage := err.Error()

	// Create the framework error with the full chain as the message
	e := newWithSkip(CodeInternal, fullMessage, 500, 3)
	e.Cause = err

	if e.Details == nil {
		e.Details = make(map[string]any)
	}

	// Extract trace points if any TracedErrors exist in the chain
	// This provides better location info than the default stack trace
	tracePoints := ExtractTracePoints(err)
	if len(tracePoints) > 0 {
		e.Details["tracePoints"] = tracePoints

		// Use the deepest trace point (last one) as the caller location
		// since it's closest to the original error
		deepest := tracePoints[len(tracePoints)-1]
		e.Caller = &CallerInfo{
			File:     deepest.File,
			Line:     deepest.Line,
			Function: deepest.Function,
		}
	} else {
		// Fallback: extract error chain messages for debugging
		chain := extractErrorChain(err)
		if len(chain) > 1 {
			e.Details["errorChain"] = chain
		}
	}

	// Clean up empty details map
	if len(e.Details) == 0 {
		e.Details = nil
	}

	// Set a generic user-facing message (safe for end users)
	e.UserMessage = "An unexpected error occurred"
	e.MarkInternal()

	return e
}

// extractErrorChain walks the error chain and returns all unique messages.
// This provides visibility into the full error context for debugging.
func extractErrorChain(err error) []string {
	var chain []string
	seen := make(map[string]bool)
	current := err

	for current != nil {
		msg := current.Error()
		// Avoid duplicate messages (common when errors just wrap without adding context)
		if !seen[msg] {
			chain = append(chain, msg)
			seen[msg] = true
		}
		current = unwrapError(current)
	}

	return chain
}

// unwrapError extracts the wrapped error if any.
// Supports both single Unwrap() and multi-error Unwrap() []error (returns first).
func unwrapError(err error) error {
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
