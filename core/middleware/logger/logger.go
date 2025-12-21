package logger

import (
	"fmt"
	"time"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func init() {
	middleware.RegisterMiddleware(&LoggerMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"logger",
			"middleware.logger",
			middleware.PriorityLogger,
			middleware.RouterAll,
		),
	})
}

// LoggerMiddleware handles request/response logging
type LoggerMiddleware struct {
	middleware.BaseMiddleware
	logger    *logrus.Logger
	skipPaths map[string]bool
	cfg       *config.Config
}

// Enabled checks if logger middleware is enabled in config
func (m *LoggerMiddleware) Enabled(cfg any) bool {
	if cfg == nil {
		return true // Enabled by default
	}

	loggerCfg, ok := cfg.(*config.LoggerMiddlewareConfig)
	if !ok {
		return true
	}

	return loggerCfg.Enabled
}

// Configure initializes the middleware with logger from client registry
func (m *LoggerMiddleware) Configure(cfg any) error {
	// Get logger from client registry
	loggerClient, err := clients.GetTyped[*logger.Logger]("logger")
	if err != nil {
		return fmt.Errorf("failed to get logger client: %w", err)
	}

	m.logger = loggerClient.GetLogger()

	// Store config for dev mode check
	if mainCfg, ok := cfg.(*config.Config); ok {
		m.cfg = mainCfg
	}

	// Parse skip paths from config
	m.skipPaths = make(map[string]bool)
	if loggerCfg, ok := cfg.(*config.LoggerMiddlewareConfig); ok {
		for _, path := range loggerCfg.SkipPaths {
			m.skipPaths[path] = true
		}
	}

	return nil
}

// Handler returns the logger middleware function
func (m *LoggerMiddleware) Handler() echo.MiddlewareFunc {
	log := m.logger
	if log == nil {
		log = logrus.StandardLogger()
	}

	skipPaths := m.skipPaths
	if skipPaths == nil {
		skipPaths = make(map[string]bool)
	}

	// Determine dev mode
	devMode := false
	if m.cfg != nil {
		devMode = m.cfg.IsDevMode()
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

			if devMode {
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
