package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHealthChecker(t *testing.T) {
	ClearHealthCheckers()
	defer ClearHealthCheckers()

	checker := func() error { return nil }
	RegisterHealthChecker(checker)

	assert.Equal(t, 1, HealthCheckersCount())
}

func TestClearHealthCheckers(t *testing.T) {
	RegisterHealthChecker(func() error { return nil })
	assert.True(t, HealthCheckersCount() > 0)

	ClearHealthCheckers()
	assert.Equal(t, 0, HealthCheckersCount())
}

func TestHealthCheckersCount(t *testing.T) {
	ClearHealthCheckers()
	defer ClearHealthCheckers()

	assert.Equal(t, 0, HealthCheckersCount())

	RegisterHealthChecker(func() error { return nil })
	assert.Equal(t, 1, HealthCheckersCount())

	RegisterHealthChecker(func() error { return nil })
	assert.Equal(t, 2, HealthCheckersCount())
}

func TestRegisterHealthRoutes(t *testing.T) {
	e := echo.New()
	RegisterHealthRoutes(e)

	routes := e.Routes()
	paths := make(map[string]bool)
	for _, r := range routes {
		paths[r.Path] = true
	}

	assert.True(t, paths["/health/alive"])
	assert.True(t, paths["/health/ready"])
}

func TestRegisterHealthRoutesOnRouter(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	RegisterHealthRoutesOnRouter(r)

	routes := r.Routes()
	paths := make(map[string]bool)
	for _, route := range routes {
		paths[route.Path] = true
	}

	assert.True(t, paths["/health/alive"])
	assert.True(t, paths["/health/ready"])
}

func TestHandleAlive(t *testing.T) {
	e := echo.New()
	RegisterHealthRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/health/alive", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"alive"`)
}

func TestHandleReady_Healthy(t *testing.T) {
	ClearHealthCheckers()
	defer ClearHealthCheckers()

	RegisterHealthChecker(func() error { return nil })

	e := echo.New()
	RegisterHealthRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ready"`)
}

func TestHandleReady_Unhealthy(t *testing.T) {
	ClearHealthCheckers()
	defer ClearHealthCheckers()

	RegisterHealthChecker(func() error { return errors.New("database connection failed") })

	e := echo.New()
	RegisterHealthRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"not ready"`)
	assert.Contains(t, rec.Body.String(), "database connection failed")
}

func TestHandleReady_NoCheckers(t *testing.T) {
	ClearHealthCheckers()
	defer ClearHealthCheckers()

	e := echo.New()
	RegisterHealthRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ready"`)
}

func TestRegisterNamedHealthChecker(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	RegisterNamedHealthChecker("database", func() error { return nil })

	healthy, results := RunHealthChecks()
	assert.True(t, healthy)
	assert.Equal(t, "ok", results["database"])
}

func TestClearNamedHealthCheckers(t *testing.T) {
	RegisterNamedHealthChecker("test", func() error { return nil })

	ClearNamedHealthCheckers()

	_, results := RunHealthChecks()
	assert.Empty(t, results)
}

func TestRunHealthChecks_AllHealthy(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	RegisterNamedHealthChecker("db", func() error { return nil })
	RegisterNamedHealthChecker("redis", func() error { return nil })

	healthy, results := RunHealthChecks()

	assert.True(t, healthy)
	assert.Equal(t, "ok", results["db"])
	assert.Equal(t, "ok", results["redis"])
}

func TestRunHealthChecks_SomeUnhealthy(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	RegisterNamedHealthChecker("db", func() error { return nil })
	RegisterNamedHealthChecker("redis", func() error { return errors.New("connection refused") })

	healthy, results := RunHealthChecks()

	assert.False(t, healthy)
	assert.Equal(t, "ok", results["db"])
	assert.Equal(t, "connection refused", results["redis"])
}

func TestRunHealthChecks_Empty(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	healthy, results := RunHealthChecks()

	assert.True(t, healthy)
	assert.Empty(t, results)
}

// HealthHandler Tests

func TestHealthHandler_AutoRegistration(t *testing.T) {
	// Verify HealthHandler is in global registry
	handlers := AllHandlers()
	found := false
	for _, h := range handlers {
		if _, ok := h.(*HealthHandler); ok {
			found = true
			break
		}
	}
	assert.True(t, found, "HealthHandler should be auto-registered")
}

