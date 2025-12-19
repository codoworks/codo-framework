package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORSConfig holds CORS middleware configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
	DevMode          bool
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		MaxAge:       86400,
	}
}

// CORS returns a CORS middleware
func CORS(cfg *CORSConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultCORSConfig()
	}

	origins := cfg.AllowOrigins
	if cfg.DevMode || len(origins) == 0 {
		origins = []string{"*"}
	}

	methods := cfg.AllowMethods
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}

	headers := cfg.AllowHeaders
	if len(headers) == 0 {
		headers = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}
	}

	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = 86400
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     methods,
		AllowHeaders:     headers,
		AllowCredentials: cfg.AllowCredentials,
		ExposeHeaders:    cfg.ExposeHeaders,
		MaxAge:           maxAge,
	})
}

// CORSWithOrigins returns a CORS middleware with the specified origins
func CORSWithOrigins(origins ...string) echo.MiddlewareFunc {
	return CORS(&CORSConfig{AllowOrigins: origins})
}
