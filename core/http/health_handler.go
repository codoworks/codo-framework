package http

import (
	"net/http"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/labstack/echo/v4"
)

func init() {
	// Auto-register health handler on framework import
	RegisterHandler(&HealthHandler{})
}

// HealthHandler provides health check endpoints following Kubernetes standards
type HealthHandler struct {
	cfg *config.Config
}

// Prefix returns the URL prefix for health routes
func (h *HealthHandler) Prefix() string {
	return "/health"
}

// Scope returns the router scope (Public only - no auth required)
func (h *HealthHandler) Scope() RouterScope {
	return ScopePublic
}

// Middlewares returns handler-specific middlewares (none needed)
func (h *HealthHandler) Middlewares() []echo.MiddlewareFunc {
	return nil
}

// Initialize performs any required initialization
func (h *HealthHandler) Initialize() error {
	// Get config from global accessor
	h.cfg = GetGlobalConfig()
	return nil
}

// Routes registers health check routes
func (h *HealthHandler) Routes(g *echo.Group) {
	// Check if health endpoints are disabled via config
	if h.cfg != nil && !h.cfg.Middleware.Health.Enabled {
		return // Don't register any routes
	}

	// Root /health - redirect to /health/ready
	g.GET("", h.handleHealthRoot)
	g.HEAD("", h.handleHealthRoot)

	// /health/live - liveness probe (always returns 200)
	g.GET("/live", h.handleLive)
	g.HEAD("/live", h.handleLive)

	// /health/ready - readiness probe (checks dependencies)
	g.GET("/ready", h.handleReady)
	g.HEAD("/ready", h.handleReady)
}

// handleHealthRoot redirects to /health/ready
func (h *HealthHandler) handleHealthRoot(c echo.Context) error {
	// For HEAD requests, just return 200
	if c.Request().Method == http.MethodHead {
		return c.NoContent(http.StatusOK)
	}

	// For GET, redirect to /health/ready
	return c.Redirect(http.StatusMovedPermanently, "/health/ready")
}

// handleLive returns liveness status (always healthy)
func (h *HealthHandler) handleLive(c echo.Context) error {
	if c.Request().Method == http.MethodHead {
		return c.NoContent(http.StatusOK)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// handleReady returns readiness status (checks dependencies)
func (h *HealthHandler) handleReady(c echo.Context) error {
	// Run all named health checkers
	healthy, results := RunHealthChecks()

	// Determine if we should show details
	showDetails := h.shouldShowDetails()

	// Build response
	response := HealthResponse{
		Status: "ready",
	}

	if !healthy {
		response.Status = "not ready"

		// Only include details if configured
		if showDetails {
			response.Details = results
		}
	}

	// For HEAD requests, just return status code
	if c.Request().Method == http.MethodHead {
		if healthy {
			return c.NoContent(http.StatusOK)
		}
		return c.NoContent(http.StatusServiceUnavailable)
	}

	// For GET, return JSON
	statusCode := http.StatusOK
	if !healthy {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, response)
}

// shouldShowDetails determines if health check details should be shown
func (h *HealthHandler) shouldShowDetails() bool {
	if h.cfg == nil {
		return false
	}

	// Show details if in dev mode
	if h.cfg.IsDevMode() {
		return true
	}

	// Check config override for production
	return h.cfg.Middleware.Health.ShowDetailsInProd
}
