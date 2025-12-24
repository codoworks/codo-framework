package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")

	assert.NotNil(t, r)
	assert.Equal(t, ScopePublic, r.Scope())
	assert.Equal(t, ":8080", r.Addr())
	assert.NotNil(t, r.Echo())
	assert.NotNil(t, r.Echo().Validator, "Validator should be auto-set")
}

func TestRouter_Scope(t *testing.T) {
	tests := []struct {
		scope RouterScope
	}{
		{ScopePublic},
		{ScopeProtected},
		{ScopeHidden},
	}

	for _, tt := range tests {
		t.Run(tt.scope.String(), func(t *testing.T) {
			r := NewRouter(tt.scope, ":8080")
			assert.Equal(t, tt.scope, r.Scope())
		})
	}
}

func TestRouter_Addr(t *testing.T) {
	r := NewRouter(ScopePublic, ":9999")
	assert.Equal(t, ":9999", r.Addr())
}

func TestRouter_Echo(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	e := r.Echo()
	assert.NotNil(t, e)
	assert.IsType(t, &echo.Echo{}, e)
}

func TestRouter_Use(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")

	called := false
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			called = true
			return next(c)
		}
	})

	r.Echo().GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.True(t, called)
}

func TestRouter_Group(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	g := r.Group("/api")

	assert.NotNil(t, g)
}

func TestRouter_HTTPMethods(t *testing.T) {
	tests := []struct {
		method     string
		register   func(r *Router, path string, h HandlerFunc)
	}{
		{"GET", func(r *Router, path string, h HandlerFunc) { r.GET(path, h) }},
		{"POST", func(r *Router, path string, h HandlerFunc) { r.POST(path, h) }},
		{"PUT", func(r *Router, path string, h HandlerFunc) { r.PUT(path, h) }},
		{"PATCH", func(r *Router, path string, h HandlerFunc) { r.PATCH(path, h) }},
		{"DELETE", func(r *Router, path string, h HandlerFunc) { r.DELETE(path, h) }},
		{"OPTIONS", func(r *Router, path string, h HandlerFunc) { r.OPTIONS(path, h) }},
		{"HEAD", func(r *Router, path string, h HandlerFunc) { r.HEAD(path, h) }},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			r := NewRouter(ScopePublic, ":8080")
			tt.register(r, "/test", func(c *Context) error {
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tt.method == "HEAD" {
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				assert.Equal(t, http.StatusOK, rec.Code)
			}
		})
	}
}

func TestRouter_Any(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	routes := r.Any("/any", func(c *Context) error {
		return c.String(http.StatusOK, "ok")
	})

	assert.NotEmpty(t, routes)

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/any", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRouter_Static(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	route := r.Static("/static", ".")
	assert.NotNil(t, route)
}

func TestRouter_File(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	route := r.File("/test", "router_test.go")
	assert.NotNil(t, route)
}

func TestRouter_Routes(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	r.GET("/test1", func(c *Context) error { return nil })
	r.POST("/test2", func(c *Context) error { return nil })

	routes := r.Routes()
	assert.NotEmpty(t, routes)
}

func TestRouter_RegisterHandlers(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	t.Run("success", func(t *testing.T) {
		ClearHandlers()
		h := &mockHandler{
			prefix: "/test",
			scope:  ScopePublic,
			routes: func(g *echo.Group) {
				g.GET("", func(c echo.Context) error {
					return c.String(http.StatusOK, "ok")
				})
			},
		}
		RegisterHandler(h)

		r := NewRouter(ScopePublic, ":8080")
		err := r.RegisterHandlers()

		assert.NoError(t, err)
		assert.True(t, h.initialized)
	})

	t.Run("initialize error", func(t *testing.T) {
		ClearHandlers()
		h := &mockHandler{
			prefix:  "/test",
			scope:   ScopePublic,
			initErr: errors.New("init failed"),
		}
		RegisterHandler(h)

		r := NewRouter(ScopePublic, ":8080")
		err := r.RegisterHandlers()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "init failed")
	})
}

func TestRouter_SetValidator(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")

	validator := &testValidator{}
	r.SetValidator(validator)

	assert.Equal(t, validator, r.Echo().Validator)
}

type testValidator struct{}

func (v *testValidator) Validate(i interface{}) error {
	return nil
}

func TestRouter_SetErrorHandler(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")

	called := false
	r.SetErrorHandler(func(err error, c echo.Context) {
		called = true
	})

	// Trigger error handler by returning an error from a handler
	r.Echo().GET("/test", func(c echo.Context) error {
		return errors.New("test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.True(t, called)
}

func TestRouter_ServeHTTP(t *testing.T) {
	r := NewRouter(ScopePublic, ":8080")
	r.GET("/test", func(c *Context) error {
		return c.String(http.StatusOK, "hello")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())
}

func TestRouter_Shutdown(t *testing.T) {
	r := NewRouter(ScopePublic, ":0")

	// Shutdown without starting should be OK
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := r.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestRouter_Start(t *testing.T) {
	r := NewRouter(ScopePublic, "127.0.0.1:0")
	r.GET("/health", func(c *Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- r.Start()
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := r.Shutdown(ctx)
	assert.NoError(t, err)

	// Wait for Start to return
	select {
	case err := <-errCh:
		assert.NoError(t, err) // Should be nil since we closed gracefully
	case <-time.After(time.Second):
		t.Fatal("Start did not return after Shutdown")
	}
}

func TestRouter_StartTLS(t *testing.T) {
	r := NewRouter(ScopePublic, "127.0.0.1:0")

	// Start server with invalid cert files - will fail immediately
	errCh := make(chan error, 1)
	go func() {
		errCh <- r.StartTLS("nonexistent.crt", "nonexistent.key")
	}()

	// Wait for Start to return with error
	select {
	case err := <-errCh:
		assert.Error(t, err) // Should error due to missing cert files
	case <-time.After(time.Second):
		t.Fatal("StartTLS did not return")
	}
}

func TestRouter_Start_Error(t *testing.T) {
	// Try to start on an invalid address to trigger error path
	r := NewRouter(ScopePublic, "invalid-address-that-will-fail:99999999")

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.Start()
	}()

	select {
	case err := <-errCh:
		assert.Error(t, err) // Should error due to invalid address
	case <-time.After(time.Second):
		t.Fatal("Start did not return")
	}
}
