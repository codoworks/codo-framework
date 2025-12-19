package forms

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate validates a struct using go-playground/validator
func Validate(v any) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		return NewValidationErrors(validationErrs)
	}

	return err
}

// ValidationErrors holds a list of validation errors
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

// FieldError represents a single field validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
}

// NewValidationErrors creates ValidationErrors from validator errors
func NewValidationErrors(errs validator.ValidationErrors) *ValidationErrors {
	fieldErrors := make([]FieldError, 0, len(errs))

	for _, e := range errs {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   toSnakeCase(e.Field()),
			Message: formatValidationMessage(e),
			Tag:     e.Tag(),
			Value:   e.Value(),
		})
	}

	return &ValidationErrors{Errors: fieldErrors}
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}

	msgs := make([]string, len(v.Errors))
	for i, e := range v.Errors {
		msgs[i] = e.Message
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// GetErrors returns the error messages as strings
func (v *ValidationErrors) GetErrors() []string {
	msgs := make([]string, len(v.Errors))
	for i, e := range v.Errors {
		msgs[i] = e.Message
	}
	return msgs
}

// formatValidationMessage creates a human-readable error message
func formatValidationMessage(e validator.FieldError) string {
	field := toSnakeCase(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be at least %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be at most %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "eqfield":
		return fmt.Sprintf("%s must equal %s", field, toSnakeCase(e.Param()))
	default:
		return fmt.Sprintf("%s failed %s validation", field, e.Tag())
	}
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// SetValidator sets a custom validator instance
func SetValidator(v *validator.Validate) {
	validate = v
}

// GetValidator returns the current validator instance
func GetValidator() *validator.Validate {
	return validate
}

// RegisterValidation registers a custom validation function
func RegisterValidation(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}

// Must panics if validation fails
func Must(v any) {
	if err := Validate(v); err != nil {
		panic(err)
	}
}
