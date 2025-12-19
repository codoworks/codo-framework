package clients

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockClient is a test implementation of the Client interface.
type mockClient struct {
	name          string
	initialized   bool
	healthy       bool
	initErr       error
	healthErr     error
	shutdownErr   error
	initCalled    bool
	healthCalled  bool
	shutdownCalled bool
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
	m.initCalled = true
	if m.initErr != nil {
		return m.initErr
	}
	m.initialized = true
	return nil
}

func (m *mockClient) Health() error {
	m.healthCalled = true
	if !m.healthy {
		return m.healthErr
	}
	return nil
}

func (m *mockClient) Shutdown() error {
	m.shutdownCalled = true
	if m.shutdownErr != nil {
		return m.shutdownErr
	}
	m.initialized = false
	return nil
}

func TestClient_Interface(t *testing.T) {
	t.Run("mockClient implements Client", func(t *testing.T) {
		var _ Client = (*mockClient)(nil)
	})
}

func TestMockClient_Name(t *testing.T) {
	client := newMockClient("test-client")
	assert.Equal(t, "test-client", client.Name())
}

func TestMockClient_Initialize(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		client := newMockClient("test")

		err := client.Initialize(nil)

		assert.NoError(t, err)
		assert.True(t, client.initialized)
		assert.True(t, client.initCalled)
	})

	t.Run("initialization with config", func(t *testing.T) {
		client := newMockClient("test")
		cfg := map[string]string{"key": "value"}

		err := client.Initialize(cfg)

		assert.NoError(t, err)
		assert.True(t, client.initialized)
	})

	t.Run("initialization error", func(t *testing.T) {
		client := newMockClient("test")
		client.initErr = errors.New("init failed")

		err := client.Initialize(nil)

		assert.Error(t, err)
		assert.Equal(t, "init failed", err.Error())
		assert.False(t, client.initialized)
	})
}

func TestMockClient_Health(t *testing.T) {
	t.Run("healthy client", func(t *testing.T) {
		client := newMockClient("test")
		client.healthy = true

		err := client.Health()

		assert.NoError(t, err)
		assert.True(t, client.healthCalled)
	})

	t.Run("unhealthy client", func(t *testing.T) {
		client := newMockClient("test")
		client.healthy = false
		client.healthErr = errors.New("connection lost")

		err := client.Health()

		assert.Error(t, err)
		assert.Equal(t, "connection lost", err.Error())
	})
}

func TestMockClient_Shutdown(t *testing.T) {
	t.Run("successful shutdown", func(t *testing.T) {
		client := newMockClient("test")
		client.Initialize(nil)

		err := client.Shutdown()

		assert.NoError(t, err)
		assert.False(t, client.initialized)
		assert.True(t, client.shutdownCalled)
	})

	t.Run("shutdown error", func(t *testing.T) {
		client := newMockClient("test")
		client.shutdownErr = errors.New("shutdown failed")

		err := client.Shutdown()

		assert.Error(t, err)
		assert.Equal(t, "shutdown failed", err.Error())
	})
}

func TestBaseClient_NewBaseClient(t *testing.T) {
	bc := NewBaseClient("my-client")

	assert.Equal(t, "my-client", bc.Name())
	assert.False(t, bc.IsInitialized())
}

func TestBaseClient_Name(t *testing.T) {
	bc := NewBaseClient("test-name")
	assert.Equal(t, "test-name", bc.Name())
}

func TestBaseClient_Initialize(t *testing.T) {
	bc := NewBaseClient("test")

	err := bc.Initialize(nil)

	assert.NoError(t, err)
	assert.True(t, bc.IsInitialized())
}

func TestBaseClient_Health(t *testing.T) {
	bc := NewBaseClient("test")

	err := bc.Health()

	assert.NoError(t, err)
}

func TestBaseClient_Shutdown(t *testing.T) {
	bc := NewBaseClient("test")
	bc.Initialize(nil)

	err := bc.Shutdown()

	assert.NoError(t, err)
	assert.False(t, bc.IsInitialized())
}

func TestBaseClient_IsInitialized(t *testing.T) {
	bc := NewBaseClient("test")

	assert.False(t, bc.IsInitialized())

	bc.Initialize(nil)
	assert.True(t, bc.IsInitialized())

	bc.Shutdown()
	assert.False(t, bc.IsInitialized())
}

func TestBaseClient_ImplementsClient(t *testing.T) {
	// Ensure BaseClient can be used as a Client
	bc := NewBaseClient("test")
	var _ Client = &bc
}

// configurableTestClient is a test implementation of Configurable.
type configurableTestClient struct {
	BaseClient
}

func (c *configurableTestClient) ConfigType() any {
	return struct{ Port int }{}
}

// startableTestClient is a test implementation of Startable.
type startableTestClient struct {
	BaseClient
	started bool
}

func (s *startableTestClient) Start() error {
	s.started = true
	return nil
}

// stoppableTestClient is a test implementation of Stoppable.
type stoppableTestClient struct {
	BaseClient
	stopped bool
}

func (s *stoppableTestClient) Stop() error {
	s.stopped = true
	return nil
}

// reloadableTestClient is a test implementation of Reloadable.
type reloadableTestClient struct {
	BaseClient
	reloaded bool
}

func (r *reloadableTestClient) Reload(cfg any) error {
	r.reloaded = true
	return nil
}

// TestOptionalInterfaces verifies that optional interfaces can be implemented
func TestOptionalInterfaces(t *testing.T) {
	t.Run("Configurable interface", func(t *testing.T) {
		client := &configurableTestClient{BaseClient: NewBaseClient("configurable")}
		var _ Configurable = client
		assert.NotNil(t, client.ConfigType())
	})

	t.Run("Startable interface", func(t *testing.T) {
		impl := &startableTestClient{BaseClient: NewBaseClient("startable")}
		var _ Startable = impl
		assert.NoError(t, impl.Start())
		assert.True(t, impl.started)
	})

	t.Run("Stoppable interface", func(t *testing.T) {
		impl := &stoppableTestClient{BaseClient: NewBaseClient("stoppable")}
		var _ Stoppable = impl
		assert.NoError(t, impl.Stop())
		assert.True(t, impl.stopped)
	})

	t.Run("Reloadable interface", func(t *testing.T) {
		impl := &reloadableTestClient{BaseClient: NewBaseClient("reloadable")}
		var _ Reloadable = impl
		assert.NoError(t, impl.Reload(nil))
		assert.True(t, impl.reloaded)
	})
}
