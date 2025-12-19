// Package clients provides client interfaces and registry for framework services.
package clients

// Client is the interface all framework clients must implement.
type Client interface {
	// Name returns the unique name of the client.
	Name() string

	// Initialize sets up the client with configuration.
	Initialize(cfg any) error

	// Health checks if the client is healthy.
	Health() error

	// Shutdown gracefully shuts down the client.
	Shutdown() error
}

// Configurable is an optional interface for clients that support typed configuration.
type Configurable interface {
	// ConfigType returns the type of configuration the client expects.
	ConfigType() any
}

// Startable is an optional interface for clients that need to be started.
type Startable interface {
	// Start starts the client after initialization.
	Start() error
}

// Stoppable is an optional interface for clients that need to be stopped before shutdown.
type Stoppable interface {
	// Stop stops the client.
	Stop() error
}

// Reloadable is an optional interface for clients that support configuration reload.
type Reloadable interface {
	// Reload reloads the client configuration.
	Reload(cfg any) error
}

// BaseClient provides a base implementation of the Client interface.
// Embed this in your client implementations for common functionality.
type BaseClient struct {
	name        string
	initialized bool
}

// NewBaseClient creates a new base client with the given name.
func NewBaseClient(name string) BaseClient {
	return BaseClient{name: name}
}

// Name returns the client name.
func (b *BaseClient) Name() string {
	return b.name
}

// Initialize marks the client as initialized.
// Override this method in your implementation.
func (b *BaseClient) Initialize(cfg any) error {
	b.initialized = true
	return nil
}

// Health returns nil (healthy).
// Override this method in your implementation.
func (b *BaseClient) Health() error {
	return nil
}

// Shutdown marks the client as not initialized.
// Override this method in your implementation.
func (b *BaseClient) Shutdown() error {
	b.initialized = false
	return nil
}

// IsInitialized returns whether the client has been initialized.
func (b *BaseClient) IsInitialized() bool {
	return b.initialized
}
