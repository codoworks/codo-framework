package middleware

import (
	"github.com/labstack/echo/v4"
)

// XSSConfig holds XSS protection middleware configuration
type XSSConfig struct {
	XSSProtection         string
	ContentTypeNosniff    string
	XFrameOptions         string
	ContentSecurityPolicy string
	ReferrerPolicy        string
}

// DefaultXSSConfig returns a default XSS protection configuration
func DefaultXSSConfig() *XSSConfig {
	return &XSSConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// XSS returns a middleware that sets XSS protection headers
func XSS(cfg *XSSConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultXSSConfig()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.XSSProtection != "" {
				c.Response().Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}
			if cfg.ContentTypeNosniff != "" {
				c.Response().Header().Set("X-Content-Type-Options", cfg.ContentTypeNosniff)
			}
			if cfg.XFrameOptions != "" {
				c.Response().Header().Set("X-Frame-Options", cfg.XFrameOptions)
			}
			if cfg.ContentSecurityPolicy != "" {
				c.Response().Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}
			if cfg.ReferrerPolicy != "" {
				c.Response().Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			return next(c)
		}
	}
}

// DefaultXSS returns an XSS middleware with default settings
func DefaultXSS() echo.MiddlewareFunc {
	return XSS(nil)
}

// SecureHeaders returns a middleware with all secure headers configured
func SecureHeaders() echo.MiddlewareFunc {
	return XSS(&XSSConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	})
}
