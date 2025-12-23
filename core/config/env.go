package config

import (
	"net/url"
	"time"
)

// EnvVarType defines the data type of an environment variable
type EnvVarType string

const (
	// EnvTypeString is a string environment variable
	EnvTypeString EnvVarType = "string"
	// EnvTypeInt is an integer environment variable
	EnvTypeInt EnvVarType = "int"
	// EnvTypeBool is a boolean environment variable
	EnvTypeBool EnvVarType = "bool"
	// EnvTypeDuration is a time.Duration environment variable (e.g., "30s", "5m")
	EnvTypeDuration EnvVarType = "duration"
	// EnvTypeFloat is a float64 environment variable
	EnvTypeFloat EnvVarType = "float"
	// EnvTypeURL is a URL environment variable (validated format)
	EnvTypeURL EnvVarType = "url"
)

// EnvVarValidator is a custom validation function for environment variables
type EnvVarValidator func(value any) error

// EnvVarDescriptor describes a single environment variable requirement
type EnvVarDescriptor struct {
	// Name is the full environment variable name (e.g., "STRIPE_API_KEY", "KRATOS_PUBLIC_URL")
	// Consumers control the naming convention - no automatic prefix is applied
	Name string

	// Type is the data type for parsing and validation
	Type EnvVarType

	// Required indicates whether this variable must be set
	// If true and the variable is not set, validation fails
	Required bool

	// Default is the default value if not set (only used for optional variables)
	// The type should match the EnvVarType (e.g., string for EnvTypeString)
	Default any

	// Description provides documentation for error messages and introspection
	Description string

	// Sensitive marks the value for masking in logs and error output
	Sensitive bool

	// Validator is an optional custom validation function
	// Called after type conversion with the converted value
	Validator EnvVarValidator

	// Group associates the variable with a client or component (e.g., "stripe", "kratos")
	// Used for retrieving all variables for a specific client via GetGroup()
	Group string
}

// EnvVarValue holds a resolved environment variable value
type EnvVarValue struct {
	// Descriptor is the original descriptor for this variable
	Descriptor *EnvVarDescriptor

	// RawValue is the raw string value from the environment
	RawValue string

	// Value is the type-converted value
	// The concrete type depends on Descriptor.Type:
	//   - EnvTypeString: string
	//   - EnvTypeInt: int
	//   - EnvTypeBool: bool
	//   - EnvTypeDuration: time.Duration
	//   - EnvTypeFloat: float64
	//   - EnvTypeURL: *url.URL
	Value any

	// IsSet is true if the variable was explicitly set in the environment
	// false if using the default value
	IsSet bool
}

// String returns the string value, panics if type mismatch
func (v *EnvVarValue) String() string {
	if s, ok := v.Value.(string); ok {
		return s
	}
	panic("EnvVarValue.String() called on non-string value")
}

// Int returns the int value, panics if type mismatch
func (v *EnvVarValue) Int() int {
	if i, ok := v.Value.(int); ok {
		return i
	}
	panic("EnvVarValue.Int() called on non-int value")
}

// Bool returns the bool value, panics if type mismatch
func (v *EnvVarValue) Bool() bool {
	if b, ok := v.Value.(bool); ok {
		return b
	}
	panic("EnvVarValue.Bool() called on non-bool value")
}

// Duration returns the time.Duration value, panics if type mismatch
func (v *EnvVarValue) Duration() time.Duration {
	if d, ok := v.Value.(time.Duration); ok {
		return d
	}
	panic("EnvVarValue.Duration() called on non-duration value")
}

// Float returns the float64 value, panics if type mismatch
func (v *EnvVarValue) Float() float64 {
	if f, ok := v.Value.(float64); ok {
		return f
	}
	panic("EnvVarValue.Float() called on non-float value")
}

// URL returns the *url.URL value, panics if type mismatch
func (v *EnvVarValue) URL() *url.URL {
	if u, ok := v.Value.(*url.URL); ok {
		return u
	}
	panic("EnvVarValue.URL() called on non-URL value")
}

// MaskedValue returns the value for display, masking sensitive values
func (v *EnvVarValue) MaskedValue() string {
	if v.Descriptor.Sensitive {
		if len(v.RawValue) == 0 {
			return ""
		}
		return "***MASKED***"
	}
	return v.RawValue
}
