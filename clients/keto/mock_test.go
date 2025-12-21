package keto

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMockClient(t *testing.T) {
	client := NewMockClient()

	assert.NotNil(t, client)
	assert.NotNil(t, client.CheckFunc)
	assert.NotNil(t, client.HealthFunc)
}

func TestMockClient_Name(t *testing.T) {
	client := NewMockClient()
	assert.Equal(t, "keto", client.Name())
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

func TestMockClient_CheckPermission(t *testing.T) {
	client := NewMockClient()

	allowed, err := client.CheckPermission(context.Background(), "user", "viewer", "docs", "doc1")

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestMockClient_CheckPermission_WithFunc(t *testing.T) {
	client := NewMockClient()
	client.CheckFunc = func(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
		if relation == "admin" {
			return false, nil
		}
		return true, nil
	}

	// Should be allowed for viewer
	allowed, err := client.CheckPermission(context.Background(), "user", "viewer", "docs", "doc1")
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Should be denied for admin
	allowed, err = client.CheckPermission(context.Background(), "user", "admin", "docs", "doc1")
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestMockClient_CheckPermission_Error(t *testing.T) {
	client := NewMockClient()
	client.CheckFunc = func(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
		return false, fmt.Errorf("permission check error")
	}

	allowed, err := client.CheckPermission(context.Background(), "user", "viewer", "docs", "doc1")

	assert.Error(t, err)
	assert.False(t, allowed)
}

func TestMockClient_CheckPermission_NilFunc(t *testing.T) {
	client := &MockClient{}

	allowed, err := client.CheckPermission(context.Background(), "user", "viewer", "docs", "doc1")

	assert.NoError(t, err)
	assert.True(t, allowed)
}
