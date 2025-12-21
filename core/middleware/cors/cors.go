package cors

import (
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func init() {
	middleware.RegisterMiddleware(&CORSMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"cors",
			"middleware.cors",
			middleware.PriorityCORS,
			middleware.RouterAll,
		),
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	middleware.BaseMiddleware
	origins          []string
	methods          []string
	headers          []string
	exposeHeaders    []string
	allowCredentials bool
	maxAge           int
	cfg              *config.Config
}

// Enabled checks if CORS middleware is enabled in config
func (m *CORSMiddleware) Enabled(cfg any) bool {
	if cfg == nil {
		return true // Enabled by default
	}

	corsCfg, ok := cfg.(*config.CORSMiddlewareConfig)
	if !ok {
		return true
	}

	return corsCfg.Enabled
}

// Configure initializes the middleware with config values
func (m *CORSMiddleware) Configure(cfg any) error {
	// Use defaults
	m.origins = []string{"*"}
	m.methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	m.headers = []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}
	m.exposeHeaders = []string{"X-Request-ID"}
	m.maxAge = 86400
	m.allowCredentials = false

	// Override with config if provided
	if corsCfg, ok := cfg.(*config.CORSMiddlewareConfig); ok {
		if len(corsCfg.AllowOrigins) > 0 {
			m.origins = corsCfg.AllowOrigins
		}
		if len(corsCfg.AllowMethods) > 0 {
			m.methods = corsCfg.AllowMethods
		}
		if len(corsCfg.AllowHeaders) > 0 {
			m.headers = corsCfg.AllowHeaders
		}
		if len(corsCfg.ExposeHeaders) > 0 {
			m.exposeHeaders = corsCfg.ExposeHeaders
		}
		if corsCfg.MaxAge > 0 {
			m.maxAge = corsCfg.MaxAge
		}
		m.allowCredentials = corsCfg.AllowCredentials
	}

	return nil
}

// Handler returns the CORS middleware function
func (m *CORSMiddleware) Handler() echo.MiddlewareFunc {
	origins := m.origins
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	// Dev mode: allow all origins
	if m.cfg != nil && m.cfg.IsDevMode() {
		origins = []string{"*"}
	}

	methods := m.methods
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}

	headers := m.headers
	if len(headers) == 0 {
		headers = []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}
	}

	maxAge := m.maxAge
	if maxAge <= 0 {
		maxAge = 86400
	}

	return echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     methods,
		AllowHeaders:     headers,
		AllowCredentials: m.allowCredentials,
		ExposeHeaders:    m.exposeHeaders,
		MaxAge:           maxAge,
	})
}
