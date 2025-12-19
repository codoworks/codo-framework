package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLoggerConfig(t *testing.T) {
	cfg := DefaultLoggerConfig()

	assert.NotNil(t, cfg.Logger)
	assert.Nil(t, cfg.SkipPaths)
	assert.False(t, cfg.DevMode)
}

func TestLogger(t *testing.T) {
	t.Run("logs successful request", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Logger(&LoggerConfig{Logger: log}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, buf.String(), "Request completed")
		assert.Contains(t, buf.String(), "/test")
	})

	t.Run("logs failed request", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Logger(&LoggerConfig{Logger: log}))
		e.GET("/test", func(c echo.Context) error {
			return errors.New("test error")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Contains(t, buf.String(), "Request failed")
		assert.Contains(t, buf.String(), "test error")
	})

	t.Run("logs 5xx as error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Logger(&LoggerConfig{Logger: log}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusInternalServerError, "error")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Contains(t, buf.String(), "server error")
	})

	t.Run("logs 4xx as warn", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Logger(&LoggerConfig{Logger: log}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusBadRequest, "bad request")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Contains(t, buf.String(), "client error")
	})

	t.Run("skips configured paths", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Logger(&LoggerConfig{
			Logger:    log,
			SkipPaths: []string{"/health"},
		}))
		e.GET("/health", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, buf.String())
	})

	t.Run("dev mode includes extra fields", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Logger(&LoggerConfig{
			Logger:  log,
			DevMode: true,
		}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
		req.Header.Set("User-Agent", "TestAgent")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Contains(t, buf.String(), "foo=bar")
		assert.Contains(t, buf.String(), "TestAgent")
	})

	t.Run("includes request ID if present", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Logger(&LoggerConfig{Logger: log}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "test-request-id")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Contains(t, buf.String(), "test-request-id")
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		e := echo.New()
		e.Use(Logger(nil))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("nil logger in config uses standard", func(t *testing.T) {
		e := echo.New()
		e.Use(Logger(&LoggerConfig{Logger: nil}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
