package config

import (
	"fmt"
)

// ServiceConfig holds service identity configuration
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Name:        "codo-app",
		Version:     "0.0.0",
		Environment: "development",
	}
}

// Validate validates service configuration
func (c *ServiceConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("service.name is required")
	}
	return nil
}
