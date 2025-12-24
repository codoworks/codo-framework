package config

// LoggerConfig holds logger client configuration
type LoggerConfig struct {
	Level  string `yaml:"level"`  // debug|info|warn|error|fatal|panic
	Format string `yaml:"format"` // json|text
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:  "info",
		Format: "json",
	}
}
