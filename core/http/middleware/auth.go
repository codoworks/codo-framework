package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/codoworks/codo-framework/core/auth"
	"github.com/codoworks/codo-framework/clients/kratos"
)

// SessionValidator validates sessions
type SessionValidator interface {
	ValidateSession(ctx context.Context, cookie string) (*auth.Identity, error)
	GetCookieName() string
}

// AuthConfig holds authentication middleware configuration
type AuthConfig struct {
	Kratos      SessionValidator
	SkipPaths   []string
	DevMode     bool
	DevIdentity *auth.Identity
}

// Auth returns an authentication middleware
func Auth(cfg *AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check skip paths
			path := c.Request().URL.Path
			for _, skip := range cfg.SkipPaths {
				if strings.HasPrefix(path, skip) {
					return next(c)
				}
			}

			// Dev mode bypass
			if cfg.DevMode && cfg.DevIdentity != nil {
				auth.SetIdentity(c, cfg.DevIdentity)
				return next(c)
			}

			// Get session cookie
			cookie, err := c.Cookie(cfg.Kratos.GetCookieName())
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"code":    "UNAUTHORIZED",
					"message": "No session cookie",
				})
			}

			// Validate session
			identity, err := cfg.Kratos.ValidateSession(c.Request().Context(), cookie.Value)
			if err != nil {
				status := http.StatusUnauthorized
				message := "Invalid session"

				if err == kratos.ErrSessionExpired {
					message = "Session expired"
				}

				return c.JSON(status, map[string]string{
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

// AuthOption is a function that modifies AuthConfig
type AuthOption func(*AuthConfig)

// WithSkipPaths sets paths to skip authentication
func WithSkipPaths(paths ...string) AuthOption {
	return func(cfg *AuthConfig) {
		cfg.SkipPaths = append(cfg.SkipPaths, paths...)
	}
}

// WithDevMode enables dev mode bypass
func WithDevMode(enabled bool, identity *auth.Identity) AuthOption {
	return func(cfg *AuthConfig) {
		cfg.DevMode = enabled
		cfg.DevIdentity = identity
	}
}

// NewAuthConfig creates an AuthConfig with options
func NewAuthConfig(kratos SessionValidator, opts ...AuthOption) *AuthConfig {
	cfg := &AuthConfig{
		Kratos:    kratos,
		SkipPaths: []string{"/health"},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
