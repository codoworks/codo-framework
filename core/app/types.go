package app

import (
	"context"
	"sync"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/db/migrations"
	"github.com/codoworks/codo-framework/core/db/seeds"
	"github.com/codoworks/codo-framework/core/http"
)

// AppMode defines the type of application to bootstrap
type AppMode int

const (
	// HTTPServer bootstraps a full multi-router HTTP server
	// - Creates all 3 routers (public, protected, hidden)
	// - Initializes middleware on all routers
	// - Registers handlers on all routers
	// - Ready to start serving HTTP traffic
	HTTPServer AppMode = iota

	// HTTPRouter bootstraps a single HTTP router
	// - Creates only one router (scope specified in options)
	// - Initializes middleware for that router only
	// - Registers handlers for that scope only
	// - Efficient for single-router deployment
	HTTPRouter

	// WorkerDaemon bootstraps a background worker process
	// - NO HTTP server
	// - Initializes clients (DB, RabbitMQ, etc.)
	// - Registers background workers/consumers
	// - For queue processing, scheduled tasks, etc.
	WorkerDaemon

	// RouteInspector bootstraps for route introspection
	// - Creates HTTP server but doesn't start it
	// - Registers all routes for display
	// - Used by `info routes` command
	RouteInspector

	// ConfigInspector bootstraps for configuration introspection
	// - Loads config (YAML + defaults + .env + env overrides)
	// - Initializes EnvVarRegistry (if registrar provided)
	// - NO client initialization
	// - NO migrations
	// - NO HTTP server
	// - Used by `info env` command
	ConfigInspector
)

// String returns the mode name for logging
func (m AppMode) String() string {
	switch m {
	case HTTPServer:
		return "HTTPServer"
	case HTTPRouter:
		return "HTTPRouter"
	case WorkerDaemon:
		return "WorkerDaemon"
	case RouteInspector:
		return "RouteInspector"
	case ConfigInspector:
		return "ConfigInspector"
	default:
		return "Unknown"
	}
}

// MigrationAdder is a function that adds migrations to a runner
type MigrationAdder func(*migrations.Runner)

// SeedAdder is a function that adds database seeds to a seeder
type SeedAdder func(*seeds.Seeder)

// HandlerRegistrar is a function that registers HTTP handlers
// It receives a database client for handler initialization
type HandlerRegistrar func() error

// CustomClientInitializer is a function that registers and initializes custom clients
type CustomClientInitializer func(cfg *config.Config) error

// WorkerRegistrar is a function that registers background workers
// It receives a DaemonApp to register workers with
type WorkerRegistrar func(daemon DaemonApp) error

// EnvVarRegistrar is a function that registers consumer environment variable requirements
// It is called during bootstrap before config loading to declare env var requirements
// The registered env vars are validated and resolved before client initialization
type EnvVarRegistrar func(registry *config.EnvVarRegistry) error

// BootstrapOptions configures application bootstrap
type BootstrapOptions struct {
	// Mode determines what type of app to bootstrap
	Mode AppMode

	// RouterScope specifies which router to create (required for HTTPRouter mode)
	RouterScope *http.RouterScope

	// HandlerRegistrar registers HTTP handlers (used by HTTPServer, HTTPRouter, RouteInspector)
	HandlerRegistrar HandlerRegistrar

	// MigrationAdder adds database migrations (used by all modes except RouteInspector)
	MigrationAdder MigrationAdder

	// SeedAdder adds database seeds (optional, used by db seed command)
	SeedAdder SeedAdder

	// CustomClientInit initializes custom clients (optional)
	CustomClientInit CustomClientInitializer

	// WorkerRegistrar registers background workers (required for WorkerDaemon mode)
	WorkerRegistrar WorkerRegistrar

	// EnvVarRegistrar registers consumer environment variable requirements (optional)
	// Called during bootstrap to declare env vars that will be validated and resolved
	// before client initialization. Resolved values are stored in Config.EnvRegistry.
	EnvVarRegistrar EnvVarRegistrar
}

// Initializer is a function that returns bootstrap options for the consumer application.
// The consumer app registers this in their init() function.
// The Mode field will be set by the CLI command, so consumers should not set it.
type Initializer func(*config.Config) (BootstrapOptions, error)

var (
	registeredInitializer Initializer
	initMu                sync.RWMutex
)

// RegisterInitializer registers the consumer application's initializer function.
// This should be called in the consumer's init() function.
func RegisterInitializer(init Initializer) {
	initMu.Lock()
	defer initMu.Unlock()
	registeredInitializer = init
}

// GetInitializer returns the registered initializer function.
// This is used internally by the framework CLI commands.
func GetInitializer() Initializer {
	initMu.RLock()
	defer initMu.RUnlock()
	return registeredInitializer
}

// BaseApp is the base interface for all application types
type BaseApp interface {
	// Config returns the application configuration
	Config() *config.Config

	// Shutdown gracefully shuts down the application
	Shutdown(ctx context.Context) error

	// Mode returns the bootstrap mode used
	Mode() AppMode
}

// HTTPApp extends BaseApp for HTTP server modes
type HTTPApp interface {
	BaseApp

	// Server returns the HTTP server
	Server() *http.Server

	// Start starts the HTTP server
	Start(ctx context.Context) error
}

// SingleRouterApp extends BaseApp for single router mode
type SingleRouterApp interface {
	BaseApp

	// Router returns the single HTTP router
	Router() *http.Router

	// Scope returns the router scope
	Scope() http.RouterScope

	// Start starts the router
	Start(ctx context.Context) error
}

// DaemonApp extends BaseApp for worker daemon mode
type DaemonApp interface {
	BaseApp

	// RegisterWorker registers a background worker
	RegisterWorker(worker Worker)

	// Start starts all registered workers
	Start(ctx context.Context) error

	// Stop stops all workers
	Stop(ctx context.Context) error
}

// ConfigApp extends BaseApp for config inspection mode
type ConfigApp interface {
	BaseApp

	// EnvRegistry returns the resolved environment variable registry
	// Returns nil if no EnvVarRegistrar was provided during bootstrap
	EnvRegistry() *config.EnvVarRegistry
}

// Worker represents a background worker process
type Worker interface {
	// Name returns the worker name for logging
	Name() string

	// Start starts the worker (blocks until context is cancelled)
	Start(ctx context.Context) error

	// Stop stops the worker gracefully
	Stop(ctx context.Context) error
}
