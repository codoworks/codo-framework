package services

import (
	"testing"
)

func TestNewContactService(t *testing.T) {
	// NewContactService with nil client should not panic
	// (actual database operations would fail, but creation should work)
	service := NewContactService(nil)
	if service == nil {
		t.Error("NewContactService should return non-nil service")
	}
}

func TestContactService_Structure(t *testing.T) {
	service := NewContactService(nil)

	// Verify service has the expected structure
	if service.repo == nil {
		t.Error("service.repo should be initialized")
	}
	if service.groupRepo == nil {
		t.Error("service.groupRepo should be initialized")
	}
}
