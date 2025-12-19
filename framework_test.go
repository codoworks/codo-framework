package codo

import (
	"context"
	"errors"
	"testing"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/stretchr/testify/assert"
)

// mockClient for testing
type mockClient struct {
	name          string
	initialized   bool
	healthy       bool
	initErr       error
	healthErr     error
	shutdownErr   error
}

func newMockClient(name string) *mockClient {
	return &mockClient{
		name:    name,
		healthy: true,
	}
}

func (m *mockClient) Name() string {
	return m.name
}

func (m *mockClient) Initialize(cfg any) error {
	if m.initErr != nil {
		return m.initErr
	}
	m.initialized = true
	return nil
}

func (m *mockClient) Health() error {
	if !m.healthy {
		return m.healthErr
	}
	return nil
}

func (m *mockClient) Shutdown() error {
	if m.shutdownErr != nil {
		return m.shutdownErr
	}
	m.initialized = false
	return nil
}

func setupTest(t *testing.T) {
	t.Helper()
	clients.ResetRegistry()
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateNew, "new"},
		{StateInitialized, "initialized"},
		{StateRunning, "running"},
		{StateStopped, "stopped"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "codo-app", cfg.Name)
	assert.Equal(t, "development", cfg.Environment)
	assert.NotNil(t, cfg.Clients)
}

func TestNew(t *testing.T) {
	f := New()

	assert.NotNil(t, f)
	assert.NotNil(t, f.config)
	assert.Equal(t, StateNew, f.state)
}

func TestNewWithConfig(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		cfg := &Config{
			Name:        "my-app",
			Environment: "production",
		}

		f := NewWithConfig(cfg)

		assert.Equal(t, "my-app", f.config.Name)
		assert.Equal(t, "production", f.config.Environment)
	})

	t.Run("with nil config", func(t *testing.T) {
		f := NewWithConfig(nil)

		assert.NotNil(t, f.config)
		assert.Equal(t, "codo-app", f.config.Name)
	})
}

func TestFramework_Config(t *testing.T) {
	cfg := &Config{Name: "test-app"}
	f := NewWithConfig(cfg)

	assert.Equal(t, cfg, f.Config())
}

func TestFramework_State(t *testing.T) {
	f := New()

	assert.Equal(t, StateNew, f.State())
}

func TestFramework_IsInitialized(t *testing.T) {
	setupTest(t)
	f := New()

	assert.False(t, f.IsInitialized())

	f.Initialize()
	assert.True(t, f.IsInitialized())
}

func TestFramework_IsRunning(t *testing.T) {
	setupTest(t)
	f := New()

	assert.False(t, f.IsRunning())

	f.Run()
	assert.True(t, f.IsRunning())
}

func TestFramework_Initialize(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		setupTest(t)
		f := New()

		err := f.Initialize()

		assert.NoError(t, err)
		assert.Equal(t, StateInitialized, f.state)
	})

	t.Run("already initialized", func(t *testing.T) {
		setupTest(t)
		f := New()

		f.Initialize()
		err := f.Initialize()

		assert.NoError(t, err)
	})

	t.Run("with clients", func(t *testing.T) {
		setupTest(t)
		client := newMockClient("test-client")
		clients.Register(client)

		f := NewWithConfig(&Config{
			Clients: map[string]any{
				"test-client": map[string]string{"key": "value"},
			},
		})

		err := f.Initialize()

		assert.NoError(t, err)
		assert.True(t, client.initialized)
	})

	t.Run("client initialization error", func(t *testing.T) {
		setupTest(t)
		client := newMockClient("failing-client")
		client.initErr = errors.New("init failed")
		clients.Register(client)

		f := New()

		err := f.Initialize()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize clients")
	})
}

func TestFramework_Run(t *testing.T) {
	t.Run("run from new state", func(t *testing.T) {
		setupTest(t)
		f := New()

		err := f.Run()

		assert.NoError(t, err)
		assert.Equal(t, StateRunning, f.state)
	})

	t.Run("run from initialized state", func(t *testing.T) {
		setupTest(t)
		f := New()
		f.Initialize()

		err := f.Run()

		assert.NoError(t, err)
		assert.Equal(t, StateRunning, f.state)
	})

	t.Run("already running", func(t *testing.T) {
		setupTest(t)
		f := New()
		f.Run()

		err := f.Run()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")
	})

	t.Run("initialization fails during run", func(t *testing.T) {
		setupTest(t)
		client := newMockClient("failing")
		client.initErr = errors.New("init failed")
		clients.Register(client)

		f := New()

		err := f.Run()

		assert.Error(t, err)
	})
}

func TestFramework_Shutdown(t *testing.T) {
	t.Run("successful shutdown", func(t *testing.T) {
		setupTest(t)
		f := New()
		f.Run()

		err := f.Shutdown(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, StateStopped, f.state)
	})

	t.Run("already stopped", func(t *testing.T) {
		setupTest(t)
		f := New()
		f.Run()
		f.Shutdown(context.Background())

		err := f.Shutdown(context.Background())

		assert.NoError(t, err)
	})

	t.Run("shutdown with clients", func(t *testing.T) {
		setupTest(t)
		client := newMockClient("shutdown-test")
		clients.Register(client)

		f := New()
		f.Run()

		err := f.Shutdown(context.Background())

		assert.NoError(t, err)
		assert.False(t, client.initialized)
	})

	t.Run("client shutdown error", func(t *testing.T) {
		setupTest(t)
		client := newMockClient("failing-shutdown")
		client.shutdownErr = errors.New("shutdown failed")
		clients.Register(client)

		f := New()
		f.Run()

		err := f.Shutdown(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to shutdown clients")
	})
}

func TestFramework_Health(t *testing.T) {
	t.Run("all healthy", func(t *testing.T) {
		setupTest(t)
		client := newMockClient("healthy")
		clients.Register(client)

		f := New()
		f.Initialize()

		err := f.Health()

		assert.NoError(t, err)
	})

	t.Run("unhealthy client", func(t *testing.T) {
		setupTest(t)
		client := newMockClient("unhealthy")
		client.healthy = false
		client.healthErr = errors.New("connection lost")
		clients.Register(client)

		f := New()
		f.Initialize()

		err := f.Health()

		assert.Error(t, err)
	})
}

func TestFramework_HealthAll(t *testing.T) {
	setupTest(t)

	healthy := newMockClient("healthy")
	unhealthy := newMockClient("unhealthy")
	unhealthy.healthy = false
	unhealthy.healthErr = errors.New("error")

	clients.Register(healthy)
	clients.Register(unhealthy)

	f := New()
	f.Initialize()

	results := f.HealthAll()

	assert.Len(t, results, 2)
	assert.NoError(t, results["healthy"])
	assert.Error(t, results["unhealthy"])
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	assert.Equal(t, "0.0.0-dev", version)
}

func TestVersion(t *testing.T) {
	assert.Equal(t, "0.0.0-dev", Version)
}
