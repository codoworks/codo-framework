package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// LoggerConfig holds logger middleware configuration
type LoggerConfig struct {
	Logger    *logrus.Logger
	SkipPaths []string
	DevMode   bool
}

// DefaultLoggerConfig returns a default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Logger:    logrus.StandardLogger(),
		SkipPaths: nil,
		DevMode:   false,
	}
}

// Logger returns a request logging middleware
func Logger(cfg *LoggerConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultLoggerConfig()
	}

	log := cfg.Logger
	if log == nil {
		log = logrus.StandardLogger()
	}

	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip logging for certain paths
			if skipPaths[c.Request().URL.Path] {
				return next(c)
			}

			start := time.Now()

			err := next(c)

			latency := time.Since(start)

			fields := logrus.Fields{
				"method":     c.Request().Method,
				"path":       c.Request().URL.Path,
				"status":     c.Response().Status,
				"latency_ms": latency.Milliseconds(),
				"ip":         c.RealIP(),
			}

			if requestID := c.Request().Header.Get("X-Request-ID"); requestID != "" {
				fields["request_id"] = requestID
			}

			if cfg.DevMode {
				fields["query"] = c.Request().URL.RawQuery
				fields["user_agent"] = c.Request().UserAgent()
			}

			if err != nil {
				fields["error"] = err.Error()
				log.WithFields(fields).Error("Request failed")
			} else if c.Response().Status >= 500 {
				log.WithFields(fields).Error("Request completed with server error")
			} else if c.Response().Status >= 400 {
				log.WithFields(fields).Warn("Request completed with client error")
			} else {
				log.WithFields(fields).Info("Request completed")
			}

			return err
		}
	}
}
