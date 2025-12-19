package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestIDConfig holds request ID middleware configuration
type RequestIDConfig struct {
	Generator func() string
	Header    string
}

// DefaultRequestIDConfig returns a default request ID configuration
func DefaultRequestIDConfig() *RequestIDConfig {
	return &RequestIDConfig{
		Generator: func() string { return uuid.New().String() },
		Header:    "X-Request-ID",
	}
}

// RequestID returns a request ID middleware
func RequestID(cfg *RequestIDConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultRequestIDConfig()
	}

	generator := cfg.Generator
	if generator == nil {
		generator = func() string { return uuid.New().String() }
	}

	header := cfg.Header
	if header == "" {
		header = "X-Request-ID"
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get(header)
			if requestID == "" {
				requestID = generator()
			}

			c.Request().Header.Set(header, requestID)
			c.Response().Header().Set(header, requestID)

			return next(c)
		}
	}
}

// DefaultRequestID returns a request ID middleware with default settings
func DefaultRequestID() echo.MiddlewareFunc {
	return RequestID(nil)
}

// RequestIDWithGenerator returns a request ID middleware with a custom generator
func RequestIDWithGenerator(generator func() string) echo.MiddlewareFunc {
	return RequestID(&RequestIDConfig{Generator: generator})
}
