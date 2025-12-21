package kratos

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/auth"
)

func TestNewMockClient(t *testing.T) {
	client := NewMockClient()

	assert.NotNil(t, client)
	assert.NotNil(t, client.ValidateFunc)
	assert.NotNil(t, client.HealthFunc)
}

func TestMockClient_Name(t *testing.T) {
	client := NewMockClient()
	assert.Equal(t, "kratos", client.Name())
}

func TestMockClient_Initialize(t *testing.T) {
	client := NewMockClient()
	err := client.Initialize(nil)
	assert.NoError(t, err)
}

func TestMockClient_Shutdown(t *testing.T) {
	client := NewMockClient()
	err := client.Shutdown()
	assert.NoError(t, err)
}

func TestMockClient_Health(t *testing.T) {
	client := NewMockClient()
	err := client.Health()
	assert.NoError(t, err)
}

func TestMockClient_Health_WithFunc(t *testing.T) {
	client := NewMockClient()
	client.HealthFunc = func() error {
		return fmt.Errorf("mock error")
	}

	err := client.Health()
	assert.Error(t, err)
	assert.Equal(t, "mock error", err.Error())
}

func TestMockClient_Health_NilFunc(t *testing.T) {
	client := &MockClient{}
	err := client.Health()
	assert.NoError(t, err)
}

func TestMockClient_ValidateSession(t *testing.T) {
	client := NewMockClient()

	identity, err := client.ValidateSession(context.Background(), "test-cookie")

	assert.NoError(t, err)
	assert.Equal(t, "test-user", identity.ID)
}

func TestMockClient_ValidateSession_WithFunc(t *testing.T) {
	client := NewMockClient()
	client.ValidateFunc = func(ctx context.Context, cookie string) (*auth.Identity, error) {
		return &auth.Identity{
			ID: "custom-user",
			Traits: map[string]any{
				"email": "custom@example.com",
			},
		}, nil
	}

	identity, err := client.ValidateSession(context.Background(), "test-cookie")

	assert.NoError(t, err)
	assert.Equal(t, "custom-user", identity.ID)
	assert.Equal(t, "custom@example.com", identity.GetTraitString("email"))
}

func TestMockClient_ValidateSession_Error(t *testing.T) {
	client := NewMockClient()
	client.ValidateFunc = func(ctx context.Context, cookie string) (*auth.Identity, error) {
		return nil, ErrInvalidSession
	}

	identity, err := client.ValidateSession(context.Background(), "invalid-cookie")

	assert.ErrorIs(t, err, ErrInvalidSession)
	assert.Nil(t, identity)
}

func TestMockClient_ValidateSession_NilFunc(t *testing.T) {
	client := &MockClient{}

	identity, err := client.ValidateSession(context.Background(), "test-cookie")

	assert.NoError(t, err)
	assert.Equal(t, "test-user", identity.ID)
}

func TestMockClient_GetCookieName(t *testing.T) {
	client := NewMockClient()
	assert.Equal(t, "ory_kratos_session", client.GetCookieName())
}
