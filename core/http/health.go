package http

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// HealthChecker is a function that checks health of a component
type HealthChecker func() error

var (
	healthCheckers   []HealthChecker
	healthCheckersMu sync.RWMutex
)

// RegisterHealthChecker adds a health checker to the registry
func RegisterHealthChecker(checker HealthChecker) {
	healthCheckersMu.Lock()
	defer healthCheckersMu.Unlock()
	healthCheckers = append(healthCheckers, checker)
}

// ClearHealthCheckers removes all health checkers (for testing)
func ClearHealthCheckers() {
	healthCheckersMu.Lock()
	defer healthCheckersMu.Unlock()
	healthCheckers = nil
}

// HealthCheckersCount returns the number of registered health checkers
func HealthCheckersCount() int {
	healthCheckersMu.RLock()
	defer healthCheckersMu.RUnlock()
	return len(healthCheckers)
}

// RegisterHealthRoutes registers health check routes on an Echo instance.
//
// Deprecated: Health routes now auto-register via HealthHandler.
// This function is kept for backwards compatibility but will be removed in v4.0.
// Remove any manual calls to RegisterHealthRoutes() from your code.
func RegisterHealthRoutes(e *echo.Echo) {
	g := e.Group("/health")
	g.GET("/alive", handleAlive)
	g.GET("/ready", handleReady)
}

// RegisterHealthRoutesOnRouter registers health check routes on a Router.
//
// Deprecated: Health routes now auto-register via HealthHandler.
// This function is kept for backwards compatibility but will be removed in v4.0.
// Remove any manual calls to RegisterHealthRoutesOnRouter() from your code.
func RegisterHealthRoutesOnRouter(r *Router) {
	g := r.Echo().Group("/health")
	g.GET("/alive", handleAlive)
	g.GET("/ready", handleReady)
}

func handleAlive(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

func handleReady(c echo.Context) error {
	healthCheckersMu.RLock()
	checkers := make([]HealthChecker, len(healthCheckers))
	copy(checkers, healthCheckers)
	healthCheckersMu.RUnlock()

	for _, checker := range checkers {
		if err := checker(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "not ready",
				"error":  err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

// NamedHealthChecker is a health checker with a name
type NamedHealthChecker struct {
	Name    string
	Checker HealthChecker
}

var (
	namedHealthCheckers   []NamedHealthChecker
	namedHealthCheckersMu sync.RWMutex
)

// RegisterNamedHealthChecker adds a named health checker
func RegisterNamedHealthChecker(name string, checker HealthChecker) {
	namedHealthCheckersMu.Lock()
	defer namedHealthCheckersMu.Unlock()
	namedHealthCheckers = append(namedHealthCheckers, NamedHealthChecker{
		Name:    name,
		Checker: checker,
	})
}

// ClearNamedHealthCheckers removes all named health checkers (for testing)
func ClearNamedHealthCheckers() {
	namedHealthCheckersMu.Lock()
	defer namedHealthCheckersMu.Unlock()
	namedHealthCheckers = nil
}

// RunHealthChecks runs all named health checkers and returns results
func RunHealthChecks() (bool, map[string]string) {
	namedHealthCheckersMu.RLock()
	checkers := make([]NamedHealthChecker, len(namedHealthCheckers))
	copy(checkers, namedHealthCheckers)
	namedHealthCheckersMu.RUnlock()

	results := make(map[string]string)
	healthy := true

	for _, nc := range checkers {
		if err := nc.Checker(); err != nil {
			results[nc.Name] = err.Error()
			healthy = false
		} else {
			results[nc.Name] = "ok"
		}
	}

	return healthy, results
}
