package http

import (
	"context"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()

	assert.Equal(t, ":8081", cfg.PublicAddr)
	assert.Equal(t, ":8080", cfg.ProtectedAddr)
	assert.Equal(t, ":8079", cfg.HiddenAddr)
	assert.Equal(t, 30*time.Second, cfg.ShutdownGrace)
}

func TestNewServer(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		cfg := &ServerConfig{
			PublicAddr:    ":9081",
			ProtectedAddr: ":9080",
			HiddenAddr:    ":9079",
		}

		s := NewServer(cfg)

		assert.NotNil(t, s)
		assert.NotNil(t, s.PublicRouter())
		assert.NotNil(t, s.ProtectedRouter())
		assert.NotNil(t, s.HiddenRouter())
		assert.Equal(t, ":9081", s.PublicRouter().Addr())
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		s := NewServer(nil)

		assert.NotNil(t, s)
		assert.Equal(t, ":8081", s.PublicRouter().Addr())
		assert.Equal(t, ":8080", s.ProtectedRouter().Addr())
		assert.Equal(t, ":8079", s.HiddenRouter().Addr())
	})
}

func TestServer_PublicRouter(t *testing.T) {
	s := NewServer(nil)
	r := s.PublicRouter()

	assert.NotNil(t, r)
	assert.Equal(t, ScopePublic, r.Scope())
}

func TestServer_ProtectedRouter(t *testing.T) {
	s := NewServer(nil)
	r := s.ProtectedRouter()

	assert.NotNil(t, r)
	assert.Equal(t, ScopeProtected, r.Scope())
}

func TestServer_HiddenRouter(t *testing.T) {
	s := NewServer(nil)
	r := s.HiddenRouter()

	assert.NotNil(t, r)
	assert.Equal(t, ScopeHidden, r.Scope())
}

func TestServer_Router(t *testing.T) {
	s := NewServer(nil)

	t.Run("public", func(t *testing.T) {
		r := s.Router(ScopePublic)
		assert.Equal(t, s.PublicRouter(), r)
	})

	t.Run("protected", func(t *testing.T) {
		r := s.Router(ScopeProtected)
		assert.Equal(t, s.ProtectedRouter(), r)
	})

	t.Run("hidden", func(t *testing.T) {
		r := s.Router(ScopeHidden)
		assert.Equal(t, s.HiddenRouter(), r)
	})

	t.Run("unknown", func(t *testing.T) {
		r := s.Router(RouterScope(99))
		assert.Nil(t, r)
	})
}

func TestServer_Use(t *testing.T) {
	s := NewServer(nil)

	called := 0
	s.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			called++
			return next(c)
		}
	})

	// The middleware should be added to all routers
	// We can't easily test this without starting the server
	// but we can verify no panic occurs
	assert.NotPanics(t, func() {
		s.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		})
	})
}

func TestServer_UsePublic(t *testing.T) {
	s := NewServer(nil)

	assert.NotPanics(t, func() {
		s.UsePublic(func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		})
	})
}

func TestServer_UseProtected(t *testing.T) {
	s := NewServer(nil)

	assert.NotPanics(t, func() {
		s.UseProtected(func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		})
	})
}

func TestServer_UseHidden(t *testing.T) {
	s := NewServer(nil)

	assert.NotPanics(t, func() {
		s.UseHidden(func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		})
	})
}

func TestServer_Config(t *testing.T) {
	cfg := &ServerConfig{
		PublicAddr: ":9999",
	}
	s := NewServer(cfg)

	assert.Equal(t, cfg, s.Config())
}

func TestServer_IsStarted(t *testing.T) {
	s := NewServer(nil)

	assert.False(t, s.IsStarted())
}

func TestServer_Shutdown_NotStarted(t *testing.T) {
	s := NewServer(nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestServer_Start(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	cfg := &ServerConfig{
		PublicAddr:    "127.0.0.1:0",
		ProtectedAddr: "127.0.0.1:0",
		HiddenAddr:    "127.0.0.1:0",
		ShutdownGrace: time.Second,
	}
	s := NewServer(cfg)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	assert.True(t, s.IsStarted())

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	assert.NoError(t, err)

	// Wait for Start to return
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after Shutdown")
	}
}

func TestServer_Start_AlreadyStarted(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	cfg := &ServerConfig{
		PublicAddr:    "127.0.0.1:0",
		ProtectedAddr: "127.0.0.1:0",
		HiddenAddr:    "127.0.0.1:0",
		ShutdownGrace: time.Second,
	}
	s := NewServer(cfg)

	// Start server in goroutine
	go func() {
		s.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Try to start again - should error
	err := s.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.Shutdown(ctx)
}
