package handlers

import (
	"testing"

	"github.com/codoworks/codo-framework/core/http"
)

func TestRegisterAll(t *testing.T) {
	// Clear any existing handlers
	http.ClearHandlers()

	// Register all handlers (with nil db client for unit test)
	RegisterAll(nil)

	// Verify handlers were registered
	count := http.HandlerCount()
	if count != 2 {
		t.Errorf("RegisterAll() registered %d handlers, want 2", count)
	}

	// Verify we have protected scope handlers
	protectedHandlers := http.GetHandlers(http.ScopeProtected)
	if len(protectedHandlers) != 2 {
		t.Errorf("GetHandlers(ScopeProtected) = %d, want 2", len(protectedHandlers))
	}

	// Cleanup
	http.ClearHandlers()
}

func TestRegisterContact(t *testing.T) {
	http.ClearHandlers()

	RegisterContact(nil)

	count := http.HandlerCount()
	if count != 1 {
		t.Errorf("RegisterContact() registered %d handlers, want 1", count)
	}

	handlers := http.AllHandlers()
	if len(handlers) != 1 {
		t.Fatalf("Expected 1 handler, got %d", len(handlers))
	}

	if handlers[0].Prefix() != "/api/v1/contacts" {
		t.Errorf("Handler prefix = %s, want /api/v1/contacts", handlers[0].Prefix())
	}

	http.ClearHandlers()
}

func TestRegisterGroup(t *testing.T) {
	http.ClearHandlers()

	RegisterGroup(nil)

	count := http.HandlerCount()
	if count != 1 {
		t.Errorf("RegisterGroup() registered %d handlers, want 1", count)
	}

	handlers := http.AllHandlers()
	if len(handlers) != 1 {
		t.Fatalf("Expected 1 handler, got %d", len(handlers))
	}

	if handlers[0].Prefix() != "/api/v1/groups" {
		t.Errorf("Handler prefix = %s, want /api/v1/groups", handlers[0].Prefix())
	}

	http.ClearHandlers()
}

func TestRegisterAll_HandlerPrefixes(t *testing.T) {
	http.ClearHandlers()
	defer http.ClearHandlers()

	RegisterAll(nil)

	handlers := http.AllHandlers()
	prefixes := make(map[string]bool)
	for _, h := range handlers {
		prefixes[h.Prefix()] = true
	}

	expectedPrefixes := []string{"/api/v1/contacts", "/api/v1/groups"}
	for _, prefix := range expectedPrefixes {
		if !prefixes[prefix] {
			t.Errorf("Expected handler with prefix %s to be registered", prefix)
		}
	}
}

func TestRegisterAll_AllProtectedScope(t *testing.T) {
	http.ClearHandlers()
	defer http.ClearHandlers()

	RegisterAll(nil)

	handlers := http.AllHandlers()
	for _, h := range handlers {
		if h.Scope() != http.ScopeProtected {
			t.Errorf("Handler %s has scope %v, want ScopeProtected", h.Prefix(), h.Scope())
		}
	}
}
