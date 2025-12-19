package kratos

import (
	"context"

	"github.com/codoworks/codo-framework/core/auth"
)

// MockClient is a mock Kratos client for testing
type MockClient struct {
	ValidateFunc func(ctx context.Context, cookie string) (*auth.Identity, error)
	HealthFunc   func() error
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		ValidateFunc: func(ctx context.Context, cookie string) (*auth.Identity, error) {
			return &auth.Identity{ID: "test-user"}, nil
		},
		HealthFunc: func() error {
			return nil
		},
	}
}

// Name returns the client name
func (m *MockClient) Name() string { return "kratos" }

// Initialize sets up the mock client
func (m *MockClient) Initialize(cfg any) error { return nil }

// Shutdown closes the mock client
func (m *MockClient) Shutdown() error { return nil }

// Health checks if the mock Kratos is healthy
func (m *MockClient) Health() error {
	if m.HealthFunc != nil {
		return m.HealthFunc()
	}
	return nil
}

// ValidateSession validates a session using the mock function
func (m *MockClient) ValidateSession(ctx context.Context, cookie string) (*auth.Identity, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx, cookie)
	}
	return &auth.Identity{ID: "test-user"}, nil
}

// GetCookieName returns the session cookie name
func (m *MockClient) GetCookieName() string {
	return "ory_kratos_session"
}
