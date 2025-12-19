package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()

	assert.Equal(t, 60*time.Second, cfg.Timeout)
}

func TestTimeout(t *testing.T) {
	t.Run("request completes in time", func(t *testing.T) {
		e := echo.New()
		e.Use(Timeout(&TimeoutConfig{Timeout: time.Second}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("request times out", func(t *testing.T) {
		e := echo.New()
		e.Use(Timeout(&TimeoutConfig{Timeout: 50 * time.Millisecond}))
		e.GET("/test", func(c echo.Context) error {
			time.Sleep(200 * time.Millisecond)
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusGatewayTimeout, rec.Code)
		assert.Contains(t, rec.Body.String(), "TIMEOUT")
	})

	t.Run("nil config uses default", func(t *testing.T) {
		e := echo.New()
		e.Use(Timeout(nil))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("zero timeout uses default", func(t *testing.T) {
		e := echo.New()
		e.Use(Timeout(&TimeoutConfig{Timeout: 0}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("negative timeout uses default", func(t *testing.T) {
		e := echo.New()
		e.Use(Timeout(&TimeoutConfig{Timeout: -1}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestTimeoutWithDuration(t *testing.T) {
	e := echo.New()
	e.Use(TimeoutWithDuration(time.Second))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
