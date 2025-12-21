package handlers

import (
	"testing"

	"github.com/codoworks/codo-framework/core/http"
)

func TestGroupHandler_Prefix(t *testing.T) {
	h := &GroupHandler{}
	if got := h.Prefix(); got != "/api/v1/groups" {
		t.Errorf("Prefix() = %v, want %v", got, "/api/v1/groups")
	}
}

func TestGroupHandler_Scope(t *testing.T) {
	h := &GroupHandler{}
	if got := h.Scope(); got != http.ScopeProtected {
		t.Errorf("Scope() = %v, want %v", got, http.ScopeProtected)
	}
}

func TestGroupHandler_Middlewares(t *testing.T) {
	h := &GroupHandler{}
	if got := h.Middlewares(); got != nil {
		t.Errorf("Middlewares() = %v, want nil", got)
	}
}

func TestGroupHandler_Initialize(t *testing.T) {
	h := &GroupHandler{}
	if err := h.Initialize(); err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}
}

func TestGroupHandler_Implements_Handler(t *testing.T) {
	// Compile-time check that GroupHandler implements http.Handler
	var _ http.Handler = &GroupHandler{}
}