func TestHealthHandler_Scope(t *testing.T) {
	h := &HealthHandler{}
	assert.Equal(t, ScopePublic, h.Scope())
}

func TestHealthHandler_Prefix(t *testing.T) {
	h := &HealthHandler{}
	assert.Equal(t, "/health", h.Prefix())
}

func TestHealthHandler_Middlewares(t *testing.T) {
	h := &HealthHandler{}
	assert.Nil(t, h.Middlewares())
}

func TestHealthHandler_Initialize(t *testing.T) {
	h := &HealthHandler{}
	err := h.Initialize()
	assert.NoError(t, err)
}

func TestHealthHandler_Routes(t *testing.T) {
	h := &HealthHandler{}
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	routes := e.Routes()
	pathMethods := make(map[string][]string) // path -> methods
	for _, r := range routes {
		pathMethods[r.Path] = append(pathMethods[r.Path], r.Method)
	}

	// Check all paths exist with both GET and HEAD
	assert.Contains(t, pathMethods["/health"], "GET")
	assert.Contains(t, pathMethods["/health"], "HEAD")

	assert.Contains(t, pathMethods["/health/live"], "GET")
	assert.Contains(t, pathMethods["/health/live"], "HEAD")

	assert.Contains(t, pathMethods["/health/ready"], "GET")
	assert.Contains(t, pathMethods["/health/ready"], "HEAD")
}

func TestHealthHandler_HEADSupport(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	h := &HealthHandler{}
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	// Test HEAD /health/live
	req := httptest.NewRequest(http.MethodHead, "/health/live", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String(), "HEAD should have no body")

	// Test HEAD /health/ready (healthy)
	req = httptest.NewRequest(http.MethodHead, "/health/ready", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String(), "HEAD should have no body")

	// Test HEAD /health/ready (unhealthy)
	RegisterNamedHealthChecker("db", func() error { return errors.New("failed") })
	req = httptest.NewRequest(http.MethodHead, "/health/ready", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Empty(t, rec.Body.String(), "HEAD should have no body")
}

func TestHealthHandler_RootRedirect(t *testing.T) {
	h := &HealthHandler{}
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.handleHealthRoot(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/health/ready", rec.Header().Get("Location"))
}

func TestHealthHandler_LiveEndpoint(t *testing.T) {
	h := &HealthHandler{}
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"alive"`)
}

func TestHealthHandler_ReadyEndpoint_Healthy(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	RegisterNamedHealthChecker("db", func() error { return nil })

	h := &HealthHandler{}
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ready"`)
}

func TestHealthHandler_ReadyEndpoint_Unhealthy(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	RegisterNamedHealthChecker("db", func() error { return errors.New("connection failed") })

	h := &HealthHandler{}
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"not ready"`)
}

func TestHealthHandler_DevModeDetails(t *testing.T) {
	ClearNamedHealthCheckers()
	defer ClearNamedHealthCheckers()

	RegisterNamedHealthChecker("db", func() error { return errors.New("connection timeout") })

	// Test production mode (no details)
	prodConfig := config.NewWithDefaults()
	prodConfig.DevMode = false
	prodConfig.Middleware.Health.Enabled = true
	prodConfig.Middleware.Health.ShowDetailsInProd = false
	SetGlobalConfig(prodConfig)

	h := &HealthHandler{}
	h.Initialize()
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.NotContains(t, rec.Body.String(), "connection timeout", "No details in prod")

	// Test dev mode (with details)
	devConfig := config.NewWithDefaults()
	devConfig.DevMode = true
	devConfig.Middleware.Health.Enabled = true
	SetGlobalConfig(devConfig)

	h2 := &HealthHandler{}
	h2.Initialize()
	e2 := echo.New()
	g2 := e2.Group("/health")
	h2.Routes(g2)

	req = httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec = httptest.NewRecorder()
	e2.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "connection timeout", "Details in dev mode")

	// Clean up
	SetGlobalConfig(nil)
}

func TestHealthHandler_DisabledViaConfig(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Middleware.Health.Enabled = false
	SetGlobalConfig(cfg)

	h := &HealthHandler{}
	h.Initialize()
	e := echo.New()
	g := e.Group("/health")
	h.Routes(g)

	routes := e.Routes()
	assert.Empty(t, routes, "No routes should be registered when disabled")

	// Clean up
	SetGlobalConfig(nil)
}
