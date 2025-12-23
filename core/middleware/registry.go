package middleware

import (
	"os"
	"sync"

	"github.com/codoworks/codo-framework/core/errors"
)

// Registry holds all registered middleware
type Registry struct {
	middlewares map[string]Middleware
	mu          sync.RWMutex
}

// Global registry for middleware auto-registration via init()
var globalRegistry = &Registry{
	middlewares: make(map[string]Middleware),
}

// RegisterMiddleware registers a middleware in the global registry
// This is typically called from init() functions in middleware packages
// Exits with error if a middleware with the same name is already registered
func RegisterMiddleware(m Middleware) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	name := m.Name()
	if _, exists := globalRegistry.middlewares[name]; exists {
		frameworkErr := errors.Conflict("Middleware already registered").
			WithPhase(errors.PhaseMiddleware).
			WithDetail("middleware_name", name)
		errors.RenderCLI(frameworkErr)
		os.Exit(1)
	}

	globalRegistry.middlewares[name] = m
}

// GetGlobalRegistry returns the global middleware registry
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// All returns all registered middleware
func (r *Registry) All() []Middleware {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Middleware, 0, len(r.middlewares))
	for _, m := range r.middlewares {
		result = append(result, m)
	}
	return result
}

// Get returns a middleware by name if it exists
func (r *Registry) Get(name string) (Middleware, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.middlewares[name]
	return m, exists
}

// Count returns the number of registered middleware
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.middlewares)
}

// Names returns all registered middleware names
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.middlewares))
	for name := range r.middlewares {
		names = append(names, name)
	}
	return names
}
