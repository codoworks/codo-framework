package services

import (
	"testing"
)

func TestNewGroupService(t *testing.T) {
	// NewGroupService with nil client should not panic
	service := NewGroupService(nil)
	if service == nil {
		t.Error("NewGroupService should return non-nil service")
	}
}

func TestGroupService_Structure(t *testing.T) {
	service := NewGroupService(nil)

	// Verify service has the expected structure
	if service.repo == nil {
		t.Error("service.repo should be initialized")
	}
	if service.contactRepo == nil {
		t.Error("service.contactRepo should be initialized")
	}
}

func TestGroupWithCount_Structure(t *testing.T) {
	gwc := GroupWithCount{
		Group: nil,
		Count: 42,
	}

	if gwc.Count != 42 {
		t.Errorf("Count = %v, want %v", gwc.Count, 42)
	}
}
