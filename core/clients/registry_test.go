package clients

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) {
	t.Helper()
	ResetRegistry()
}

func TestRegister(t *testing.T) {
	setupTest(t)

	client := newMockClient("test-client")
	err := Register(client)

	assert.NoError(t, err)
	assert.True(t, Has("test-client"))
}

func TestRegister_Duplicate(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("duplicate")
	client2 := newMockClient("duplicate")

	err1 := Register(client1)
	err2 := Register(client2)

	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already registered")
}

func TestMustRegister(t *testing.T) {
	setupTest(t)

	t.Run("successful registration", func(t *testing.T) {
		client := newMockClient("must-register")
		assert.NotPanics(t, func() {
			MustRegister(client)
		})
		assert.True(t, Has("must-register"))
	})
}

func TestMustRegister_Panics(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("panic-test")
	client2 := newMockClient("panic-test")

	MustRegister(client1)

	assert.Panics(t, func() {
		MustRegister(client2)
	})
}

func TestGet(t *testing.T) {
	setupTest(t)

	expected := newMockClient("get-test")
	Register(expected)

	client, err := Get("get-test")

	assert.NoError(t, err)
	assert.Equal(t, expected, client)
}

func TestGet_NotFound(t *testing.T) {
	setupTest(t)

	client, err := Get("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, client)
}

func TestMustGet(t *testing.T) {
	setupTest(t)

	expected := newMockClient("must-get-test")
	Register(expected)

	var client Client
	assert.NotPanics(t, func() {
		client = MustGet("must-get-test")
	})
	assert.Equal(t, expected, client)
}

func TestMustGet_Panics(t *testing.T) {
	setupTest(t)

	assert.Panics(t, func() {
		MustGet("nonexistent")
	})
}

func TestHas(t *testing.T) {
	setupTest(t)

	client := newMockClient("has-test")
	Register(client)

	assert.True(t, Has("has-test"))
	assert.False(t, Has("nonexistent"))
}

// typedMockClient is a specific type for testing GetTyped
type typedMockClient struct {
	*mockClient
	specialField string
}

func newTypedMockClient(name, special string) *typedMockClient {
	return &typedMockClient{
		mockClient:   newMockClient(name),
		specialField: special,
	}
}

func TestGetTyped(t *testing.T) {
	setupTest(t)

	expected := newTypedMockClient("typed-test", "special-value")
	Register(expected)

	client, err := GetTyped[*typedMockClient]("typed-test")

	assert.NoError(t, err)
	assert.Equal(t, expected, client)
	assert.Equal(t, "special-value", client.specialField)
}

func TestGetTyped_NotFound(t *testing.T) {
	setupTest(t)

	client, err := GetTyped[*typedMockClient]("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, client)
}

func TestGetTyped_WrongType(t *testing.T) {
	setupTest(t)

	// Register a regular mock client
	regularClient := newMockClient("regular")
	Register(regularClient)

	// Try to get it as a typed mock client
	client, err := GetTyped[*typedMockClient]("regular")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not of expected type")
	assert.Nil(t, client)
}

func TestMustGetTyped(t *testing.T) {
	setupTest(t)

	expected := newTypedMockClient("must-typed-test", "value")
	Register(expected)

	var client *typedMockClient
	assert.NotPanics(t, func() {
		client = MustGetTyped[*typedMockClient]("must-typed-test")
	})
	assert.Equal(t, expected, client)
}

func TestMustGetTyped_Panics(t *testing.T) {
	setupTest(t)

	t.Run("not found", func(t *testing.T) {
		assert.Panics(t, func() {
			MustGetTyped[*typedMockClient]("nonexistent")
		})
	})

	t.Run("wrong type", func(t *testing.T) {
		setupTest(t)
		regularClient := newMockClient("regular")
		Register(regularClient)

		assert.Panics(t, func() {
			MustGetTyped[*typedMockClient]("regular")
		})
	})
}

func TestAll(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("client1")
	client2 := newMockClient("client2")
	client3 := newMockClient("client3")

	Register(client1)
	Register(client2)
	Register(client3)

	all := All()

	assert.Len(t, all, 3)
	assert.Equal(t, client1, all["client1"])
	assert.Equal(t, client2, all["client2"])
	assert.Equal(t, client3, all["client3"])
}

func TestNames(t *testing.T) {
	setupTest(t)

	Register(newMockClient("alpha"))
	Register(newMockClient("beta"))
	Register(newMockClient("gamma"))

	names := Names()

	assert.Len(t, names, 3)
	assert.Contains(t, names, "alpha")
	assert.Contains(t, names, "beta")
	assert.Contains(t, names, "gamma")
}

func TestCount(t *testing.T) {
	setupTest(t)

	assert.Equal(t, 0, Count())

	Register(newMockClient("one"))
	assert.Equal(t, 1, Count())

	Register(newMockClient("two"))
	assert.Equal(t, 2, Count())
}

func TestInitializeAll(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("client1")
	client2 := newMockClient("client2")

	Register(client1)
	Register(client2)

	configs := map[string]any{
		"client1": map[string]string{"key": "value1"},
		"client2": map[string]string{"key": "value2"},
	}

	err := InitializeAll(configs)

	assert.NoError(t, err)
	assert.True(t, client1.initCalled)
	assert.True(t, client2.initCalled)
	assert.True(t, client1.initialized)
	assert.True(t, client2.initialized)
}

func TestInitializeAll_Error(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("client1")
	client2 := newMockClient("client2")
	client2.initErr = errors.New("init failed")

	Register(client1)
	Register(client2)

	err := InitializeAll(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize client")
}

func TestInitializeAll_EmptyConfigs(t *testing.T) {
	setupTest(t)

	client := newMockClient("client")
	Register(client)

	err := InitializeAll(nil)

	assert.NoError(t, err)
	assert.True(t, client.initCalled)
}

func TestShutdownAll(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("client1")
	client2 := newMockClient("client2")

	Register(client1)
	Register(client2)

	client1.Initialize(nil)
	client2.Initialize(nil)

	err := ShutdownAll()

	assert.NoError(t, err)
	assert.True(t, client1.shutdownCalled)
	assert.True(t, client2.shutdownCalled)
}

func TestShutdownAll_WithErrors(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("client1")
	client2 := newMockClient("client2")
	client2.shutdownErr = errors.New("shutdown failed")

	Register(client1)
	Register(client2)

	err := ShutdownAll()

	// Should return error but still try to shutdown all
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to shutdown client")
	assert.True(t, client1.shutdownCalled)
	assert.True(t, client2.shutdownCalled)
}

func TestHealthAll(t *testing.T) {
	setupTest(t)

	client1 := newMockClient("healthy1")
	client2 := newMockClient("healthy2")
	client3 := newMockClient("unhealthy")
	client3.healthy = false
	client3.healthErr = errors.New("connection lost")

	Register(client1)
	Register(client2)
	Register(client3)

	results := HealthAll()

	assert.Len(t, results, 3)
	assert.NoError(t, results["healthy1"])
	assert.NoError(t, results["healthy2"])
	assert.Error(t, results["unhealthy"])
	assert.Equal(t, "connection lost", results["unhealthy"].Error())
}

func TestHealthCheck(t *testing.T) {
	t.Run("all healthy", func(t *testing.T) {
		setupTest(t)

		Register(newMockClient("healthy1"))
		Register(newMockClient("healthy2"))

		err := HealthCheck()

		assert.NoError(t, err)
	})

	t.Run("some unhealthy", func(t *testing.T) {
		setupTest(t)

		healthy := newMockClient("healthy")
		unhealthy := newMockClient("unhealthy")
		unhealthy.healthy = false
		unhealthy.healthErr = errors.New("not ready")

		Register(healthy)
		Register(unhealthy)

		err := HealthCheck()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unhealthy")
	})
}

func TestInitialize(t *testing.T) {
	setupTest(t)

	client := newMockClient("single-init")
	Register(client)

	err := Initialize("single-init", map[string]string{"key": "value"})

	assert.NoError(t, err)
	assert.True(t, client.initCalled)
}

func TestInitialize_NotFound(t *testing.T) {
	setupTest(t)

	err := Initialize("nonexistent", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestShutdown(t *testing.T) {
	setupTest(t)

	client := newMockClient("single-shutdown")
	Register(client)

	err := Shutdown("single-shutdown")

	assert.NoError(t, err)
	assert.True(t, client.shutdownCalled)
}

func TestShutdown_NotFound(t *testing.T) {
	setupTest(t)

	err := Shutdown("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealth(t *testing.T) {
	t.Run("healthy client", func(t *testing.T) {
		setupTest(t)

		client := newMockClient("health-check")
		Register(client)

		err := Health("health-check")

		assert.NoError(t, err)
	})

	t.Run("unhealthy client", func(t *testing.T) {
		setupTest(t)

		client := newMockClient("unhealthy")
		client.healthy = false
		client.healthErr = errors.New("connection refused")
		Register(client)

		err := Health("unhealthy")

		assert.Error(t, err)
		assert.Equal(t, "connection refused", err.Error())
	})

	t.Run("not found", func(t *testing.T) {
		setupTest(t)

		err := Health("nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestResetRegistry(t *testing.T) {
	setupTest(t)

	Register(newMockClient("test1"))
	Register(newMockClient("test2"))
	assert.Equal(t, 2, Count())

	ResetRegistry()
	assert.Equal(t, 0, Count())

	// Should be able to register again
	err := Register(newMockClient("test1"))
	assert.NoError(t, err)
}
