package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultGzipConfig(t *testing.T) {
	cfg := DefaultGzipConfig()

	assert.Equal(t, 5, cfg.Level)
}

func TestGzip(t *testing.T) {
	t.Run("compresses response", func(t *testing.T) {
		e := echo.New()
		e.Use(Gzip(nil))
		e.GET("/test", func(c echo.Context) error {
			// Return a large response that would benefit from compression
			return c.String(http.StatusOK, strings.Repeat("hello", 100))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	})

	t.Run("does not compress without accept-encoding", func(t *testing.T) {
		e := echo.New()
		e.Use(Gzip(nil))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, strings.Repeat("hello", 100))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("custom level", func(t *testing.T) {
		e := echo.New()
		e.Use(Gzip(&GzipConfig{Level: 9}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, strings.Repeat("hello", 100))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	})

	t.Run("level below 1 uses default", func(t *testing.T) {
		e := echo.New()
		e.Use(Gzip(&GzipConfig{Level: 0}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, strings.Repeat("hello", 100))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("level above 9 uses default", func(t *testing.T) {
		e := echo.New()
		e.Use(Gzip(&GzipConfig{Level: 10}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, strings.Repeat("hello", 100))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestGzipWithLevel(t *testing.T) {
	e := echo.New()
	e.Use(GzipWithLevel(1))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, strings.Repeat("hello", 100))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}

func TestDefaultGzip(t *testing.T) {
	e := echo.New()
	e.Use(DefaultGzip())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, strings.Repeat("hello", 100))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}
