package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/metadata"
)

// mockApp implements BaseApp interface without MetadataProvider
type mockApp struct{}

func (m *mockApp) Config() *config.Config            { return nil }
func (m *mockApp) Shutdown(ctx context.Context) error { return nil }
func (m *mockApp) Mode() AppMode                      { return HTTPServer }

// mockAppWithMetadata implements both BaseApp and MetadataProvider interfaces
type mockAppWithMetadata struct{}

func (m *mockAppWithMetadata) Config() *config.Config            { return nil }
func (m *mockAppWithMetadata) Shutdown(ctx context.Context) error { return nil }
func (m *mockAppWithMetadata) Mode() AppMode                      { return HTTPServer }
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
