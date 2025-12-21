package http

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

// ServerConfig holds configuration for the multi-router server.
type ServerConfig struct {
	PublicAddr    string
	ProtectedAddr string
	HiddenAddr    string
	ShutdownGrace time.Duration
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		PublicAddr:    ":8081",
		ProtectedAddr: ":8080",
		HiddenAddr:    ":8079",
		ShutdownGrace: 30 * time.Second,
	}
}

// Server manages multiple routers.
type Server struct {
	config    *ServerConfig
	public    *Router
	protected *Router
	hidden    *Router
	started   bool
	mu        sync.Mutex
}

// NewServer creates a new multi-router server.
func NewServer(cfg *ServerConfig) *Server {
	if cfg == nil {
		cfg = DefaultServerConfig()
	}
	return &Server{
		config:    cfg,
		public:    NewRouter(ScopePublic, cfg.PublicAddr),
		protected: NewRouter(ScopeProtected, cfg.ProtectedAddr),
		hidden:    NewRouter(ScopeHidden, cfg.HiddenAddr),
	}
}

// PublicRouter returns the public router.
func (s *Server) PublicRouter() *Router {
	return s.public
}

// ProtectedRouter returns the protected router.
func (s *Server) ProtectedRouter() *Router {
	return s.protected
}

// HiddenRouter returns the hidden router.
func (s *Server) HiddenRouter() *Router {
	return s.hidden
}

// Router returns the router for a specific scope.
func (s *Server) Router(scope RouterScope) *Router {
	switch scope {
	case ScopePublic:
		return s.public
	case ScopeProtected:
		return s.protected
	case ScopeHidden:
		return s.hidden
	default:
		return nil
	}
}

// PrepareRoutes registers handlers on all routers without starting the servers.
// This is useful for commands like `info routes` that need to introspect routes
// without actually starting the HTTP listeners.
func (s *Server) PrepareRoutes() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.public.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register public handlers: %w", err)
	}

	if err := s.protected.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register protected handlers: %w", err)
	}

	if err := s.hidden.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register hidden handlers: %w", err)
	}

	return nil
}

// Use adds middleware to all routers.
func (s *Server) Use(middleware ...echo.MiddlewareFunc) {
	s.public.Use(middleware...)
	s.protected.Use(middleware...)
	s.hidden.Use(middleware...)
}

// UsePublic adds middleware to the public router only.
func (s *Server) UsePublic(middleware ...echo.MiddlewareFunc) {
	s.public.Use(middleware...)
}

// UseProtected adds middleware to the protected router only.
func (s *Server) UseProtected(middleware ...echo.MiddlewareFunc) {
	s.protected.Use(middleware...)
}

// UseHidden adds middleware to the hidden router only.
func (s *Server) UseHidden(middleware ...echo.MiddlewareFunc) {
	s.hidden.Use(middleware...)
}

// Start starts all routers.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}
	s.started = true
	s.mu.Unlock()

	// Register handlers for each router
	if err := s.public.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register public handlers: %w", err)
	}
	if err := s.protected.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register protected handlers: %w", err)
	}
	if err := s.hidden.RegisterHandlers(); err != nil {
		return fmt.Errorf("failed to register hidden handlers: %w", err)
	}

	g := new(errgroup.Group)

	g.Go(func() error {
		return s.public.Start()
	})

	g.Go(func() error {
		return s.protected.Start()
	})

	g.Go(func() error {
		return s.hidden.Start()
	})

	return g.Wait()
}

// Shutdown gracefully shuts down all routers.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	g := new(errgroup.Group)

	g.Go(func() error {
		return s.public.Shutdown(ctx)
	})

	g.Go(func() error {
		return s.protected.Shutdown(ctx)
	})

	g.Go(func() error {
		return s.hidden.Shutdown(ctx)
	})

	return g.Wait()
}

// Config returns the server configuration.
func (s *Server) Config() *ServerConfig {
	return s.config
}

// IsStarted returns whether the server has been started.
func (s *Server) IsStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}
