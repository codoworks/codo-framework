package handlers

import (
	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/core/http"
)

// RegisterAll registers all application handlers with the HTTP registry.
// This should be called during application bootstrap after initializing
// the database client.
//
// Example usage:
//
//	dbClient, err := db.NewClient(cfg.Database)
//	if err != nil {
//	    return err
//	}
//	handlers.RegisterAll(dbClient)
func RegisterAll(dbClient *db.Client) {
	http.RegisterHandler(NewContactHandler(dbClient))
	http.RegisterHandler(NewGroupHandler(dbClient))
}

// RegisterContact registers only the contact handler.
// Useful for testing or when only contact functionality is needed.
func RegisterContact(dbClient *db.Client) {
	http.RegisterHandler(NewContactHandler(dbClient))
}

// RegisterGroup registers only the group handler.
// Useful for testing or when only group functionality is needed.
func RegisterGroup(dbClient *db.Client) {
	http.RegisterHandler(NewGroupHandler(dbClient))
}
