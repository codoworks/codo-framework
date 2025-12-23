package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/codoworks/codo-framework/clients/kratos"
	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/core/auth"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/errors"
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

// CachedSession holds a validated session with expiration
type CachedSession struct {
	Identity  *auth.Identity
	ExpiresAt time.Time
}

// AuthMiddleware implements the Middleware interface for authentication
type AuthMiddleware struct {
	middleware.BaseMiddleware
	kratosClient  SessionValidator
	skipPaths     map[string]bool
	devMode       bool // Enables verbose logging
	devBypassAuth bool // Skip real auth when true
	devIdentity   *auth.Identity
	// Session cache
	cache        map[string]*CachedSession
	cacheMutex   sync.RWMutex
	cacheEnabled bool
	cacheTTL     time.Duration
	// Logger for dev mode
	logger *logrus.Logger
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
	m.devBypassAuth = false
	m.devIdentity = nil

	// Initialize cache with defaults
	m.cache = make(map[string]*CachedSession)
	m.cacheEnabled = true
	m.cacheTTL = 15 * time.Minute

	// Get logger for dev mode logging
	if loggerClient, err := clients.GetTyped[*logger.Logger]("logger"); err == nil {
		m.logger = loggerClient.GetLogger()
	}

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
	m.devBypassAuth = authCfg.DevBypassAuth
	if authCfg.DevIdentity != nil {
		m.devIdentity = &auth.Identity{
			ID:     authCfg.DevIdentity.ID,
			Traits: authCfg.DevIdentity.Traits,
		}
	}

	// Cache settings
	m.cacheEnabled = authCfg.CacheEnabled
	if authCfg.CacheTTL > 0 {
		m.cacheTTL = authCfg.CacheTTL
	}

	return nil
}

// Handler returns the authentication middleware function
func (m *AuthMiddleware) Handler() echo.MiddlewareFunc {
	kratosClient := m.kratosClient
	skipPaths := m.skipPaths
	devMode := m.devMode
	devBypassAuth := m.devBypassAuth
	devIdentity := m.devIdentity
	cacheEnabled := m.cacheEnabled
	log := m.logger

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check skip paths (prefix matching)
			path := c.Request().URL.Path
			for skipPath := range skipPaths {
				if strings.HasPrefix(path, skipPath) {
					return next(c)
				}
			}

			// Dev bypass: only if explicitly enabled AND dev_identity configured
			if devBypassAuth && devIdentity != nil {
				if devMode && log != nil {
					log.WithFields(logrus.Fields{
						"user_id":   devIdentity.ID,
						"user_name": devIdentity.Name(),
					}).Info("[Auth] Dev bypass - using static identity")
				}
				auth.SetIdentity(c, devIdentity)
				return next(c)
			}

			// Get session cookie
			cookie, err := c.Cookie(kratosClient.GetCookieName())
			if err != nil {
				return errors.Unauthorized("No session cookie").
					WithPhase(errors.PhaseMiddleware)
			}

			var identity *auth.Identity

			// Check cache first
			if cacheEnabled {
				if cached, found := m.getCachedSession(cookie.Value); found {
					identity = cached
					// Dev mode: log cache hit
					if devMode && log != nil {
						log.WithFields(logrus.Fields{
							"user_id":   identity.ID,
							"user_name": identity.Name(),
							"cache":     "hit",
						}).Info("[Auth] Session validated from cache")
					}
				}
			}

			// Cache miss - validate with Kratos
			if identity == nil {
				identity, err = kratosClient.ValidateSession(c.Request().Context(), cookie.Value)
				if err != nil {
					message := "Invalid session"
					if err == kratos.ErrSessionExpired {
						message = "Session expired"
					}

					return errors.Unauthorized(message).
						WithPhase(errors.PhaseMiddleware)
				}

				// Store in cache if enabled
				if cacheEnabled {
					m.setCachedSession(cookie.Value, identity)
				}

				// Dev mode: always log user info
				if devMode && log != nil {
					fields := logrus.Fields{
						"user_id":   identity.ID,
						"user_name": identity.Name(),
					}
					if cacheEnabled {
						fields["cache"] = "miss"
						log.WithFields(fields).Info("[Auth] Session validated from Kratos")
					} else {
						log.WithFields(fields).Info("[Auth] Session validated")
					}
				}
			}

			// Set identity in context
			auth.SetIdentity(c, identity)

			return next(c)
		}
	}
}

// getCachedSession retrieves a session from cache if valid (with lazy cleanup)
func (m *AuthMiddleware) getCachedSession(cookie string) (*auth.Identity, bool) {
	m.cacheMutex.RLock()
	cached, exists := m.cache[cookie]
	m.cacheMutex.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(cached.ExpiresAt) {
		// Lazy cleanup: remove expired entry
		m.cacheMutex.Lock()
		delete(m.cache, cookie)
		m.cacheMutex.Unlock()
		return nil, false
	}

	return cached.Identity, true
}

// setCachedSession stores a session in cache
func (m *AuthMiddleware) setCachedSession(cookie string, identity *auth.Identity) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.cache[cookie] = &CachedSession{
		Identity:  identity,
		ExpiresAt: time.Now().Add(m.cacheTTL),
	}
}
