// Package app provides application initialization and bootstrap functionality for consumer applications.
package app

import (
	"context"

	"github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/core/metadata"
)

// App is the legacy interface for initialized applications.
// Deprecated: Use BaseApp, HTTPApp, SingleRouterApp, or DaemonApp instead.
// This interface is kept for backward compatibility but should not be used in new code.
type App interface {
	// Server returns the HTTP server instance.
	// Deprecated: Use HTTPApp.Server() instead.
	Server() *http.Server

	// Shutdown gracefully shuts down the application.
	// Deprecated: Use BaseApp.Shutdown() instead.
	Shutdown(ctx context.Context) error
}

// MetadataProvider is an optional interface that apps can implement
// to provide custom CLI metadata.
// If not implemented, the framework will use default metadata.
type MetadataProvider interface {
	Metadata() metadata.Metadata
}

// GetMetadata returns the app's metadata if provided, otherwise returns framework defaults.
// This allows apps to optionally override CLI metadata (name, short description, long description).
func GetMetadata(application BaseApp) metadata.Metadata {
	if provider, ok := application.(MetadataProvider); ok {
		return provider.Metadata()
	}
	return metadata.Default()
}
