package errorhandler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/errors"
	httpPkg "github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/core/middleware"
)

func TestErrorHandlerMiddleware_Properties(t *testing.T) {
	// Use the middleware as registered (with BaseMiddleware initialized)
	m := &ErrorHandlerMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"errorhandler",
			"middleware.errorhandler",
			middleware.PriorityErrorHandler,
			middleware.RouterAll,
		),
	}

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "errorhandler", m.Name())
	})

	t.Run("always enabled", func(t *testing.T) {
		assert.True(t, m.Enabled(nil))
	})

	t.Run("priority", func(t *testing.T) {
		assert.Equal(t, middleware.PriorityErrorHandler, m.Priority())
	})

	t.Run("routers", func(t *testing.T) {
		assert.Equal(t, middleware.RouterAll, m.Routers())
	})
}

func TestErrorHandlerMiddleware_Handler(t *testing.T) {
	m := &ErrorHandlerMiddleware{}
	handler := m.Handler()

	t.Run("no error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("framework error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler(func(c echo.Context) error {
			return errors.NotFound("Resource not found")
		})

		err := h(c)
		assert.NoError(t, err) // Error is handled, not propagated
		assert.Equal(t, http.StatusNotFound, rec.Code)

		var resp httpPkg.Response
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, errors.CodeNotFound, resp.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler(func(c echo.Context) error {
			return &httpPkg.BindError{Cause: assert.AnError}
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp httpPkg.Response
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, errors.CodeBadRequest, resp.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler(func(c echo.Context) error {
			return &httpPkg.ValidationErrorList{
				Errors: []httpPkg.ValidationError{
					{Field: "email", Message: "required"},
				},
			}
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

		var resp httpPkg.Response
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, errors.CodeValidation, resp.Code)
		assert.Len(t, resp.Errors, 1)
	})

	t.Run("unknown error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handler(func(c echo.Context) error {
			return assert.AnError
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var resp httpPkg.Response
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, errors.CodeInternal, resp.Code)
	})
}

func TestErrorHandlerMiddleware_RequestContextEnrichment(t *testing.T) {
	m := &ErrorHandlerMiddleware{}
	handler := m.Handler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	req.Header.Set(echo.HeaderXRequestID, "test-request-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// We need to capture the framework error to check enrichment
	// But ErrorHandlerMiddleware returns the response, so we verify via logging
	// For now, just verify the middleware runs without panic

	h := handler(func(c echo.Context) error {
		return errors.BadRequest("test error")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
