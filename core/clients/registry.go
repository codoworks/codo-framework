package clients

import (
	"fmt"
	"sync"

	"github.com/codoworks/codo-framework/core/errors"
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
// Displays a beautifully formatted error via the framework's error engine before panicking.
func MustRegister(client Client) {
	if err := Register(client); err != nil {
		frameworkErr := errors.Conflict("Client already registered").
			WithCause(err).
			WithPhase(errors.PhaseClient).
			WithDetail("client_name", client.Name())
		errors.RenderCLI(frameworkErr)
		panic(frameworkErr)
	}
}

// Get retrieves a client by name.
func Get(name string) (Client, error) {
	return getRegistry().Get(name)
}

// MustGet retrieves a client or panics.
// Displays a beautifully formatted error via the framework's error engine before panicking.
func MustGet(name string) Client {
	client, err := Get(name)
	if err != nil {
		frameworkErr := errors.NotFound("Client not found").
			WithCause(err).
			WithPhase(errors.PhaseClient).
			WithDetail("client_name", name)
		errors.RenderCLI(frameworkErr)
		panic(frameworkErr)
	}
	return client
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
// Displays a beautifully formatted error via the framework's error engine before panicking.
func MustGetTyped[T Client](name string) T {
	client, err := GetTyped[T](name)
	if err != nil {
		// Determine if it's a "not found" or "type mismatch" error
		var frameworkErr *errors.Error
		if _, getErr := Get(name); getErr != nil {
			frameworkErr = errors.NotFound("Client not found").
				WithCause(err).
				WithPhase(errors.PhaseClient).
				WithDetail("client_name", name)
		} else {
			frameworkErr = errors.BadRequest("Client type mismatch").
				WithCause(err).
				WithPhase(errors.PhaseClient).
				WithDetail("client_name", name)
		}
		errors.RenderCLI(frameworkErr)
		panic(frameworkErr)
	}
	return client
}

// GetOptionalTyped retrieves an optional client with type assertion.
// Returns (client, true, nil) if found and correct type
// Returns (nil, false, nil) if not found and client is optional
// Returns (nil, false, error) if not found and client is required
// Returns (nil, false, error) if found but wrong type
func GetOptionalTyped[T Client](name string) (T, bool, error) {
	var zero T

	client, err := Get(name)
	if err != nil {
		// Check if client is optional
		if !IsRequired(name) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("required client %q not found", name)
	}

	typed, ok := client.(T)
	if !ok {
		return zero, false, fmt.Errorf("client %q is not of expected type", name)
	}
	return typed, true, nil
}

// MustGetOptionalTyped retrieves an optional typed client.
// Returns (client, true) if found
// Returns (nil, false) if not found and optional
// Panics if not found and required, or if wrong type
// Displays a beautifully formatted error via the framework's error engine before panicking.
func MustGetOptionalTyped[T Client](name string) (T, bool) {
	client, found, err := GetOptionalTyped[T](name)
	if err != nil {
		// Determine if it's a "required but not found" or "type mismatch" error
		var frameworkErr *errors.Error
		if _, getErr := Get(name); getErr != nil {
			frameworkErr = errors.NotFound("Required client not found").
				WithCause(err).
				WithPhase(errors.PhaseClient).
				WithDetail("client_name", name)
		} else {
			frameworkErr = errors.BadRequest("Client type mismatch").
				WithCause(err).
				WithPhase(errors.PhaseClient).
				WithDetail("client_name", name)
		}
		errors.RenderCLI(frameworkErr)
		panic(frameworkErr)
	}
	return client, found
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

// Logger interface for bootstrap logging (subset of logger.Logger)
type Logger interface {
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
}

// InitializeAllWithMetadata initializes all registered clients with enhanced error handling
// based on client metadata. Required clients fail immediately on error. Optional clients
// log warnings and are removed from the registry on failure, allowing the app to continue.
func InitializeAllWithMetadata(configs map[string]any, log Logger) error {
	var optionalErrors []error

	for name, client := range getRegistry().All() {
		cfg := configs[name]

		if err := client.Initialize(cfg); err != nil {
			meta, _ := GetMetadata(name)

			if meta.Requirement == ClientRequired {
				// Required client failed - this is fatal
				if log != nil {
					log.Errorf("Failed to initialize required client %q: %v", name, err)
				}
				return fmt.Errorf("failed to initialize required client %q: %w", name, err)
			} else {
				// Optional client failed - log warning and continue
				if log != nil {
					log.Warnf("Failed to initialize optional client %q: %v", name, err)
				}
				optionalErrors = append(optionalErrors, fmt.Errorf("optional client %q: %w", name, err))

				// Remove failed optional client from registry
				getRegistry().Remove(name)
			}
		} else {
			if log != nil {
				log.Infof("Initialized client: %s", name)
			}
		}
	}

	// Log summary of optional client failures if any
	if len(optionalErrors) > 0 && log != nil {
		log.Warnf("%d optional client(s) failed to initialize but application will continue", len(optionalErrors))
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
