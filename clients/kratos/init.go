package kratos

import (
	"github.com/codoworks/codo-framework/core/errors"
)

func init() {
	// Get the global error mapper
	mapper := errors.GetMapper()

	// Register Kratos-specific error types for automatic mapping
	// These will be mapped when authentication errors flow through the error handler

	// ErrNoSession - No session cookie present (401)
	mapper.RegisterSentinel(ErrNoSession, errors.MappingSpec{
		Code:       errors.CodeUnauthorized,
		HTTPStatus: 401,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Authentication required",
	})

	// ErrInvalidSession - Session validation failed (401)
	mapper.RegisterSentinel(ErrInvalidSession, errors.MappingSpec{
		Code:       errors.CodeUnauthorized,
		HTTPStatus: 401,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Invalid session",
	})

	// ErrSessionExpired - Session has expired (401)
	mapper.RegisterSentinel(ErrSessionExpired, errors.MappingSpec{
		Code:       errors.CodeUnauthorized,
		HTTPStatus: 401,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Session expired",
	})
}
