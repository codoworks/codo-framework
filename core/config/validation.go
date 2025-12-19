package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return "configuration validation failed: " + strings.Join(msgs, "; ")
}

// Add adds a validation error
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, &ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// ToError returns nil if no errors, otherwise returns the ValidationErrors
func (e ValidationErrors) ToError() error {
	if !e.HasErrors() {
		return nil
	}
	return e
}
