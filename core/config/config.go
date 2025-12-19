// Package config provides configuration management for the Codo Framework.
package config

// Config holds all framework configuration
type Config struct {
	Service  ServiceConfig  `yaml:"service"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Features FeaturesConfig `yaml:"features"`
	DevMode  bool           `yaml:"-"` // Not loaded from YAML
}

// NewWithDefaults creates a new Config with all default values
func NewWithDefaults() *Config {
	return &Config{
		Service:  DefaultServiceConfig(),
		Server:   DefaultServerConfig(),
		Database: DefaultDatabaseConfig(),
		Auth:     DefaultAuthConfig(),
		RabbitMQ: DefaultRabbitMQConfig(),
		Features: DefaultFeaturesConfig(),
		DevMode:  false,
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
	// Only validate RabbitMQ if enabled
	if c.RabbitMQ.IsEnabled() {
		if err := c.RabbitMQ.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// IsProd returns true if running in production mode (not dev mode)
func (c *Config) IsProd() bool {
	return !c.DevMode
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := &Config{
		Service:  c.Service,
		Server:   c.Server,
		Database: c.Database,
		Auth:     c.Auth,
		RabbitMQ: c.RabbitMQ,
		Features: FeaturesConfig{
			DisabledFeatures: make([]string, len(c.Features.DisabledFeatures)),
		},
		DevMode: c.DevMode,
	}
	copy(clone.Features.DisabledFeatures, c.Features.DisabledFeatures)
	return clone
}
