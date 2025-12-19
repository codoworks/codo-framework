package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestWrapHandler(t *testing.T) {
	e := echo.New()

	// Test successful handler
	t.Run("success", func(t *testing.T) {
		handler := WrapHandler(func(c *Context) error {
			return c.String(http.StatusOK, "hello")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "hello", rec.Body.String())
	})

	// Test handler that returns error
	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("test error")
		handler := WrapHandler(func(c *Context) error {
			return expectedErr
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.Equal(t, expectedErr, err)
	})
}

func TestWrapMiddleware(t *testing.T) {
	e := echo.New()

	t.Run("middleware executes", func(t *testing.T) {
		executed := false
		middleware := WrapMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				executed = true
				return next(c)
			}
		})

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "hello")
		}

		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := wrapped(c)
		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("middleware can modify response", func(t *testing.T) {
		middleware := WrapMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				c.Response().Header().Set("X-Custom", "value")
				return next(c)
			}
		})

		handler := func(c echo.Context) error {
			return c.String(http.StatusOK, "hello")
		}

		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := wrapped(c)
		assert.NoError(t, err)
		assert.Equal(t, "value", rec.Header().Get("X-Custom"))
	})
}
