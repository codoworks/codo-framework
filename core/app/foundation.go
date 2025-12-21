package app

import (
	"context"
	"fmt"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/core/db/migrations"
)

// foundation holds the core infrastructure components (clients, config)
type foundation struct {
	config *config.Config
}

// initFoundation initializes the foundation layer (clients and config)
func initFoundation(cfg *config.Config, opts BootstrapOptions) (*foundation, error) {
	// 1. Register and initialize framework clients
	if err := registerFrameworkClients(cfg); err != nil {
		return nil, fmt.Errorf("failed to register framework clients: %w", err)
	}

	// 2. Register and initialize custom clients
	if opts.CustomClientInit != nil {
		if err := opts.CustomClientInit(cfg); err != nil {
			return nil, fmt.Errorf("failed to initialize custom clients: %w", err)
		}
	}

	return &foundation{
		config: cfg,
	}, nil
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
		return fmt.Errorf("failed to get database client: %w", err)
	}

	// Create migration runner
	runner := migrations.NewRunner(dbClient.DB())

	// Add migrations
	adder(runner)

	// Run migrations
	if _, err := runner.Up(context.Background()); err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}

	return nil
}
