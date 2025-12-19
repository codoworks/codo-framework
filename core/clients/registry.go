package clients

import (
	"fmt"
	"sync"

	"github.com/codoworks/codo-framework/core/registry"
)

var (
	globalRegistry *registry.Registry[Client]
	once           sync.Once
)

func getRegistry() *registry.Registry[Client] {
	once.Do(func() {
		globalRegistry = registry.New[Client]()
	})
	return globalRegistry
}

// ResetRegistry resets the global registry. For testing only.
func ResetRegistry() {
	once = sync.Once{}
	globalRegistry = nil
}

// Register adds a client to the global registry.
func Register(client Client) error {
	return getRegistry().Register(client.Name(), client)
}

// MustRegister adds a client to the global registry and panics on error.
func MustRegister(client Client) {
	if err := Register(client); err != nil {
		panic(err)
	}
}

// Get retrieves a client by name.
func Get(name string) (Client, error) {
	return getRegistry().Get(name)
}

// MustGet retrieves a client or panics.
func MustGet(name string) Client {
	return getRegistry().MustGet(name)
}

// Has checks if a client exists.
func Has(name string) bool {
	return getRegistry().Has(name)
}

// GetTyped retrieves a client with type assertion.
func GetTyped[T Client](name string) (T, error) {
	client, err := Get(name)
	if err != nil {
		var zero T
		return zero, err
	}
	typed, ok := client.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("client %q is not of expected type", name)
	}
	return typed, nil
}

// MustGetTyped retrieves a typed client or panics.
func MustGetTyped[T Client](name string) T {
	client, err := GetTyped[T](name)
	if err != nil {
		panic(err)
	}
	return client
}

// All returns all registered clients.
func All() map[string]Client {
	return getRegistry().All()
}

// Names returns all registered client names.
func Names() []string {
	return getRegistry().Keys()
}

// Count returns the number of registered clients.
func Count() int {
	return getRegistry().Count()
}

// InitializeAll initializes all registered clients with their configurations.
func InitializeAll(configs map[string]any) error {
	for name, client := range getRegistry().All() {
		cfg := configs[name]
		if err := client.Initialize(cfg); err != nil {
			return fmt.Errorf("failed to initialize client %q: %w", name, err)
		}
	}
	return nil
}

// ShutdownAll shuts down all registered clients.
// Returns the last error encountered, if any.
func ShutdownAll() error {
	var lastErr error
	for name, client := range getRegistry().All() {
		if err := client.Shutdown(); err != nil {
			lastErr = fmt.Errorf("failed to shutdown client %q: %w", name, err)
		}
	}
	return lastErr
}

// HealthAll checks health of all clients.
// Returns a map of client names to their health errors (nil if healthy).
func HealthAll() map[string]error {
	results := make(map[string]error)
	for name, client := range getRegistry().All() {
		results[name] = client.Health()
	}
	return results
}

// HealthCheck returns nil if all clients are healthy, or an error if any are unhealthy.
func HealthCheck() error {
	for name, err := range HealthAll() {
		if err != nil {
			return fmt.Errorf("client %q is unhealthy: %w", name, err)
		}
	}
	return nil
}

// Initialize initializes a specific client by name.
func Initialize(name string, cfg any) error {
	client, err := Get(name)
	if err != nil {
		return err
	}
	return client.Initialize(cfg)
}

// Shutdown shuts down a specific client by name.
func Shutdown(name string) error {
	client, err := Get(name)
	if err != nil {
		return err
	}
	return client.Shutdown()
}

// Health checks the health of a specific client by name.
func Health(name string) error {
	client, err := Get(name)
	if err != nil {
		return err
	}
	return client.Health()
}
