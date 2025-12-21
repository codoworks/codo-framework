package handlers

import (
	"testing"

	"github.com/codoworks/codo-framework/core/http"
)

func TestContactHandler_Prefix(t *testing.T) {
	h := &ContactHandler{}
	if got := h.Prefix(); got != "/api/v1/contacts" {
		t.Errorf("Prefix() = %v, want %v", got, "/api/v1/contacts")
	}
}

func TestContactHandler_Scope(t *testing.T) {
	h := &ContactHandler{}
	if got := h.Scope(); got != http.ScopeProtected {
		t.Errorf("Scope() = %v, want %v", got, http.ScopeProtected)
	}
}

func TestContactHandler_Middlewares(t *testing.T) {
	h := &ContactHandler{}
	if got := h.Middlewares(); got != nil {
		t.Errorf("Middlewares() = %v, want nil", got)
	}
}

func TestContactHandler_Initialize(t *testing.T) {
	h := &ContactHandler{}
	if err := h.Initialize(); err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}
}

func TestContactHandler_Implements_Handler(t *testing.T) {
	// Compile-time check that ContactHandler implements http.Handler
	var _ http.Handler = &ContactHandler{}
}
