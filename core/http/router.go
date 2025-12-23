package http

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Router wraps an Echo instance for a specific scope.
type Router struct {
	echo   *echo.Echo
	scope  RouterScope
	addr   string
	server *http.Server
}

// NewRouter creates a new router for the given scope and address.
func NewRouter(scope RouterScope, addr string) *Router {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	return &Router{
		echo:  e,
		scope: scope,
		addr:  addr,
	}
}

// Scope returns the router's scope.
func (r *Router) Scope() RouterScope {
	return r.scope
}

// Addr returns the router's address.
func (r *Router) Addr() string {
	return r.addr
}

// Echo returns the underlying Echo instance.
func (r *Router) Echo() *echo.Echo {
	return r.echo
}

// Use adds middleware to the router.
func (r *Router) Use(middleware ...echo.MiddlewareFunc) {
	r.echo.Use(middleware...)
}

// Group creates a new route group.
func (r *Router) Group(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group {
	return r.echo.Group(prefix, middleware...)
}

// GET registers a GET route.
func (r *Router) GET(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route {
	return r.echo.GET(path, WrapHandler(h), middleware...)
}

// POST registers a POST route.
func (r *Router) POST(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route {
	return r.echo.POST(path, WrapHandler(h), middleware...)
}

// PUT registers a PUT route.
func (r *Router) PUT(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route {
	return r.echo.PUT(path, WrapHandler(h), middleware...)
}

// PATCH registers a PATCH route.
func (r *Router) PATCH(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route {
	return r.echo.PATCH(path, WrapHandler(h), middleware...)
}

// DELETE registers a DELETE route.
func (r *Router) DELETE(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route {
	return r.echo.DELETE(path, WrapHandler(h), middleware...)
}

// OPTIONS registers an OPTIONS route.
func (r *Router) OPTIONS(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route {
	return r.echo.OPTIONS(path, WrapHandler(h), middleware...)
}

// HEAD registers a HEAD route.
func (r *Router) HEAD(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route {
	return r.echo.HEAD(path, WrapHandler(h), middleware...)
}

// Any registers a route for all HTTP methods.
func (r *Router) Any(path string, h HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route {
	return r.echo.Any(path, WrapHandler(h), middleware...)
}

// Static serves static files from a directory.
func (r *Router) Static(prefix, root string) *echo.Route {
	return r.echo.Static(prefix, root)
}

// File serves a single file.
func (r *Router) File(path, file string) *echo.Route {
	return r.echo.File(path, file)
}

// Routes returns all registered routes.
func (r *Router) Routes() []*echo.Route {
	return r.echo.Routes()
}

// RegisterHandlers registers all handlers for this router's scope.
func (r *Router) RegisterHandlers() error {
	handlers := GetHandlers(r.scope)
	for _, h := range handlers {
		if err := h.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize handler %s: %w", h.Prefix(), err)
		}
		g := r.echo.Group(h.Prefix(), h.Middlewares()...)
		h.Routes(g)
	}
	return nil
}

// Start starts the router.
func (r *Router) Start() error {
	r.server = &http.Server{
		Addr:    r.addr,
		Handler: r.echo,
	}

	if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Listen creates the network listener without serving.
// This allows checking if the port is available before starting the server.
// Use with Serve() for a two-phase startup that fails fast on port conflicts.
func (r *Router) Listen() (net.Listener, error) {
	return net.Listen("tcp", r.addr)
}

// Serve starts serving HTTP on an existing listener.
// Use this with Listen() for two-phase startup.
func (r *Router) Serve(ln net.Listener) error {
	r.server = &http.Server{
		Addr:    r.addr,
		Handler: r.echo,
	}

	if err := r.server.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// StartTLS starts the router with TLS.
func (r *Router) StartTLS(certFile, keyFile string) error {
	r.server = &http.Server{
		Addr:    r.addr,
		Handler: r.echo,
	}

	if err := r.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the router.
func (r *Router) Shutdown(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	return r.server.Shutdown(ctx)
}

// SetValidator sets the validator for the router.
func (r *Router) SetValidator(v echo.Validator) {
	r.echo.Validator = v
}

// SetErrorHandler sets a custom error handler.
func (r *Router) SetErrorHandler(h echo.HTTPErrorHandler) {
	r.echo.HTTPErrorHandler = h
}

// ServeHTTP implements http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.echo.ServeHTTP(w, req)
}
