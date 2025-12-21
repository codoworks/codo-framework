package timeout

import (
	"context"
	"net/http"
	"time"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/labstack/echo/v4"
)

func init() {
	middleware.RegisterMiddleware(&TimeoutMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"timeout",
			"middleware.timeout",
			middleware.PriorityTimeout,
			middleware.RouterAll,
		),
	})
}

// TimeoutMiddleware handles request timeouts
type TimeoutMiddleware struct {
	middleware.BaseMiddleware
	timeout time.Duration
}

// Enabled checks if timeout middleware is enabled in config
func (m *TimeoutMiddleware) Enabled(cfg any) bool {
	if cfg == nil {
		return true // Enabled by default
	}

	timeoutCfg, ok := cfg.(*config.TimeoutMiddlewareConfig)
	if !ok {
		return true
	}

	return timeoutCfg.Enabled
}

// Configure initializes the middleware with timeout duration from config
func (m *TimeoutMiddleware) Configure(cfg any) error {
	// Default timeout
	m.timeout = 60 * time.Second

	// Override with config if provided
	if timeoutCfg, ok := cfg.(*config.TimeoutMiddlewareConfig); ok {
		if timeoutCfg.Duration > 0 {
			m.timeout = timeoutCfg.Duration
		}
	}

	return nil
}

// Handler returns the timeout middleware function
func (m *TimeoutMiddleware) Handler() echo.MiddlewareFunc {
	timeout := m.timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(ctx))

			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return c.JSON(http.StatusGatewayTimeout, map[string]string{
					"code":    "TIMEOUT",
					"message": "Request timed out",
				})
			}
		}
	}
}
