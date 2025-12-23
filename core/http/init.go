package http

import (
	"github.com/codoworks/codo-framework/core/errors"
)

func init() {
	// Get the global error mapper
	mapper := errors.GetMapper()

	// Register HTTP-specific error types for automatic mapping
	// These will be mapped when errors flow through the error handler middleware

	// APIError - Custom HTTP errors with specific codes and status
	// Use a custom converter to preserve the code, message, and status
	mapper.RegisterConverter(func(err error) *errors.Error {
		if apiErr, ok := err.(*APIError); ok {
			return errors.New(apiErr.Code, apiErr.Message, apiErr.HTTPStatus)
		}
		return nil
	})

	// BindError - JSON/form binding failures
	mapper.RegisterType((*BindError)(nil), errors.MappingSpec{
		Code:       errors.CodeBadRequest,
		HTTPStatus: 400,
		LogLevel:   errors.LogLevelWarn,
		Message:    "Invalid request body",
	})

	// ParamError - Path/query parameter validation failures
	mapper.RegisterType((*ParamError)(nil), errors.MappingSpec{
		Code:       errors.CodeBadRequest,
		HTTPStatus: 400,
		LogLevel:   errors.LogLevelWarn,
		Message:    "", // Use the error's own message
	})

	// ValidationErrorList - Struct validation failures
	// Use a custom converter to preserve the structured errors list
	mapper.RegisterConverter(func(err error) *errors.Error {
		if valErr, ok := err.(*ValidationErrorList); ok {
			fwkErr := errors.New(errors.CodeValidation, "Validation failed", 422)
			// Store the structured validation errors in Details
			fwkErr.Details = map[string]any{
				"validationErrors": valErr.Errors,
			}
			return fwkErr
		}
		return nil
	})
}
