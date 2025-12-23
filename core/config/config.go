// Package config provides configuration management for the Codo Framework.
package config

import "fmt"

// Config holds all framework configuration
type Config struct {
	Service    ServiceConfig    `yaml:"service"`
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Auth       AuthConfig       `yaml:"auth"`
	RabbitMQ   RabbitMQConfig   `yaml:"rabbitmq"`
	Features   FeaturesConfig   `yaml:"features"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Errors     ErrorsConfig     `yaml:"errors"`
	DevMode    bool             `yaml:"dev_mode"` // Loaded from YAML, overridable by env/CLI

	// Extensions captures any additional app-specific config sections
	// The ,inline tag merges unknown fields into this map instead of discarding them
	Extensions map[string]interface{} `yaml:",inline"`
}

// NewWithDefaults creates a new Config with all default values
func NewWithDefaults() *Config {
	return &Config{
		Service:    DefaultServiceConfig(),
		Server:     DefaultServerConfig(),
		Database:   DefaultDatabaseConfig(),
		Auth:       DefaultAuthConfig(),
		RabbitMQ:   DefaultRabbitMQConfig(),
		Features:   DefaultFeaturesConfig(),
		Middleware: DefaultMiddlewareConfig(),
		Errors:     DefaultErrorsConfig(),
		DevMode:    false,
		Extensions: make(map[string]interface{}),
	}
}

// IsDevMode returns whether development mode is enabled
func (c *Config) IsDevMode() bool {
	return c.DevMode
}

// Validate validates all configuration sections
func (c *Config) Validate() error {
	if err := c.Service.Validate(); err != nil {
		return err
	}
	if err := c.Server.Validate(); err != nil {
		return err
	}
	if err := c.Database.Validate(); err != nil {
		return err
	}
	if err := c.Auth.Validate(); err != nil {
		return err
	}

	// Validate RabbitMQ based on feature toggle
	if c.Features.IsEnabled(FeatureRabbitMQ) {
		// Feature is enabled - config must be valid
		if !c.RabbitMQ.IsEnabled() {
			return fmt.Errorf("rabbitmq feature is enabled but no configuration provided (missing URL or host)")
		}
		if err := c.RabbitMQ.Validate(); err != nil {
			return err
		}
	}
	// Note: Warning for config-set-but-disabled case is logged in bootstrap

	return nil
}

// IsProd returns true if running in production mode (not dev mode)
func (c *Config) IsProd() bool {
	return !c.DevMode
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := &Config{
		Service:    c.Service,
		Server:     c.Server,
		Database:   c.Database,
		Auth:       c.Auth,
		RabbitMQ:   c.RabbitMQ,
		Middleware: c.Middleware,
		Features: FeaturesConfig{
			DisabledFeatures: make([]string, len(c.Features.DisabledFeatures)),
		},
		DevMode:    c.DevMode,
		Extensions: make(map[string]interface{}),
	}
	copy(clone.Features.DisabledFeatures, c.Features.DisabledFeatures)

	// Deep copy Extensions map
	for k, v := range c.Extensions {
		clone.Extensions[k] = v
	}

	return clone
}
