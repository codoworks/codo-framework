package http

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// Register a function to get JSON tag names for validation error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		if name == "" {
			return fld.Name
		}
		return name
	})
}

// Validate validates a struct using go-playground/validator
func Validate(form any) error {
	err := validate.Struct(form)
	if err == nil {
		return nil
	}

	// Convert to our validation error format
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errs := make([]ValidationError, 0, len(validationErrs))
		for _, e := range validationErrs {
			errs = append(errs, formatValidationError(e))
		}
		return &ValidationErrorList{Errors: errs}
	}

	return err
}

// ValidationErrorList holds a list of validation errors
type ValidationErrorList struct {
	Errors []ValidationError
}

func (e *ValidationErrorList) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return e.Errors[0].Field + " " + e.Errors[0].Message
}

func formatValidationError(e validator.FieldError) ValidationError {
	return ValidationError{
		Field:   e.Field(), // Already uses JSON tag name from RegisterTagNameFunc
		Message: formatConstraintMessage(e),
	}
}

func formatConstraintMessage(e validator.FieldError) string {
	tag := e.Tag()

	// Return just the constraint message, not including the field name
	// The field name is already in ValidationError.Field
	switch tag {
	case "required":
		return "is required"
	case "min":
		return "must be at least " + e.Param() + " characters"
	case "max":
		return "must be at most " + e.Param() + " characters"
	case "email":
		return "must be a valid email"
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "gte":
		return "must be greater than or equal to " + e.Param()
	case "lte":
		return "must be less than or equal to " + e.Param()
	case "gt":
		return "must be greater than " + e.Param()
	case "lt":
		return "must be less than " + e.Param()
	case "oneof":
		return "must be one of: " + e.Param()
	case "numeric":
		return "must be numeric"
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	default:
		return "failed " + tag + " validation"
	}
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
