package app

import (
	"context"
	"fmt"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/core/middleware"
)

// httpServerApp implements HTTPApp for full multi-router server
type httpServerApp struct {
	*foundation
	server *http.Server
	mode   AppMode
}

// Server returns the HTTP server
func (a *httpServerApp) Server() *http.Server {
	return a.server
}

// Start starts the HTTP server
func (a *httpServerApp) Start(ctx context.Context) error {
	return a.server.Start()
}

// Mode returns the bootstrap mode
func (a *httpServerApp) Mode() AppMode {
	return a.mode
}

// singleRouterApp implements SingleRouterApp for single router mode
type singleRouterApp struct {
	*foundation
	router *http.Router
	scope  http.RouterScope
	mode   AppMode
}

// Router returns the single HTTP router
func (a *singleRouterApp) Router() *http.Router {
	return a.router
}

// Scope returns the router scope
func (a *singleRouterApp) Scope() http.RouterScope {
	return a.scope
}

// Start starts the router
func (a *singleRouterApp) Start(ctx context.Context) error {
	return a.router.Start()
}

// Mode returns the bootstrap mode
func (a *singleRouterApp) Mode() AppMode {
	return a.mode
}

// bootstrapHTTPServer creates a full multi-router HTTP server
func bootstrapHTTPServer(cfg *config.Config, opts BootstrapOptions) (BaseApp, error) {
	// Phase 1: Initialize foundation (clients)
	foundation, err := initFoundation(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("foundation init: %w", err)
	}

	// Phase 2: Run migrations
	if err := runMigrations(foundation, opts.MigrationAdder); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	// Phase 3: Create HTTP server with all routers
	server := http.NewServer(&http.ServerConfig{
		PublicAddr:    cfg.Server.PublicAddr(),
		ProtectedAddr: cfg.Server.ProtectedAddr(),
		HiddenAddr:    cfg.Server.HiddenAddr(),
		ShutdownGrace: cfg.Server.ShutdownGrace.Duration(),
	})

	// Initialize and apply middleware to all routers
	if err := initializeMiddleware(server, cfg); err != nil {
		return nil, fmt.Errorf("middleware init: %w", err)
	}

	// Make config accessible to handlers (must be before HandlerRegistrar so handlers can read config)
	http.SetGlobalConfig(cfg)

	// Register handlers
	if opts.HandlerRegistrar != nil {
		if err := opts.HandlerRegistrar(); err != nil {
			return nil, fmt.Errorf("handler registration: %w", err)
		}
	}

	// Prepare routes (doesn't start server)
	if err := server.PrepareRoutes(); err != nil {
		return nil, fmt.Errorf("prepare routes: %w", err)
	}

	return &httpServerApp{
		foundation: foundation,
		server:     server,
		mode:       HTTPServer,
	}, nil
}

// bootstrapHTTPRouter creates a single HTTP router
func bootstrapHTTPRouter(cfg *config.Config, opts BootstrapOptions) (BaseApp, error) {
	// Validate router scope
	if opts.RouterScope == nil {
		return nil, fmt.Errorf("RouterScope is required for HTTPRouter mode")
	}

	// Phase 1: Initialize foundation
	foundation, err := initFoundation(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("foundation init: %w", err)
	}

	// Phase 2: Run migrations
	if err := runMigrations(foundation, opts.MigrationAdder); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	// Phase 3: Create single router
	scope := *opts.RouterScope
	addr := getAddressForScope(cfg, scope)
	router := http.NewRouter(scope, addr)

	// Initialize middleware orchestrator
	orchestrator := middleware.NewOrchestrator(cfg)
	if err := orchestrator.Initialize(); err != nil {
		return nil, fmt.Errorf("middleware init: %w", err)
	}

	// Apply middleware for this router only
	routerType := routerTypeFromScope(scope)
	orchestrator.Apply(router, routerType)

	// Make config accessible to handlers (must be before HandlerRegistrar so handlers can read config)
	http.SetGlobalConfig(cfg)

	// Register handlers
	if opts.HandlerRegistrar != nil {
		if err := opts.HandlerRegistrar(); err != nil {
			return nil, fmt.Errorf("handler registration: %w", err)
		}
	}

	// Prepare routes
	if err := router.RegisterHandlers(); err != nil {
		return nil, fmt.Errorf("prepare routes: %w", err)
	}

	return &singleRouterApp{
		foundation: foundation,
		router:     router,
		scope:      scope,
		mode:       HTTPRouter,
	}, nil
}

// bootstrapRouteInspector creates server for route introspection (doesn't start)
func bootstrapRouteInspector(cfg *config.Config, opts BootstrapOptions) (BaseApp, error) {
	// Similar to HTTPServer but optimized for introspection
	// Can skip some initialization if needed in the future
	return bootstrapHTTPServer(cfg, opts)
}

// getAddressForScope returns the address for a given router scope
func getAddressForScope(cfg *config.Config, scope http.RouterScope) string {
	switch scope {
	case http.ScopePublic:
		return cfg.Server.PublicAddr()
	case http.ScopeProtected:
		return cfg.Server.ProtectedAddr()
	case http.ScopeHidden:
		return cfg.Server.HiddenAddr()
	default:
		return fmt.Sprintf(":%d", scope.DefaultPort())
	}
}

// routerTypeFromScope maps http.RouterScope to middleware.Router
func routerTypeFromScope(scope http.RouterScope) middleware.Router {
	switch scope {
	case http.ScopePublic:
		return middleware.RouterPublic
	case http.ScopeProtected:
		return middleware.RouterProtected
	case http.ScopeHidden:
		return middleware.RouterHidden
	default:
		return middleware.RouterPublic
	}
}
