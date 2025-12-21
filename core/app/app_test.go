package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/core/metadata"
)

// mockApp implements App interface without MetadataProvider
type mockApp struct{}

func (m *mockApp) Server() *http.Server             { return nil }
func (m *mockApp) Shutdown(ctx context.Context) error { return nil }

// mockAppWithMetadata implements both App and MetadataProvider interfaces
type mockAppWithMetadata struct{}

func (m *mockAppWithMetadata) Server() *http.Server             { return nil }
func (m *mockAppWithMetadata) Shutdown(ctx context.Context) error { return nil }
func (m *mockAppWithMetadata) Metadata() metadata.Metadata {
	return metadata.Info{
		AppName:  "myapp",
		AppShort: "My App",
		AppLong:  "My Application",
	}
}

func TestGetMetadata_WithProvider(t *testing.T) {
	app := &mockAppWithMetadata{}
	meta := GetMetadata(app)

	assert.Equal(t, "myapp", meta.Name())
	assert.Equal(t, "My App", meta.Short())
	assert.Equal(t, "My Application", meta.Long())
}

func TestGetMetadata_WithoutProvider(t *testing.T) {
	app := &mockApp{}
	meta := GetMetadata(app)

	// Should return framework defaults
	assert.Equal(t, "codo", meta.Name())
	assert.Equal(t, "Codo Framework CLI", meta.Short())
	assert.Equal(t, "Codo Framework - A production-ready Go backend framework", meta.Long())
}
