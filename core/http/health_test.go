package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
