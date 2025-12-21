package config

// ServiceConfig holds service runtime configuration
type ServiceConfig struct {
	Environment string `yaml:"environment"`
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Environment: "development",
	}
}

// Validate validates service configuration
func (c *ServiceConfig) Validate() error {
	// Environment is optional, no validation needed
	return nil
}
