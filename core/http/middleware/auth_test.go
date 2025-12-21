package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/auth"
	"github.com/codoworks/codo-framework/clients/kratos"
)

func TestAuthMiddleware(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	mockKratos.ValidateFunc = func(ctx context.Context, cookie string) (*auth.Identity, error) {
		return &auth.Identity{
			ID: "user-123",
			Traits: map[string]any{
				"email": "test@example.com",
			},
		}, nil
	}

	cfg := NewAuthConfig(mockKratos)
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
	e.GET("/test", func(c echo.Context) error {
		identity, err := auth.GetIdentity(c)
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
	cfg := NewAuthConfig(mockKratos)
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
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
	mockKratos.ValidateFunc = func(ctx context.Context, cookie string) (*auth.Identity, error) {
		return nil, kratos.ErrInvalidSession
	}

	cfg := NewAuthConfig(mockKratos)
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
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
	mockKratos.ValidateFunc = func(ctx context.Context, cookie string) (*auth.Identity, error) {
		return nil, kratos.ErrSessionExpired
	}

	cfg := NewAuthConfig(mockKratos)
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
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
	cfg := NewAuthConfig(mockKratos, WithSkipPaths("/public", "/api/public"))
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
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
	cfg := NewAuthConfig(mockKratos) // /health is included by default
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
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

	// Test health subpath
	req = httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_DevMode(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	devIdentity := &auth.Identity{
		ID: "dev-user",
		Traits: map[string]any{
			"email": "dev@example.com",
		},
	}
	cfg := NewAuthConfig(mockKratos, WithDevMode(true, devIdentity))
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
	e.GET("/test", func(c echo.Context) error {
		identity, err := auth.GetIdentity(c)
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
	devIdentity := &auth.Identity{ID: "dev-user"}
	cfg := NewAuthConfig(mockKratos, WithDevMode(false, devIdentity))
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
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
	cfg := NewAuthConfig(mockKratos, WithDevMode(true, nil))
	middleware := Auth(cfg)

	e := echo.New()
	e.Use(middleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should require auth since dev identity is nil
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWithSkipPaths(t *testing.T) {
	cfg := &AuthConfig{
		SkipPaths: []string{"/existing"},
	}

	opt := WithSkipPaths("/new1", "/new2")
	opt(cfg)

	assert.Contains(t, cfg.SkipPaths, "/existing")
	assert.Contains(t, cfg.SkipPaths, "/new1")
	assert.Contains(t, cfg.SkipPaths, "/new2")
}

func TestWithDevMode(t *testing.T) {
	cfg := &AuthConfig{}
	identity := &auth.Identity{ID: "test-user"}

	opt := WithDevMode(true, identity)
	opt(cfg)

	assert.True(t, cfg.DevMode)
	assert.Equal(t, identity, cfg.DevIdentity)
}

func TestNewAuthConfig(t *testing.T) {
	mockKratos := kratos.NewMockClient()

	cfg := NewAuthConfig(mockKratos)

	assert.Equal(t, mockKratos, cfg.Kratos)
	assert.Contains(t, cfg.SkipPaths, "/health")
	assert.False(t, cfg.DevMode)
	assert.Nil(t, cfg.DevIdentity)
}

func TestNewAuthConfig_WithOptions(t *testing.T) {
	mockKratos := kratos.NewMockClient()
	devIdentity := &auth.Identity{ID: "dev"}

	cfg := NewAuthConfig(
		mockKratos,
		WithSkipPaths("/public"),
		WithDevMode(true, devIdentity),
	)

	assert.Equal(t, mockKratos, cfg.Kratos)
	assert.Contains(t, cfg.SkipPaths, "/health")
	assert.Contains(t, cfg.SkipPaths, "/public")
	assert.True(t, cfg.DevMode)
	assert.Equal(t, devIdentity, cfg.DevIdentity)
}
