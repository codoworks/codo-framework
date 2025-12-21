package config

// DefaultConfig returns a Config with all default values set
// This is an alias for NewWithDefaults for clarity
func DefaultConfig() *Config {
	return NewWithDefaults()
}

// ApplyDefaults applies default values to any zero-valued fields in the config
// This is useful when loading partial configs
func (c *Config) ApplyDefaults() {
	defaults := NewWithDefaults()

	// Service defaults
	if c.Service.Environment == "" {
		c.Service.Environment = defaults.Service.Environment
	}

	// Server defaults
	if c.Server.PublicPort == 0 {
		c.Server.PublicPort = defaults.Server.PublicPort
	}
	if c.Server.ProtectedPort == 0 {
		c.Server.ProtectedPort = defaults.Server.ProtectedPort
	}
	if c.Server.HiddenPort == 0 {
		c.Server.HiddenPort = defaults.Server.HiddenPort
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = defaults.Server.ReadTimeout
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = defaults.Server.WriteTimeout
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = defaults.Server.IdleTimeout
	}
	if c.Server.ShutdownGrace == 0 {
		c.Server.ShutdownGrace = defaults.Server.ShutdownGrace
	}

	// Database defaults
	if c.Database.Driver == "" {
		c.Database.Driver = defaults.Database.Driver
	}
	if c.Database.Host == "" {
		c.Database.Host = defaults.Database.Host
	}
	if c.Database.Port == 0 {
		c.Database.Port = defaults.Database.Port
	}
	if c.Database.Name == "" {
		c.Database.Name = defaults.Database.Name
	}
	if c.Database.User == "" {
		c.Database.User = defaults.Database.User
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = defaults.Database.SSLMode
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = defaults.Database.MaxOpenConns
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = defaults.Database.MaxIdleConns
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = defaults.Database.ConnMaxLifetime
	}

	// Auth defaults
	if c.Auth.KratosPublicURL == "" {
		c.Auth.KratosPublicURL = defaults.Auth.KratosPublicURL
	}
	if c.Auth.KratosAdminURL == "" {
		c.Auth.KratosAdminURL = defaults.Auth.KratosAdminURL
	}
	if c.Auth.KetoReadURL == "" {
		c.Auth.KetoReadURL = defaults.Auth.KetoReadURL
	}
	if c.Auth.KetoWriteURL == "" {
		c.Auth.KetoWriteURL = defaults.Auth.KetoWriteURL
	}
	if c.Auth.SessionCookie == "" {
		c.Auth.SessionCookie = defaults.Auth.SessionCookie
	}

	// Features defaults - DisabledFeatures nil check
	if c.Features.DisabledFeatures == nil {
		c.Features.DisabledFeatures = defaults.Features.DisabledFeatures
	}
}
