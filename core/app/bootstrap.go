package app

import (
	"context"
	"fmt"
	"time"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/clients/rabbitmq"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/core/db/migrations"
	"github.com/codoworks/codo-framework/core/http"
)

// MigrationAdder is a function that adds migrations to a runner
type MigrationAdder func(*migrations.Runner)

// HandlerRegistrar is a function that registers handlers
type HandlerRegistrar func(*db.Client)

// CustomClientInitializer is a function that registers and initializes custom clients
type CustomClientInitializer func(cfg *config.Config) error

// BootstrapOptions holds initialization options
type BootstrapOptions struct {
	Config                  *config.Config
	MigrationAdder          MigrationAdder
	HandlerRegistrar        HandlerRegistrar
	CustomClientInitializer CustomClientInitializer
}

// Bootstrap handles common initialization patterns for applications.
// It registers framework clients, initializes custom clients, runs migrations,
// creates the HTTP server, and registers handlers.
func Bootstrap(opts BootstrapOptions) (*http.Server, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// 1. Register and initialize framework clients
	if err := registerFrameworkClients(opts.Config); err != nil {
		return nil, fmt.Errorf("failed to register framework clients: %w", err)
	}

	// 2. Register and initialize custom clients
	if opts.CustomClientInitializer != nil {
		if err := opts.CustomClientInitializer(opts.Config); err != nil {
			return nil, fmt.Errorf("failed to initialize custom clients: %w", err)
		}
	}

	// 3. Run migrations
	if opts.MigrationAdder != nil {
		dbClient := clients.MustGetTyped[*db.Client]("db")
		runner := migrations.NewRunner(dbClient.DB())
		opts.MigrationAdder(runner)
		if _, err := runner.Up(context.Background()); err != nil {
			return nil, fmt.Errorf("migration execution failed: %w", err)
		}
	}

	// 4. Create HTTP server
	server := http.NewServer(&http.ServerConfig{
		PublicAddr:    opts.Config.Server.PublicAddr(),
		ProtectedAddr: opts.Config.Server.ProtectedAddr(),
		HiddenAddr:    opts.Config.Server.HiddenAddr(),
		ShutdownGrace: opts.Config.Server.ShutdownGrace.Duration(),
	})

	// 5. Register handlers
	if opts.HandlerRegistrar != nil {
		dbClient := clients.MustGetTyped[*db.Client]("db")
		opts.HandlerRegistrar(dbClient)
	}

	return server, nil
}

// registerFrameworkClients registers and initializes standard framework clients
func registerFrameworkClients(cfg *config.Config) error {
	log := getOrCreateLogger() // Get logger early for error reporting

	// Register metadata for framework clients
	registerFrameworkClientMetadata()

	// Always register logger (required)
	clients.MustRegister(logger.New())

	// Conditionally register RabbitMQ based on feature toggle
	if cfg.Features.IsEnabled(config.FeatureRabbitMQ) {
		clients.MustRegister(rabbitmq.New())
		log.Info("RabbitMQ client registered")
	} else {
		log.Warn("RabbitMQ client disabled via feature toggle")
		// Warn if config is set but feature is disabled
		if cfg.RabbitMQ.IsEnabled() {
			log.Warnf("RabbitMQ configuration detected (URL set) but feature is disabled")
		}
	}

	// Always register database (required)
	dbClient := db.NewClient(nil)
	clients.MustRegister(dbClient)

	// Build configs only for registered clients
	clientConfigs := make(map[string]any)
	clientConfigs["logger"] = nil
	clientConfigs["db"] = &db.ClientConfig{
		Driver:          cfg.Database.Driver,
		DSN:             cfg.Database.DSN(),
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
	}

	// Only add RabbitMQ config if registered
	if clients.Has(rabbitmq.ClientName) {
		clientConfigs["rabbitmq"] = &cfg.RabbitMQ
	}

	// Initialize with enhanced error handling
	if err := clients.InitializeAllWithMetadata(clientConfigs, log); err != nil {
		return fmt.Errorf("failed to initialize clients: %w", err)
	}

	return nil
}

// registerFrameworkClientMetadata registers metadata for all framework clients
func registerFrameworkClientMetadata() {
	clients.RegisterMetadata(clients.ClientMetadata{
		Name:        logger.ClientName,
		Requirement: clients.ClientRequired,
	})

	clients.RegisterMetadata(clients.ClientMetadata{
		Name:        rabbitmq.ClientName,
		Requirement: clients.ClientOptional,
		FeatureName: config.FeatureRabbitMQ,
	})

	clients.RegisterMetadata(clients.ClientMetadata{
		Name:        "db",
		Requirement: clients.ClientRequired,
	})
}

// getOrCreateLogger gets the logger client if available, or creates a temporary one
func getOrCreateLogger() *logger.Logger {
	if clients.Has(logger.ClientName) {
		log, err := clients.GetTyped[*logger.Logger](logger.ClientName)
		if err == nil {
			return log
		}
	}
	// Create temporary logger for bootstrap logging
	log := logger.New()
	log.Initialize(nil)
	return log
}
