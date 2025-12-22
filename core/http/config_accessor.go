package http

import "github.com/codoworks/codo-framework/core/config"

var globalConfig *config.Config

// SetGlobalConfig stores the config for handlers to access.
// This is called during bootstrap before handler initialization.
func SetGlobalConfig(cfg *config.Config) {
	globalConfig = cfg
}

// GetGlobalConfig returns the current config.
// Returns nil if SetGlobalConfig has not been called yet.
func GetGlobalConfig() *config.Config {
	return globalConfig
}
