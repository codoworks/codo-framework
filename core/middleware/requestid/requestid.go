package requestid

import (
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func init() {
	middleware.RegisterMiddleware(&RequestIDMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"request-id",
			"", // No configuration needed
			middleware.PriorityRequestID,
			middleware.RouterAll,
		),
	})
}

// RequestIDMiddleware generates or propagates request IDs
type RequestIDMiddleware struct {
	middleware.BaseMiddleware
	generator func() string
	header    string
}

// Enabled always returns true for core middleware
func (m *RequestIDMiddleware) Enabled(cfg any) bool {
	return true // Core middleware is always enabled
}

// Configure initializes the middleware
func (m *RequestIDMiddleware) Configure(cfg any) error {
	// Use default UUID generator
	m.generator = func() string {
		return uuid.New().String()
	}

	// Use standard header
	m.header = "X-Request-ID"

	return nil
}

// Handler returns the request ID middleware function
func (m *RequestIDMiddleware) Handler() echo.MiddlewareFunc {
	generator := m.generator
	if generator == nil {
		generator = func() string { return uuid.New().String() }
	}

	header := m.header
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
