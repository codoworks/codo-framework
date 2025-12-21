package xss

import (
	"fmt"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/labstack/echo/v4"
)

func init() {
	middleware.RegisterMiddleware(&XSSMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"xss",
			"middleware.xss",
			middleware.PrioritySecurityHeaders,
			middleware.RouterAll,
		),
	})
}

// XSSMiddleware handles security headers (XSS protection, frame options, etc.)
type XSSMiddleware struct {
	middleware.BaseMiddleware
	xssProtection      string
	contentTypeNosniff string
	xFrameOptions      string
	hstsMaxAge         int
}

// Enabled checks if XSS middleware is enabled in config
func (m *XSSMiddleware) Enabled(cfg any) bool {
	if cfg == nil {
		return true // Enabled by default
	}

	xssCfg, ok := cfg.(*config.XSSMiddlewareConfig)
	if !ok {
		return true
	}

	return xssCfg.Enabled
}

// Configure initializes the middleware with security header settings from config
func (m *XSSMiddleware) Configure(cfg any) error {
	// Defaults
	m.xssProtection = "1; mode=block"
	m.contentTypeNosniff = "nosniff"
	m.xFrameOptions = "SAMEORIGIN"
	m.hstsMaxAge = 31536000 // 1 year

	// Override with config if provided
	if xssCfg, ok := cfg.(*config.XSSMiddlewareConfig); ok {
		if xssCfg.XSSProtection != "" {
			m.xssProtection = xssCfg.XSSProtection
		}
		if xssCfg.ContentTypeNosniff != "" {
			m.contentTypeNosniff = xssCfg.ContentTypeNosniff
		}
		if xssCfg.XFrameOptions != "" {
			m.xFrameOptions = xssCfg.XFrameOptions
		}
		if xssCfg.HSTSMaxAge > 0 {
			m.hstsMaxAge = xssCfg.HSTSMaxAge
		}
	}

	return nil
}

// Handler returns the XSS protection middleware function
func (m *XSSMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.xssProtection != "" {
				c.Response().Header().Set("X-XSS-Protection", m.xssProtection)
			}
			if m.contentTypeNosniff != "" {
				c.Response().Header().Set("X-Content-Type-Options", m.contentTypeNosniff)
			}
			if m.xFrameOptions != "" {
				c.Response().Header().Set("X-Frame-Options", m.xFrameOptions)
			}
			if m.hstsMaxAge > 0 {
				c.Response().Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d", m.hstsMaxAge))
			}

			return next(c)
		}
	}
}
