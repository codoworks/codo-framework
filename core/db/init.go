package db

import (
	"github.com/codoworks/codo-framework/core/errors"
)

func init() {
	// Get the global error mapper
	mapper := errors.GetMapper()

	// Register database-specific error types for automatic mapping
	// These will be mapped when database errors flow through the error handler

	// ErrNotFound - Record not found in database (404)
	mapper.RegisterSentinel(ErrNotFound, errors.MappingSpec{
		Code:       errors.CodeNotFound,
		HTTPStatus: 404,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Resource not found",
	})

	// ErrDuplicateKey - Unique constraint violation (409)
	mapper.RegisterSentinel(ErrDuplicateKey, errors.MappingSpec{
		Code:       errors.CodeConflict,
		HTTPStatus: 409,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Resource already exists",
	})

	// ErrInvalidModel - Invalid model for operation (400)
	mapper.RegisterSentinel(ErrInvalidModel, errors.MappingSpec{
		Code:       errors.CodeBadRequest,
		HTTPStatus: 400,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Invalid data model",
	})

	// ErrNotPersisted - Operation requires persisted record (400)
	mapper.RegisterSentinel(ErrNotPersisted, errors.MappingSpec{
		Code:       errors.CodeBadRequest,
		HTTPStatus: 400,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Record must be saved first",
	})

	// ErrNotInitialized - Database not initialized (503)
	mapper.RegisterSentinel(ErrNotInitialized, errors.MappingSpec{
		Code:       errors.CodeUnavailable,
		HTTPStatus: 503,
		LogLevel:   errors.LogLevelError,
		Message:    "Database service unavailable",
	})
}
