// Package codo provides the main framework entry point for the Codo Framework.
package codo

import (
	"context"
	"fmt"

	"github.com/codoworks/codo-framework/core/clients"
)

// Version is the framework version, set at build time.
var Version = "0.0.0-dev"

// State represents the framework state.
type State int

const (
	// StateNew indicates the framework is newly created.
	StateNew State = iota
	// StateInitialized indicates the framework is initialized.
	StateInitialized
	// StateRunning indicates the framework is running.
	StateRunning
	// StateStopped indicates the framework is stopped.
	StateStopped
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateNew:
		return "new"
	case StateInitialized:
		return "initialized"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Config holds framework configuration.
type Config struct {
	// Name is the application name.
	Name string `json:"name" yaml:"name"`

	// Environment is the application environment (e.g., development, production).
	Environment string `json:"environment" yaml:"environment"`

	// Clients holds client configurations keyed by client name.
	Clients map[string]any `json:"clients" yaml:"clients"`
}

// DefaultConfig returns default framework configuration.
func DefaultConfig() *Config {
	return &Config{
		Name:        "codo-app",
		Environment: "development",
		Clients:     make(map[string]any),
	}
}

// Framework represents the main framework instance.
type Framework struct {
	config *Config
	state  State
}

// New creates a new framework instance.
func New() *Framework {
	return &Framework{
		config: DefaultConfig(),
		state:  StateNew,
	}
}

// NewWithConfig creates a new framework instance with configuration.
func NewWithConfig(cfg *Config) *Framework {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Framework{
		config: cfg,
		state:  StateNew,
	}
}

// Config returns the framework configuration.
func (f *Framework) Config() *Config {
	return f.config
}

// State returns the current framework state.
func (f *Framework) State() State {
	return f.state
}

// IsInitialized returns true if the framework is initialized.
func (f *Framework) IsInitialized() bool {
	return f.state >= StateInitialized
}

// IsRunning returns true if the framework is running.
func (f *Framework) IsRunning() bool {
	return f.state == StateRunning
}

// Initialize sets up the framework and all registered clients.
func (f *Framework) Initialize() error {
	if f.state >= StateInitialized {
		return nil
	}

	// Initialize all registered clients
	if err := clients.InitializeAll(f.config.Clients); err != nil {
		return fmt.Errorf("failed to initialize clients: %w", err)
	}

	f.state = StateInitialized
	return nil
}

// Run starts the framework.
// This is a placeholder for HTTP server startup in future workstreams.
func (f *Framework) Run() error {
	if f.state < StateInitialized {
		if err := f.Initialize(); err != nil {
			return err
		}
	}

	if f.state == StateRunning {
		return fmt.Errorf("framework is already running")
	}

	f.state = StateRunning
	return nil
}

// Shutdown gracefully shuts down the framework.
func (f *Framework) Shutdown(ctx context.Context) error {
	if f.state == StateStopped {
		return nil
	}

	// Shutdown all clients
	if err := clients.ShutdownAll(); err != nil {
		return fmt.Errorf("failed to shutdown clients: %w", err)
	}

	f.state = StateStopped
	return nil
}

// Health checks the health of all framework components.
func (f *Framework) Health() error {
	return clients.HealthCheck()
}

// HealthAll returns health status of all clients.
func (f *Framework) HealthAll() map[string]error {
	return clients.HealthAll()
}

// GetVersion returns the framework version.
func GetVersion() string {
	return Version
}
