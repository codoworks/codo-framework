package http

import (
	"github.com/labstack/echo/v4"
)

// Handler defines the interface for HTTP handlers
type Handler interface {
	// Prefix returns the URL prefix for all routes (e.g., "/api/v1/users")
	Prefix() string

	// Scope returns which router this handler belongs to
	Scope() RouterScope

	// Middlewares returns handler-specific middlewares
	Middlewares() []echo.MiddlewareFunc

	// Initialize is called once during startup for dependency injection
	Initialize() error

	// Routes registers the handler's routes on the group
	Routes(g *echo.Group)
}

// HandlerFunc is a function that handles HTTP requests with our extended Context
type HandlerFunc func(*Context) error

// WrapHandler wraps a HandlerFunc to work with Echo
func WrapHandler(fn HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &Context{Context: c}
		return fn(cc)
	}
}

// EchoHandler is the standard Echo handler function type
type EchoHandler = echo.HandlerFunc

// Middleware is a function that wraps a handler
type Middleware func(next HandlerFunc) HandlerFunc

// WrapMiddleware wraps our Middleware to work with Echo
func WrapMiddleware(m Middleware) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{Context: c}
			wrappedNext := func(ctx *Context) error {
				return next(ctx.Context)
			}
			return m(wrappedNext)(cc)
		}
	}
}
