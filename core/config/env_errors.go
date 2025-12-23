package config

import (
	"fmt"
	"strings"

	"github.com/codoworks/codo-framework/core/errors"
)

// EnvValidationError represents a single environment variable validation error
type EnvValidationError struct {
	// VarName is the environment variable name (e.g., "STRIPE_API_KEY")
	VarName string

	// Message describes what went wrong
	Message string

	// Required indicates if this was a required variable
	Required bool

	// Sensitive indicates if the value should be masked
	Sensitive bool
}

func (e *EnvValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.VarName, e.Message)
}

// EnvValidationErrors is a collection of environment variable validation errors
type EnvValidationErrors []*EnvValidationError

func (e EnvValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return "environment variable validation failed: " + strings.Join(msgs, "; ")
}

// Add adds a validation error for a variable
func (e *EnvValidationErrors) Add(varName, message string) {
	*e = append(*e, &EnvValidationError{
		VarName: varName,
		Message: message,
	})
}

// AddWithDetails adds a validation error with full details
func (e *EnvValidationErrors) AddWithDetails(varName, message string, required, sensitive bool) {
	*e = append(*e, &EnvValidationError{
		VarName:   varName,
		Message:   message,
		Required:  required,
		Sensitive: sensitive,
	})
}

// HasErrors returns true if there are validation errors
func (e EnvValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// ToError returns nil if no errors, otherwise returns the EnvValidationErrors
func (e EnvValidationErrors) ToError() error {
	if !e.HasErrors() {
		return nil
	}
	return e
}

// ToFrameworkError converts the validation errors to a framework error
// This integrates with the error handling engine for proper rendering
func (e EnvValidationErrors) ToFrameworkError() *errors.Error {
	if len(e) == 0 {
		return nil
	}

	// Build error messages for the validation error
	msgs := make([]string, 0, len(e))
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}

	// Build details map with structured error info
	details := make(map[string]any)
	envErrors := make([]map[string]any, 0, len(e))
	for _, err := range e {
		errInfo := map[string]any{
			"variable": err.VarName,
			"message":  err.Message,
			"required": err.Required,
		}
		envErrors = append(envErrors, errInfo)
	}
	details["envErrors"] = envErrors

	// Create framework error with config phase
	fwkErr := errors.Validation("Environment variable validation failed", msgs)
	fwkErr.Phase = errors.PhaseConfig
	fwkErr.Details = details

	return fwkErr
}

// RequiredVars returns a list of variable names that failed due to being required but not set
func (e EnvValidationErrors) RequiredVars() []string {
	var vars []string
	for _, err := range e {
		if err.Required && strings.Contains(err.Message, "required") {
			vars = append(vars, err.VarName)
		}
	}
	return vars
}

// FormatForCLI returns a formatted string suitable for CLI output
func (e EnvValidationErrors) FormatForCLI() string {
	if len(e) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Environment Variable Errors:\n")
	for _, err := range e {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", err.VarName, err.Message))
	}
	return sb.String()
}
