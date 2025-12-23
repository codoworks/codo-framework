package app

import (
	"time"

	"github.com/codoworks/codo-framework/clients/kratos"
	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/clients/rabbitmq"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/core/errors"
	"github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/core/middleware"

	// Import middleware packages to trigger auto-registration
	_ "github.com/codoworks/codo-framework/core/middleware/auth"
	_ "github.com/codoworks/codo-framework/core/middleware/cors"
	_ "github.com/codoworks/codo-framework/core/middleware/gzip"
	_ "github.com/codoworks/codo-framework/core/middleware/logger"
	_ "github.com/codoworks/codo-framework/core/middleware/pagination"
	_ "github.com/codoworks/codo-framework/core/middleware/recover"
	_ "github.com/codoworks/codo-framework/core/middleware/requestid"
	_ "github.com/codoworks/codo-framework/core/middleware/timeout"
	_ "github.com/codoworks/codo-framework/core/middleware/xss"
)

// Bootstrap initializes an application based on the specified mode
func Bootstrap(cfg *config.Config, opts BootstrapOptions) (BaseApp, error) {
	if cfg == nil {
		return nil, errors.BadRequest("Config is required").
			WithPhase(errors.PhaseBootstrap)
	}

	// Validate options based on mode
	if err := validateOptions(opts); err != nil {
		return nil, errors.WrapBadRequest(err, "Invalid bootstrap options").
			WithPhase(errors.PhaseBootstrap).
			WithDetail("mode", opts.Mode.String())
	}

	// Log bootstrap mode in dev mode
	if cfg.IsDevMode() {
		log := getOrCreateLogger()
		log.Infof("[Bootstrap] Initializing in %s mode", opts.Mode)
	}

	// Delegate to mode-specific bootstrap functions
	switch opts.Mode {
	case HTTPServer:
		return bootstrapHTTPServer(cfg, opts)

	case HTTPRouter:
		return bootstrapHTTPRouter(cfg, opts)

	case WorkerDaemon:
		return bootstrapWorkerDaemon(cfg, opts)

	case RouteInspector:
		return bootstrapRouteInspector(cfg, opts)

	case ConfigInspector:
		return bootstrapConfigInspector(cfg, opts)

	default:
		return nil, errors.BadRequest("Unknown bootstrap mode").
			WithPhase(errors.PhaseBootstrap).
			WithDetail("mode", opts.Mode.String())
	}
}

// validateOptions validates bootstrap options based on mode
func validateOptions(opts BootstrapOptions) error {
	switch opts.Mode {
	case HTTPRouter:
		if opts.RouterScope == nil {
			return errors.BadRequest("RouterScope is required for HTTPRouter mode").
				WithPhase(errors.PhaseBootstrap)
		}

	case WorkerDaemon:
		if opts.WorkerRegistrar == nil {
			return errors.BadRequest("WorkerRegistrar is required for WorkerDaemon mode").
				WithPhase(errors.PhaseBootstrap)
		}

	case HTTPServer, RouteInspector:
		if opts.HandlerRegistrar == nil {
			return errors.BadRequest("HandlerRegistrar is required").
				WithPhase(errors.PhaseBootstrap).
				WithDetail("mode", opts.Mode.String())
		}

	case ConfigInspector:
		// No special requirements - all options are optional
	}

	return nil
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

	// Conditionally register Kratos based on feature toggle
	if cfg.Features.IsEnabled(config.FeatureKratos) {
		clients.MustRegister(kratos.New())
		log.Info("Kratos client registered")
	} else {
		log.Warn("Kratos client disabled via feature toggle")
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

	// Only add Kratos config if registered
	if clients.Has(kratos.ClientName) {
		clientConfigs[kratos.ClientName] = &kratos.ClientConfig{
			PublicURL:  cfg.Auth.KratosPublicURL,
			AdminURL:   cfg.Auth.KratosAdminURL,
			CookieName: cfg.Auth.SessionCookie,
			Timeout:    10 * time.Second,
		}
	}

	// Initialize with enhanced error handling
	if err := clients.InitializeAllWithMetadata(clientConfigs, log); err != nil {
		return errors.WrapInternal(err, "Failed to initialize clients").
			WithPhase(errors.PhaseClient)
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
		Name:        kratos.ClientName,
		Requirement: clients.ClientOptional,
		FeatureName: config.FeatureKratos,
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

// initializeMiddleware initializes and applies middleware to all routers
func initializeMiddleware(server *http.Server, cfg *config.Config) error {
	// Import middleware packages to trigger auto-registration
	// Note: This is handled by blank imports at the top of the file

	// Create middleware orchestrator
	orchestrator := middleware.NewOrchestrator(cfg)

	// Initialize all registered middleware
	if err := orchestrator.Initialize(); err != nil {
		return errors.WrapInternal(err, "Failed to initialize middleware").
			WithPhase(errors.PhaseMiddleware)
	}

	// Apply middleware to each router based on scope
	orchestrator.Apply(server.PublicRouter(), middleware.RouterPublic)
	orchestrator.Apply(server.ProtectedRouter(), middleware.RouterProtected)
	orchestrator.Apply(server.HiddenRouter(), middleware.RouterHidden)

	// Dev mode: Print active middleware
	if cfg.IsDevMode() {
		printMiddlewareStatus(orchestrator)
	}

	return nil
}

// printMiddlewareStatus prints active middleware in dev mode
func printMiddlewareStatus(orchestrator *middleware.Orchestrator) {
	log := getOrCreateLogger()

	log.Info("===========================================")
	log.Info("MIDDLEWARE STATUS")
	log.Info("===========================================")

	routers := []struct {
		name       string
		routerType middleware.Router
	}{
		{"Public Router (8081)", middleware.RouterPublic},
		{"Protected Router (8080)", middleware.RouterProtected},
		{"Hidden Router (8079)", middleware.RouterHidden},
	}

	for _, r := range routers {
		middlewares := orchestrator.List(r.routerType)
		log.Infof("\n  %s:", r.name)
		if len(middlewares) == 0 {
			log.Info("    (none)")
		} else {
			for _, m := range middlewares {
				log.Infof("    %3d  %-15s [%s]  ✓ enabled",
					m.Priority(),
					m.Name(),
					priorityCategory(m.Priority()),
				)
			}
		}
	}

	log.Info("===========================================")
}

// priorityCategory returns the category label for a priority value
func priorityCategory(priority int) string {
	if priority >= middleware.PriorityCoreMin && priority <= middleware.PriorityCoreMax {
		return "core"
	}
	if priority >= middleware.PriorityFeatureMin && priority <= middleware.PriorityFeatureMax {
		return "feature"
	}
	if priority >= middleware.PriorityConsumerMin && priority <= middleware.PriorityConsumerMax {
		return "consumer"
	}
	return "unknown"
}
