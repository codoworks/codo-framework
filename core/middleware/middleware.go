package middleware

import "github.com/labstack/echo/v4"

// Middleware defines the interface that all middleware must implement
type Middleware interface {
	// Name returns the unique identifier for this middleware
	Name() string

	// ConfigKey returns the configuration section key (e.g., "middleware.logger")
	// Return empty string if no configuration is needed
	ConfigKey() string

	// Priority returns the execution order (lower runs first)
	Priority() int

	// Routers returns which routers this middleware applies to (bitmask)
	Routers() Router

	// Enabled checks if this middleware should be applied
	// Receives the config section for this middleware (if ConfigKey is non-empty)
	Enabled(cfg any) bool

	// Configure initializes the middleware with its configuration
	// Receives the config section for this middleware (if ConfigKey is non-empty)
	Configure(cfg any) error

	// Handler returns the actual middleware function
	Handler() echo.MiddlewareFunc
}

// Router defines which routers a middleware applies to using a bitmask
type Router uint8

const (
	RouterPublic    Router = 1 << iota // 1 - Port 8081, no auth
	RouterProtected                     // 2 - Port 8080, requires auth
	RouterHidden                        // 4 - Port 8079, admin only

	RouterAll = RouterPublic | RouterProtected | RouterHidden // 7 - All routers
)

// Includes checks if this router scope includes the target router
func (r Router) Includes(target Router) bool {
	return r&target != 0
}

// String returns a human-readable representation of the router scope
func (r Router) String() string {
	switch r {
	case RouterPublic:
		return "public"
	case RouterProtected:
		return "protected"
	case RouterHidden:
		return "hidden"
	case RouterAll:
		return "all"
	default:
		// Multiple routers
		var parts []string
		if r.Includes(RouterPublic) {
			parts = append(parts, "public")
		}
		if r.Includes(RouterProtected) {
			parts = append(parts, "protected")
		}
		if r.Includes(RouterHidden) {
			parts = append(parts, "hidden")
		}
		if len(parts) == 0 {
			return "none"
		}
		result := parts[0]
		for i := 1; i < len(parts); i++ {
			result += "," + parts[i]
		}
		return result
	}
}

// Priority constants define the execution order of middleware
const (
	// Core middleware (0-99): Essential, non-removable
	PriorityCoreMin     = 0
	PriorityCoreMax     = 99
	PriorityRecover     = 0  // Must be first - catches panics
	PriorityRequestID   = 10 // Generate/propagate X-Request-ID
	PriorityContextInit = 20 // Wrap echo.Context in *Context (if needed)

	// Feature middleware (100-199): Built-in, configurable
	PriorityFeatureMin      = 100
	PriorityFeatureMax      = 199
	PriorityLogger          = 100 // Request/response logging
	PriorityAuth            = 105 // Authentication (Kratos session validation)
	PriorityTimeout         = 110 // Request timeout
	PriorityCORS            = 120 // Cross-origin handling
	PriorityRateLimit       = 130 // Rate limiting per IP
	PrioritySecurityHeaders = 140 // XSS, HSTS, etc.
	PriorityCompression     = 150 // Gzip responses

	// Consumer middleware (200+): App-specific middlewares
	PriorityConsumerMin = 200
	PriorityConsumerMax = 299
)
