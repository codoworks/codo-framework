package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/codoworks/codo-framework/clients/kratos"
	"github.com/codoworks/codo-framework/core/auth"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
)

func init() {
	middleware.RegisterMiddleware(&AuthMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"auth",
			"middleware.auth",
			middleware.PriorityAuth,
			middleware.RouterProtected, // ONLY protected router
		),
	})
}

// SessionValidator defines the interface for session validation
type SessionValidator interface {
	ValidateSession(ctx context.Context, cookie string) (*auth.Identity, error)
	GetCookieName() string
}

// AuthMiddleware implements the Middleware interface for authentication
type AuthMiddleware struct {
	middleware.BaseMiddleware
	kratosClient SessionValidator
	skipPaths    map[string]bool
	devMode      bool
	devIdentity  *auth.Identity
}

// Enabled checks if auth middleware should be enabled
func (m *AuthMiddleware) Enabled(cfg any) bool {
	// Check if Kratos client is available
	if !clients.Has("kratos") {
		return false
	}

	// Check config if provided
	if cfg == nil {
		return true // Enabled by default
	}

	authCfg, ok := cfg.(*config.AuthMiddlewareConfig)
	if !ok {
		return true
	}

	return authCfg.Enabled
}

// Configure initializes the auth middleware with configuration
func (m *AuthMiddleware) Configure(cfg any) error {
	// Get Kratos client from registry
	client, err := clients.Get("kratos")
	if err != nil {
		return fmt.Errorf("failed to get kratos client: %w", err)
	}

	// Type assert to SessionValidator interface
	validator, ok := client.(SessionValidator)
	if !ok {
		return fmt.Errorf("kratos client does not implement SessionValidator interface")
	}
	m.kratosClient = validator

	// Initialize defaults
	m.skipPaths = make(map[string]bool)
	m.devMode = false
	m.devIdentity = nil

	// Parse config if provided
	authCfg, ok := cfg.(*config.AuthMiddlewareConfig)
	if !ok {
		return nil
	}

	// Build skip paths map for fast lookup
	for _, path := range authCfg.SkipPaths {
		m.skipPaths[path] = true
	}

	// Dev mode settings
	m.devMode = authCfg.DevMode
	if authCfg.DevIdentity != nil {
		m.devIdentity = &auth.Identity{
			ID:     authCfg.DevIdentity.ID,
			Traits: authCfg.DevIdentity.Traits,
		}
	}

	return nil
}

// Handler returns the authentication middleware function
func (m *AuthMiddleware) Handler() echo.MiddlewareFunc {
	kratosClient := m.kratosClient
	skipPaths := m.skipPaths
	devMode := m.devMode
	devIdentity := m.devIdentity

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check skip paths (prefix matching)
			path := c.Request().URL.Path
			for skipPath := range skipPaths {
				if strings.HasPrefix(path, skipPath) {
					return next(c)
				}
			}

			// Dev mode bypass
			if devMode && devIdentity != nil {
				auth.SetIdentity(c, devIdentity)
				return next(c)
			}

			// Get session cookie
			cookie, err := c.Cookie(kratosClient.GetCookieName())
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"code":    "UNAUTHORIZED",
					"message": "No session cookie",
				})
			}

			// Validate session
			identity, err := kratosClient.ValidateSession(c.Request().Context(), cookie.Value)
			if err != nil {
				message := "Invalid session"
				if err == kratos.ErrSessionExpired {
					message = "Session expired"
				}

				return c.JSON(http.StatusUnauthorized, map[string]string{
					"code":    "UNAUTHORIZED",
					"message": message,
				})
			}

			// Set identity in context
			auth.SetIdentity(c, identity)

			return next(c)
		}
	}
}
