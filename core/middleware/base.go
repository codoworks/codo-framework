package middleware

import "github.com/labstack/echo/v4"

// BaseMiddleware provides default implementations for the Middleware interface
// Embed this struct in your middleware to reduce boilerplate
type BaseMiddleware struct {
	name      string
	configKey string
	priority  int
	routers   Router
}

// NewBaseMiddleware creates a new base middleware with the given parameters
func NewBaseMiddleware(name, configKey string, priority int, routers Router) BaseMiddleware {
	return BaseMiddleware{
		name:      name,
		configKey: configKey,
		priority:  priority,
		routers:   routers,
	}
}

// Name returns the unique identifier for this middleware
func (b *BaseMiddleware) Name() string {
	return b.name
}

// ConfigKey returns the configuration section key
func (b *BaseMiddleware) ConfigKey() string {
	return b.configKey
}

// Priority returns the execution order
func (b *BaseMiddleware) Priority() int {
	return b.priority
}

// Routers returns which routers this middleware applies to
func (b *BaseMiddleware) Routers() Router {
	return b.routers
}

// Enabled returns true by default (middleware is enabled unless overridden)
func (b *BaseMiddleware) Enabled(cfg any) bool {
	return true
}

// Configure is a no-op by default (override in your middleware if needed)
func (b *BaseMiddleware) Configure(cfg any) error {
	return nil
}

// Handler must be implemented by the embedding struct - no default implementation
// This ensures each middleware provides its actual functionality
func (b *BaseMiddleware) Handler() echo.MiddlewareFunc {
	panic("Handler() must be implemented by the embedding middleware struct")
}
