package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/clients/kratos"
	codoauth "github.com/codoworks/codo-framework/core/auth"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/errors"
	"github.com/codoworks/codo-framework/core/middleware"
)

func setupAuthMiddleware(t *testing.T, mockKratos *kratos.MockClient, cfg *config.AuthMiddlewareConfig) (*AuthMiddleware, error) {
	// Register mock Kratos client
	clients.MustRegister(mockKratos)
	t.Cleanup(func() {
		clients.ResetRegistry()
	})

	// Create middleware
	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	// Configure
	if err := m.Configure(cfg); err != nil {
		return nil, err
	}

	return m, nil
}

// newEchoWithErrorHandler creates an Echo instance with a custom error handler
// that properly renders framework errors (since auth middleware now returns them)
func newEchoWithErrorHandler() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// Handle framework errors
		if fwkErr, ok := err.(*errors.Error); ok {
			c.JSON(fwkErr.HTTPStatus, map[string]string{
				"code":    fwkErr.Code,
				"message": fwkErr.Message,
			})
			return
		}

		// Fall back to default
		e.DefaultHTTPErrorHandler(err, c)
	}
	return e
}

func TestAuthMiddleware_ValidSession(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	mockKratos.ValidateFunc = func(ctx context.Context, cookie string) (*codoauth.Identity, error) {
		return &codoauth.Identity{
			ID: "user-123",
			Traits: map[string]any{
				"email": "test@example.com",
			},
		}, nil
	}

	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := echo.New()
	e.Use(handler)
	e.GET("/test", func(c echo.Context) error {
		identity, err := codoauth.GetIdentity(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, identity)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "ory_kratos_session", Value: "valid-session"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "user-123")
}

func TestAuthMiddleware_NoSession(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := newEchoWithErrorHandler()
	e.Use(handler)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No cookie
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "No session cookie")
}

func TestAuthMiddleware_InvalidSession(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	mockKratos.ValidateFunc = func(ctx context.Context, cookie string) (*codoauth.Identity, error) {
		return nil, kratos.ErrInvalidSession
	}

	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := newEchoWithErrorHandler()
	e.Use(handler)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "ory_kratos_session", Value: "invalid-session"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid session")
}

func TestAuthMiddleware_ExpiredSession(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	mockKratos.ValidateFunc = func(ctx context.Context, cookie string) (*codoauth.Identity, error) {
		return nil, kratos.ErrSessionExpired
	}

	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := newEchoWithErrorHandler()
	e.Use(handler)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "ory_kratos_session", Value: "expired-session"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Session expired")
}

func TestAuthMiddleware_SkipPaths(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
		SkipPaths: []string{"/public", "/api/public"},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := echo.New()
	e.Use(handler)
	e.GET("/public/resource", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.GET("/api/public/data", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test first skip path
	req := httptest.NewRequest(http.MethodGet, "/public/resource", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test second skip path
	req = httptest.NewRequest(http.MethodGet, "/api/public/data", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_SkipPaths_Health(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
		SkipPaths: []string{"/health"},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := echo.New()
	e.Use(handler)
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.GET("/health/live", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test health endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test health subpath (prefix matching)
	req = httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_DevMode(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
		DevMode:       true,
		DevBypassAuth: true, // Required for bypass
		DevIdentity: &config.DevIdentityConfig{
			ID: "dev-user",
			Traits: map[string]any{
				"email": "dev@example.com",
			},
		},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := echo.New()
	e.Use(handler)
	e.GET("/test", func(c echo.Context) error {
		identity, err := codoauth.GetIdentity(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, identity)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No cookie needed in dev mode
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "dev-user")
}

func TestAuthMiddleware_DevMode_Disabled(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
		DevMode: false,
		DevIdentity: &config.DevIdentityConfig{
			ID: "dev-user",
		},
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := newEchoWithErrorHandler()
	e.Use(handler)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should require auth since dev mode is disabled
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_DevMode_NilIdentity(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
		DevMode:     true,
		DevIdentity: nil, // nil identity
	}

	m, err := setupAuthMiddleware(t, mockKratos, cfg)
	assert.NoError(t, err)

	handler := m.Handler()

	e := newEchoWithErrorHandler()
	e.Use(handler)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should require auth since dev identity is nil
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_Enabled_WithKratos(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	clients.MustRegister(mockKratos)
	defer clients.ResetRegistry()

	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
	}

	enabled := m.Enabled(cfg)
	assert.True(t, enabled, "Auth middleware should be enabled when Kratos is available and config is enabled")
}

func TestAuthMiddleware_Enabled_WithoutKratos(t *testing.T) {
	// Don't register Kratos client

	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
	}

	enabled := m.Enabled(cfg)
	assert.False(t, enabled, "Auth middleware should be disabled when Kratos is not available")
}

func TestAuthMiddleware_Enabled_ConfigDisabled(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	clients.MustRegister(mockKratos)
	defer clients.ResetRegistry()

	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: false, // Explicitly disabled
		},
	}

	enabled := m.Enabled(cfg)
	assert.False(t, enabled, "Auth middleware should be disabled when config is disabled")
}

func TestAuthMiddleware_Configure(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	clients.MustRegister(mockKratos)
	defer clients.ResetRegistry()

	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	cfg := &config.AuthMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{
			Enabled: true,
		},
		SkipPaths: []string{"/public", "/health"},
		DevMode:   true,
		DevIdentity: &config.DevIdentityConfig{
			ID: "dev-123",
			Traits: map[string]any{
				"email": "dev@test.com",
			},
		},
	}

	err := m.Configure(cfg)
	assert.NoError(t, err)

	// Verify configuration was applied
	assert.NotNil(t, m.kratosClient)
	assert.True(t, m.skipPaths["/public"])
	assert.True(t, m.skipPaths["/health"])
	assert.True(t, m.devMode)
	assert.NotNil(t, m.devIdentity)
	assert.Equal(t, "dev-123", m.devIdentity.ID)
}

func TestAuthMiddleware_Configure_NoConfig(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	clients.MustRegister(mockKratos)
	defer clients.ResetRegistry()

	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	// Configure with nil config
	err := m.Configure(nil)
	assert.NoError(t, err)

	// Verify defaults
	assert.NotNil(t, m.kratosClient)
	assert.Empty(t, m.skipPaths)
	assert.False(t, m.devMode)
	assert.Nil(t, m.devIdentity)
}

func TestAuthMiddleware_RouterScope(t *testing.T) {
	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	// Verify auth only applies to protected router
	assert.Equal(t, middleware.RouterProtected, m.Routers())
	assert.True(t, m.Routers().Includes(middleware.RouterProtected))
	assert.False(t, m.Routers().Includes(middleware.RouterPublic))
	assert.False(t, m.Routers().Includes(middleware.RouterHidden))
}

func TestAuthMiddleware_Priority(t *testing.T) {
	m := &AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected,
		),
	}

	// Verify priority is 105
	assert.Equal(t, middleware.PriorityAuth, m.Priority())
	assert.Equal(t, 105, m.Priority())
}
