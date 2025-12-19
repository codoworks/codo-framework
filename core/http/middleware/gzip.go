package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// GzipConfig holds gzip middleware configuration
type GzipConfig struct {
	Level int // 1-9, default 5
}

// DefaultGzipConfig returns a default gzip configuration
func DefaultGzipConfig() *GzipConfig {
	return &GzipConfig{
		Level: 5,
	}
}

// Gzip returns a gzip compression middleware
func Gzip(cfg *GzipConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultGzipConfig()
	}

	level := cfg.Level
	if level < 1 || level > 9 {
		level = 5
	}

	return middleware.GzipWithConfig(middleware.GzipConfig{
		Level: level,
	})
}

// GzipWithLevel returns a gzip middleware with the specified compression level
func GzipWithLevel(level int) echo.MiddlewareFunc {
	return Gzip(&GzipConfig{Level: level})
}

// DefaultGzip returns a gzip middleware with default settings
func DefaultGzip() echo.MiddlewareFunc {
	return Gzip(nil)
}
