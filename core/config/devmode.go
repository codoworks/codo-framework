package config

// DevModeOverrides applies development mode configuration overrides
func (c *Config) DevModeOverrides() {
	if !c.DevMode {
		return
	}

	// In dev mode, we might want to:
	// - Enable more verbose logging (handled by logger)
	// - Relax CORS (handled by HTTP middleware)
	// - Enable SQL logging (handled by DB client)
	// - Show stack traces (handled by error handler)

	// The config struct itself doesn't change,
	// but other components check IsDevMode()
}

// SetDevMode sets the development mode flag
func (c *Config) SetDevMode(enabled bool) {
	c.DevMode = enabled
	if enabled {
		c.DevModeOverrides()
	}
}
