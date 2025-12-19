package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRecoverConfig(t *testing.T) {
	cfg := DefaultRecoverConfig()

	assert.NotNil(t, cfg.Logger)
	assert.False(t, cfg.DevMode)
}

func TestRecover(t *testing.T) {
	t.Run("no panic passes through", func(t *testing.T) {
		e := echo.New()
		e.Use(Recover(nil))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("recovers from panic", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Recover(&RecoverConfig{Logger: log}))
		e.GET("/test", func(c echo.Context) error {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "INTERNAL_ERROR")
		assert.Contains(t, buf.String(), "Panic recovered")
		assert.Contains(t, buf.String(), "test panic")
	})

	t.Run("recovers from error panic", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Recover(&RecoverConfig{Logger: log}))
		e.GET("/test", func(c echo.Context) error {
			panic(echo.ErrBadRequest)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, buf.String(), "Panic recovered")
	})

	t.Run("dev mode includes stack trace in response", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Recover(&RecoverConfig{Logger: log, DevMode: true}))
		e.GET("/test", func(c echo.Context) error {
			panic("dev panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "dev panic")
		assert.Contains(t, rec.Body.String(), "stack")
	})

	t.Run("production mode hides details", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)

		e := echo.New()
		e.Use(Recover(&RecoverConfig{Logger: log, DevMode: false}))
		e.GET("/test", func(c echo.Context) error {
			panic("secret panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.NotContains(t, rec.Body.String(), "secret panic")
		assert.NotContains(t, rec.Body.String(), "stack")
	})

	t.Run("nil logger uses standard", func(t *testing.T) {
		e := echo.New()
		e.Use(Recover(&RecoverConfig{Logger: nil}))
		e.GET("/test", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestDefaultRecover(t *testing.T) {
	e := echo.New()
	e.Use(DefaultRecover())
	e.GET("/test", func(c echo.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRecoverWithLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	log := logrus.New()
	log.SetOutput(buf)

	e := echo.New()
	e.Use(RecoverWithLogger(log))
	e.GET("/test", func(c echo.Context) error {
		panic("logged panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, buf.String(), "logged panic")
}
