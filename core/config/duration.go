package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a time.Duration that supports YAML/text unmarshaling from
// duration strings (e.g., "5m", "1h30s") and integers (interpreted as seconds).
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler.
// Accepts both string values ("5m", "30s") and integers (seconds for backwards compatibility).
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		// Successfully decoded as string, parse as duration
		parsed, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*d = Duration(parsed)
		return nil
	}

	// Not a string, try as integer (backwards compat: seconds)
	var i int64
	if err := value.Decode(&i); err != nil {
		return err
	}
	*d = Duration(i * int64(time.Second))
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler for environment variable support.
func (d *Duration) UnmarshalText(text []byte) error {
	s := string(text)

	// First try parsing as duration string
	parsed, err := time.ParseDuration(s)
	if err == nil {
		*d = Duration(parsed)
		return nil
	}

	// If that fails, try parsing as integer (seconds)
	var i int64
	_, err = fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return fmt.Errorf("invalid duration: %q (expected format like '30s', '5m', '1h' or integer seconds)", s)
	}
	*d = Duration(i * int64(time.Second))
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// MarshalText implements encoding.TextMarshaler.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// String returns the string representation of the duration.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// Seconds returns the duration as a floating point number of seconds.
func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}

// Milliseconds returns the duration as milliseconds.
func (d Duration) Milliseconds() int64 {
	return time.Duration(d).Milliseconds()
}

// NewDuration creates a Duration from a time.Duration.
func NewDuration(d time.Duration) Duration {
	return Duration(d)
}
