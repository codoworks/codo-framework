package gzip

import (
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func init() {
	middleware.RegisterMiddleware(&GzipMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"gzip",
			"middleware.gzip",
			middleware.PriorityCompression,
			middleware.RouterAll,
		),
	})
}

// GzipMiddleware handles response compression
type GzipMiddleware struct {
	middleware.BaseMiddleware
	level   int
	minSize int
}

// Enabled checks if gzip middleware is enabled in config
func (m *GzipMiddleware) Enabled(cfg any) bool {
	if cfg == nil {
		return true // Enabled by default
	}

	gzipCfg, ok := cfg.(*config.GzipMiddlewareConfig)
	if !ok {
		return true
	}

	return gzipCfg.Enabled
}

// Configure initializes the middleware with compression settings from config
func (m *GzipMiddleware) Configure(cfg any) error {
	// Defaults
	m.level = 5
	m.minSize = 1024

	// Override with config if provided
	if gzipCfg, ok := cfg.(*config.GzipMiddlewareConfig); ok {
		if gzipCfg.Level >= 1 && gzipCfg.Level <= 9 {
			m.level = gzipCfg.Level
		}
		if gzipCfg.MinSize > 0 {
			m.minSize = gzipCfg.MinSize
		}
	}

	return nil
}

// Handler returns the gzip compression middleware function
func (m *GzipMiddleware) Handler() echo.MiddlewareFunc {
	level := m.level
	if level < 1 || level > 9 {
		level = 5
	}

	minSize := m.minSize
	if minSize <= 0 {
		minSize = 1024
	}

	return echomiddleware.GzipWithConfig(echomiddleware.GzipConfig{
		Level:   level,
		MinLength: minSize,
	})
}
