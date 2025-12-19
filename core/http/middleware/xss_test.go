package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultXSSConfig(t *testing.T) {
	cfg := DefaultXSSConfig()

	assert.Equal(t, "1; mode=block", cfg.XSSProtection)
	assert.Equal(t, "nosniff", cfg.ContentTypeNosniff)
	assert.Equal(t, "SAMEORIGIN", cfg.XFrameOptions)
	assert.Equal(t, "strict-origin-when-cross-origin", cfg.ReferrerPolicy)
}

func TestXSS(t *testing.T) {
	t.Run("sets default headers", func(t *testing.T) {
		e := echo.New()
		e.Use(XSS(nil))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "SAMEORIGIN", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	})

	t.Run("custom config", func(t *testing.T) {
		e := echo.New()
		e.Use(XSS(&XSSConfig{
			XSSProtection:         "0",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "DENY",
			ContentSecurityPolicy: "default-src 'self'",
			ReferrerPolicy:        "no-referrer",
		}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "0", rec.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"))
		assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
	})

	t.Run("empty values not set", func(t *testing.T) {
		e := echo.New()
		e.Use(XSS(&XSSConfig{
			XSSProtection:      "",
			ContentTypeNosniff: "",
			XFrameOptions:      "",
			ReferrerPolicy:     "",
		}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("X-XSS-Protection"))
		assert.Empty(t, rec.Header().Get("X-Content-Type-Options"))
		assert.Empty(t, rec.Header().Get("X-Frame-Options"))
		assert.Empty(t, rec.Header().Get("Referrer-Policy"))
	})
}

func TestDefaultXSS(t *testing.T) {
	e := echo.New()
	e.Use(DefaultXSS())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-XSS-Protection"))
}

func TestSecureHeaders(t *testing.T) {
	e := echo.New()
	e.Use(SecureHeaders())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
}
