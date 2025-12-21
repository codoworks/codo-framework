// Package app provides application initialization hooks for consumer applications.
package app

import (
	"context"
	"fmt"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/core/metadata"
)

// App represents an initialized application.
// Consumer applications should implement this interface.
type App interface {
	// Server returns the HTTP server instance.
	Server() *http.Server

	// Shutdown gracefully shuts down the application.
	// This should handle server shutdown, client cleanup, and any other teardown logic.
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
func GetMetadata(application App) metadata.Metadata {
	if provider, ok := application.(MetadataProvider); ok {
		return provider.Metadata()
	}
	return metadata.Default()
}

// Initializer is a function that initializes an application.
// Consumer applications should implement this function to:
// 1. Register and initialize all clients
// 2. Run database migrations
// 3. Register HTTP handlers with the server
// 4. Return an App instance that can be started and shut down
type Initializer func(cfg *config.Config) (App, error)

var registeredInitializer Initializer

// RegisterInitializer sets the app initialization function.
// Consumer apps should call this in their init() function.
//
// Example:
//
//	func init() {
//	    app.RegisterInitializer(myapp.Initialize)
//	}
func RegisterInitializer(fn Initializer) {
	registeredInitializer = fn
}

// Initialize calls the registered initializer function.
// This is called by the framework's serve command.
// Returns an error if no initializer has been registered.
func Initialize(cfg *config.Config) (App, error) {
	if registeredInitializer == nil {
		return nil, fmt.Errorf("no app initializer registered")
	}
	return registeredInitializer(cfg)
}
