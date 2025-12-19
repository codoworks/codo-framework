package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRequestIDConfig(t *testing.T) {
	cfg := DefaultRequestIDConfig()

	assert.NotNil(t, cfg.Generator)
	assert.Equal(t, "X-Request-ID", cfg.Header)

	// Test that generator produces UUIDs
	id := cfg.Generator()
	assert.Len(t, id, 36) // UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
}

func TestRequestID(t *testing.T) {
	t.Run("generates new ID", func(t *testing.T) {
		e := echo.New()
		e.Use(RequestID(nil))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
		assert.Len(t, rec.Header().Get("X-Request-ID"), 36)
	})

	t.Run("preserves existing ID", func(t *testing.T) {
		e := echo.New()
		e.Use(RequestID(nil))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "existing-id-123")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "existing-id-123", rec.Header().Get("X-Request-ID"))
	})

	t.Run("custom generator", func(t *testing.T) {
		counter := 0
		e := echo.New()
		e.Use(RequestID(&RequestIDConfig{
			Generator: func() string {
				counter++
				return "custom-id"
			},
		}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "custom-id", rec.Header().Get("X-Request-ID"))
		assert.Equal(t, 1, counter)
	})

	t.Run("custom header name", func(t *testing.T) {
		e := echo.New()
		e.Use(RequestID(&RequestIDConfig{
			Header: "X-Trace-ID",
		}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-Trace-ID"))
	})

	t.Run("empty header uses default", func(t *testing.T) {
		e := echo.New()
		e.Use(RequestID(&RequestIDConfig{
			Header: "",
		}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
	})

	t.Run("nil generator uses default", func(t *testing.T) {
		e := echo.New()
		e.Use(RequestID(&RequestIDConfig{
			Generator: nil,
		}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
	})

	t.Run("sets on both request and response", func(t *testing.T) {
		var requestID string
		e := echo.New()
		e.Use(RequestID(nil))
		e.GET("/test", func(c echo.Context) error {
			requestID = c.Request().Header.Get("X-Request-ID")
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, requestID)
		assert.Equal(t, requestID, rec.Header().Get("X-Request-ID"))
	})
}

func TestDefaultRequestID(t *testing.T) {
	e := echo.New()
	e.Use(DefaultRequestID())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestRequestIDWithGenerator(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDWithGenerator(func() string {
		return "fixed-id"
	}))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "fixed-id", rec.Header().Get("X-Request-ID"))
}
