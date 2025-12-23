package db

import (
	"database/sql/driver"
	"errors"
	"strings"

	fwkErrors "github.com/codoworks/codo-framework/core/errors"
)

func init() {
	// Get the global error mapper
	mapper := fwkErrors.GetMapper()

	// Register database-specific error types for automatic mapping
	// These will be mapped when database errors flow through the error handler

	// ErrNotFound - Record not found in database (404)
	mapper.RegisterSentinel(ErrNotFound, fwkErrors.MappingSpec{
		Code:       fwkErrors.CodeNotFound,
		HTTPStatus: 404,
		LogLevel:   fwkErrors.LogLevelWarn,
		Message:    "Resource not found",
	})

	// ErrDuplicateKey - Unique constraint violation (409)
	mapper.RegisterSentinel(ErrDuplicateKey, fwkErrors.MappingSpec{
		Code:       fwkErrors.CodeConflict,
		HTTPStatus: 409,
		LogLevel:   fwkErrors.LogLevelWarn,
		Message:    "Resource already exists",
	})

	// ErrInvalidModel - Invalid model for operation (400)
	mapper.RegisterSentinel(ErrInvalidModel, fwkErrors.MappingSpec{
		Code:       fwkErrors.CodeBadRequest,
		HTTPStatus: 400,
		LogLevel:   fwkErrors.LogLevelWarn,
		Message:    "Invalid data model",
	})

	// ErrNotPersisted - Operation requires persisted record (400)
	mapper.RegisterSentinel(ErrNotPersisted, fwkErrors.MappingSpec{
		Code:       fwkErrors.CodeBadRequest,
		HTTPStatus: 400,
		LogLevel:   fwkErrors.LogLevelWarn,
		Message:    "Record must be saved first",
	})

	// ErrNotInitialized - Database not initialized (503)
	mapper.RegisterSentinel(ErrNotInitialized, fwkErrors.MappingSpec{
		Code:       fwkErrors.CodeUnavailable,
		HTTPStatus: 503,
		LogLevel:   fwkErrors.LogLevelError,
		Message:    "Database service unavailable",
	})

	// Connection pool exhaustion and database connectivity errors (503)
	// These typically indicate temporary issues that may resolve with retry
	mapper.RegisterPredicate(func(err error) bool {
		if err == nil {
			return false
		}
		// Check for driver.ErrBadConn
		if errors.Is(err, driver.ErrBadConn) {
			return true
		}
		// Check for common connection error patterns
		errStr := strings.ToLower(err.Error())
		return strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "too many connections") ||
			strings.Contains(errStr, "too many open") ||
			strings.Contains(errStr, "dial tcp") ||
			strings.Contains(errStr, "connection reset") ||
			strings.Contains(errStr, "broken pipe")
	}, fwkErrors.MappingSpec{
		Code:       fwkErrors.CodeUnavailable,
		HTTPStatus: 503,
		LogLevel:   fwkErrors.LogLevelError,
		Message:    "Database temporarily unavailable",
	})
}
