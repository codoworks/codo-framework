package app

import (
	"context"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/core/db/migrations"
	"github.com/codoworks/codo-framework/core/errors"
	"github.com/codoworks/codo-framework/core/http"
)

// foundation holds the core infrastructure components (clients, config)
type foundation struct {
	config *config.Config
}

// initFoundation initializes the foundation layer (clients and config)
func initFoundation(cfg *config.Config, opts BootstrapOptions) (*foundation, error) {
	// 0. Configure error handling from config
	configureErrorHandling(cfg)

	// 1. Register and initialize framework clients
	if err := registerFrameworkClients(cfg); err != nil {
		return nil, errors.WrapInternal(err, "Failed to register framework clients").
			WithPhase(errors.PhaseBootstrap)
	}

	// 2. Register and initialize custom clients
	if opts.CustomClientInit != nil {
		if err := opts.CustomClientInit(cfg); err != nil {
			return nil, errors.WrapInternal(err, "Failed to initialize custom clients").
				WithPhase(errors.PhaseClient)
		}
	}

	return &foundation{
		config: cfg,
	}, nil
}

// configureErrorHandling sets up the error handling configuration
func configureErrorHandling(cfg *config.Config) {
	// Configure error capture behavior
	errors.SetCaptureConfig(errors.CaptureConfig{
		StackTraceOn5xx: cfg.Errors.Capture.StackTraceOn5xx,
		StackTraceOn4xx: cfg.Errors.Capture.StackTraceOn4xx,
		StackTraceDepth: cfg.Errors.Capture.StackTraceDepth,
		AutoDetectPhase: cfg.Errors.Capture.AutoDetectPhase,
	})

	// Configure HTTP error response behavior
	http.SetHandlerConfig(http.HandlerConfig{
		ExposeDetails:     cfg.Errors.Handler.ExposeDetails,
		ExposeStackTraces: cfg.Errors.Handler.ExposeStackTraces,
		ResponseFormat:    cfg.Errors.Handler.ResponseFormat,
	})
}

// Config returns the application configuration
func (f *foundation) Config() *config.Config {
	return f.config
}

// Shutdown shuts down all clients gracefully
func (f *foundation) Shutdown(ctx context.Context) error {
	return nil // Client shutdown is handled by defer in commands
}

// runMigrations runs database migrations if a migration adder is provided
func runMigrations(foundation *foundation, adder MigrationAdder) error {
	if adder == nil {
		return nil // No migrations to run
	}

	// Get database client from registry
	dbClient, err := clients.GetTyped[*db.Client]("db")
	if err != nil {
		return errors.WrapInternal(err, "Failed to get database client").
			WithPhase(errors.PhaseMigration)
	}

	// Create migration runner
	runner := migrations.NewRunner(dbClient.DB())

	// Add migrations
	adder(runner)

	// Run migrations
	if _, err := runner.Up(context.Background()); err != nil {
		return errors.WrapInternal(err, "Migration execution failed").
			WithPhase(errors.PhaseMigration)
	}

	return nil
}
