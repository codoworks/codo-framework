package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// TimeoutConfig holds timeout middleware configuration
type TimeoutConfig struct {
	Timeout time.Duration
}

// DefaultTimeoutConfig returns a default timeout configuration
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		Timeout: 60 * time.Second,
	}
}

// Timeout returns a request timeout middleware
func Timeout(cfg *TimeoutConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultTimeoutConfig()
	}

	timeout := cfg.Timeout
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

// TimeoutWithDuration returns a timeout middleware with the specified duration
func TimeoutWithDuration(d time.Duration) echo.MiddlewareFunc {
	return Timeout(&TimeoutConfig{Timeout: d})
}
