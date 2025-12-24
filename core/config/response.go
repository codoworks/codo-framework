package config

// ResponseConfig holds configuration for API response formatting
type ResponseConfig struct {
	// Strict mode includes all fields in responses (no omitempty behavior)
	// When true: null fields are serialized as null, empty arrays as []
	// When false: null/empty fields are omitted from JSON (default)
	Strict bool `yaml:"strict"`
}

// DefaultResponseConfig returns default response configuration
func DefaultResponseConfig() ResponseConfig {
	return ResponseConfig{
		Strict: false,
	}
}
