package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Common validator functions for environment variables
// These can be used in the Validator field of EnvVarDescriptor

// NotEmpty returns a validator that ensures the string value is not empty
func NotEmpty() EnvVarValidator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("value cannot be empty")
			}
		}
		return nil
	}
}

// MinLength returns a validator that ensures string has minimum length
func MinLength(min int) EnvVarValidator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if len(s) < min {
				return fmt.Errorf("value must be at least %d characters", min)
			}
		}
		return nil
	}
}

// MaxLength returns a validator that ensures string has maximum length
func MaxLength(max int) EnvVarValidator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if len(s) > max {
				return fmt.Errorf("value must be at most %d characters", max)
			}
		}
		return nil
	}
}

// Pattern returns a validator that ensures string matches a regex pattern
func Pattern(pattern string) EnvVarValidator {
	re := regexp.MustCompile(pattern)
	return func(value any) error {
		if s, ok := value.(string); ok {
			if !re.MatchString(s) {
				return fmt.Errorf("value must match pattern: %s", pattern)
			}
		}
		return nil
	}
}

// OneOf returns a validator that ensures value is one of the allowed values
func OneOf(allowed ...string) EnvVarValidator {
	allowedSet := make(map[string]bool, len(allowed))
	for _, v := range allowed {
		allowedSet[v] = true
	}
	return func(value any) error {
		if s, ok := value.(string); ok {
			if !allowedSet[s] {
				return fmt.Errorf("value must be one of: %s", strings.Join(allowed, ", "))
			}
		}
		return nil
	}
}

// IntRange returns a validator that ensures int value is within range
func IntRange(min, max int) EnvVarValidator {
	return func(value any) error {
		if i, ok := value.(int); ok {
			if i < min || i > max {
				return fmt.Errorf("value must be between %d and %d", min, max)
			}
		}
		return nil
	}
}

// IntMin returns a validator that ensures int value is at least min
func IntMin(min int) EnvVarValidator {
	return func(value any) error {
		if i, ok := value.(int); ok {
			if i < min {
				return fmt.Errorf("value must be at least %d", min)
			}
		}
		return nil
	}
}

// IntMax returns a validator that ensures int value is at most max
func IntMax(max int) EnvVarValidator {
	return func(value any) error {
		if i, ok := value.(int); ok {
			if i > max {
				return fmt.Errorf("value must be at most %d", max)
			}
		}
		return nil
	}
}

// FloatRange returns a validator that ensures float value is within range
func FloatRange(min, max float64) EnvVarValidator {
	return func(value any) error {
		if f, ok := value.(float64); ok {
			if f < min || f > max {
				return fmt.Errorf("value must be between %f and %f", min, max)
			}
		}
		return nil
	}
}

// DurationRange returns a validator that ensures duration is within range
func DurationRange(min, max time.Duration) EnvVarValidator {
	return func(value any) error {
		if d, ok := value.(time.Duration); ok {
			if d < min || d > max {
				return fmt.Errorf("value must be between %s and %s", min, max)
			}
		}
		return nil
	}
}

// DurationMin returns a validator that ensures duration is at least min
func DurationMin(min time.Duration) EnvVarValidator {
	return func(value any) error {
		if d, ok := value.(time.Duration); ok {
			if d < min {
				return fmt.Errorf("value must be at least %s", min)
			}
		}
		return nil
	}
}

// URLSchemes returns a validator that ensures URL has one of the allowed schemes
func URLSchemes(schemes ...string) EnvVarValidator {
	schemeSet := make(map[string]bool, len(schemes))
	for _, s := range schemes {
		schemeSet[strings.ToLower(s)] = true
	}
	return func(value any) error {
		if u, ok := value.(*url.URL); ok {
			if !schemeSet[strings.ToLower(u.Scheme)] {
				return fmt.Errorf("URL scheme must be one of: %s", strings.Join(schemes, ", "))
			}
		}
		return nil
	}
}

// URLHasPath returns a validator that ensures URL has a non-empty path
func URLHasPath() EnvVarValidator {
	return func(value any) error {
		if u, ok := value.(*url.URL); ok {
			if u.Path == "" || u.Path == "/" {
				return fmt.Errorf("URL must have a path")
			}
		}
		return nil
	}
}

// Chain combines multiple validators into one
// Validation stops at the first error
func Chain(validators ...EnvVarValidator) EnvVarValidator {
	return func(value any) error {
		for _, v := range validators {
			if v != nil {
				if err := v(value); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// StartsWith returns a validator that ensures string starts with prefix
func StartsWith(prefix string) EnvVarValidator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if !strings.HasPrefix(s, prefix) {
				return fmt.Errorf("value must start with: %s", prefix)
			}
		}
		return nil
	}
}

// EndsWith returns a validator that ensures string ends with suffix
func EndsWith(suffix string) EnvVarValidator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if !strings.HasSuffix(s, suffix) {
				return fmt.Errorf("value must end with: %s", suffix)
			}
		}
		return nil
	}
}

// Contains returns a validator that ensures string contains substring
func Contains(substr string) EnvVarValidator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if !strings.Contains(s, substr) {
				return fmt.Errorf("value must contain: %s", substr)
			}
		}
		return nil
	}
}

// PositiveInt returns a validator that ensures int is positive (> 0)
func PositiveInt() EnvVarValidator {
	return IntMin(1)
}

// NonNegativeInt returns a validator that ensures int is non-negative (>= 0)
func NonNegativeInt() EnvVarValidator {
	return IntMin(0)
}

// Port returns a validator that ensures int is a valid port number (1-65535)
func Port() EnvVarValidator {
	return IntRange(1, 65535)
}

// HTTPOrHTTPS returns a validator that ensures URL uses http or https scheme
func HTTPOrHTTPS() EnvVarValidator {
	return URLSchemes("http", "https")
}

// HTTPS returns a validator that ensures URL uses https scheme
func HTTPS() EnvVarValidator {
	return URLSchemes("https")
}
