package config

// DevModeOverrides applies development mode configuration overrides
func (c *Config) DevModeOverrides() {
	if !c.DevMode {
		return
	}

	// Auto-enable auth middleware verbose logging in dev mode
	// This links the --dev flag to auth middleware's dev_mode setting
	// NOTE: dev_bypass_auth is NOT auto-enabled - it must be explicitly set
	c.Middleware.Auth.DevMode = true

	// Auto-enable error details and stack traces in dev mode
	// This makes debugging easier during development
	// NOTE: These should NEVER be true in production (security risk)
	c.Errors.Handler.ExposeDetails = true
	c.Errors.Handler.ExposeStackTraces = true
}

// SetDevMode sets the development mode flag
func (c *Config) SetDevMode(enabled bool) {
	c.DevMode = enabled
	if enabled {
		c.DevModeOverrides()
	}
}
