package keto

import (
	"context"
)

// MockClient is a mock Keto client for testing
type MockClient struct {
	CheckFunc  func(ctx context.Context, subject, relation, namespace, object string) (bool, error)
	HealthFunc func() error
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		CheckFunc: func(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
			return true, nil
		},
		HealthFunc: func() error {
			return nil
		},
	}
}

// Name returns the client name
func (m *MockClient) Name() string { return "keto" }

// Initialize sets up the mock client
func (m *MockClient) Initialize(cfg any) error { return nil }

// Shutdown closes the mock client
func (m *MockClient) Shutdown() error { return nil }

// Health checks if the mock Keto is healthy
func (m *MockClient) Health() error {
	if m.HealthFunc != nil {
		return m.HealthFunc()
	}
	return nil
}

// CheckPermission checks permission using the mock function
func (m *MockClient) CheckPermission(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx, subject, relation, namespace, object)
	}
	return true, nil
}
