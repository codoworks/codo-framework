package app

import (
	"context"

	"github.com/codoworks/codo-framework/core/config"
)

// configApp implements ConfigApp for config inspection mode
type configApp struct {
	cfg         *config.Config
	envRegistry *config.EnvVarRegistry
	mode        AppMode
}

// Config returns the application configuration
func (a *configApp) Config() *config.Config {
	return a.cfg
}

// EnvRegistry returns the resolved environment variable registry
func (a *configApp) EnvRegistry() *config.EnvVarRegistry {
	return a.envRegistry
}

// Shutdown performs cleanup (no-op for ConfigInspector)
func (a *configApp) Shutdown(ctx context.Context) error {
	return nil
}

// Mode returns the bootstrap mode
func (a *configApp) Mode() AppMode {
	return a.mode
}

// bootstrapConfigInspector creates an app for config introspection
// This is the lightest bootstrap mode - only config and env vars, no clients or side effects
func bootstrapConfigInspector(cfg *config.Config, opts BootstrapOptions) (BaseApp, error) {
	var envRegistry *config.EnvVarRegistry

	// Initialize env vars if registrar provided
	if opts.EnvVarRegistrar != nil {
		registry := config.NewEnvVarRegistry()

		// Call consumer's registrar to declare env var requirements
		if err := opts.EnvVarRegistrar(registry); err != nil {
			// For inspection, log the error but continue
			// This allows showing partial information
		}

		// Resolve all registered env vars
		if err := registry.Resolve(); err != nil {
			// For inspection, we continue even with errors
			// The registry will contain whatever could be resolved
		}

		envRegistry = registry
		cfg.EnvRegistry = registry
	}

	return &configApp{
		cfg:         cfg,
		envRegistry: envRegistry,
		mode:        ConfigInspector,
	}, nil
}
