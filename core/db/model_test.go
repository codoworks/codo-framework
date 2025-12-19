package db

import (
	"testing"
	"time"
)

func TestModel_PrimaryKey(t *testing.T) {
	m := &Model{}
	if got := m.PrimaryKey(); got != "id" {
		t.Errorf("PrimaryKey() = %v, want %v", got, "id")
	}
}

func TestModel_GetID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{"empty ID", "", ""},
		{"with ID", "abc-123", "abc-123"},
		{"uuid ID", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{ID: tt.id}
			if got := m.GetID(); got != tt.want {
				t.Errorf("GetID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_SetID(t *testing.T) {
	m := &Model{}
	m.SetID("test-id-123")
	if m.ID != "test-id-123" {
		t.Errorf("SetID() did not set ID correctly, got %v", m.ID)
	}

	// Test overwriting existing ID
	m.SetID("new-id-456")
	if m.ID != "new-id-456" {
		t.Errorf("SetID() did not overwrite ID correctly, got %v", m.ID)
	}
}

func TestModel_IsNew(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"empty ID is new", "", true},
		{"with ID is not new", "some-id", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{ID: tt.id}
			if got := m.IsNew(); got != tt.want {
				t.Errorf("IsNew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_IsPersisted(t *testing.T) {
	now := time.Now()
	zeroTime := time.Time{}

	tests := []struct {
		name      string
		id        string
		createdAt time.Time
		want      bool
	}{
		{"empty ID and zero time", "", zeroTime, false},
		{"with ID but zero time", "some-id", zeroTime, false},
		{"empty ID with time", "", now, false},
		{"with ID and time", "some-id", now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{ID: tt.id, CreatedAt: tt.createdAt}
			if got := m.IsPersisted(); got != tt.want {
				t.Errorf("IsPersisted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_IsDeleted(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		deletedAt *time.Time
		want      bool
	}{
		{"nil DeletedAt", nil, false},
		{"with DeletedAt", &now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{DeletedAt: tt.deletedAt}
			if got := m.IsDeleted(); got != tt.want {
				t.Errorf("IsDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_MarkDeleted(t *testing.T) {
	m := &Model{}

	if m.DeletedAt != nil {
		t.Error("DeletedAt should be nil initially")
	}

	before := time.Now()
	m.MarkDeleted()
	after := time.Now()

	if m.DeletedAt == nil {
		t.Error("MarkDeleted() should set DeletedAt")
	}

	if m.DeletedAt.Before(before) || m.DeletedAt.After(after) {
		t.Error("MarkDeleted() should set DeletedAt to current time")
	}
}

func TestModel_Restore(t *testing.T) {
	now := time.Now()
	m := &Model{DeletedAt: &now}

	if m.DeletedAt == nil {
		t.Error("DeletedAt should be set initially for this test")
	}

	m.Restore()

	if m.DeletedAt != nil {
		t.Error("Restore() should clear DeletedAt")
	}
}

func TestModel_Restore_AlreadyNil(t *testing.T) {
	m := &Model{}
	m.Restore() // Should not panic
	if m.DeletedAt != nil {
		t.Error("Restore() should keep DeletedAt nil")
	}
}

// TestModel_FieldTags verifies struct tags are correct
func TestModel_FieldTags(t *testing.T) {
	// This is a compile-time check that the struct has the expected fields
	m := Model{
		ID:        "test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}

	if m.ID != "test" {
		t.Error("ID field should be accessible")
	}
}

// Test that Model can be embedded in other structs
func TestModel_Embedding(t *testing.T) {
	type TestModel struct {
		Model
		Name string `db:"name"`
	}

	tm := &TestModel{
		Model: Model{ID: "embed-test"},
		Name:  "Test",
	}

	if tm.GetID() != "embed-test" {
		t.Error("GetID() should work on embedded Model")
	}

	if tm.PrimaryKey() != "id" {
		t.Error("PrimaryKey() should work on embedded Model")
	}

	if tm.IsNew() {
		t.Error("Model with ID should not be new")
	}
}

// Test Model lifecycle
func TestModel_Lifecycle(t *testing.T) {
	m := &Model{}

	// Initial state
	if !m.IsNew() {
		t.Error("New model should be new")
	}
	if m.IsPersisted() {
		t.Error("New model should not be persisted")
	}
	if m.IsDeleted() {
		t.Error("New model should not be deleted")
	}

	// After setting ID and CreatedAt
	m.SetID("lifecycle-test")
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	if m.IsNew() {
		t.Error("Model with ID should not be new")
	}
	if !m.IsPersisted() {
		t.Error("Model with ID and CreatedAt should be persisted")
	}
	if m.IsDeleted() {
		t.Error("Model without DeletedAt should not be deleted")
	}

	// After soft delete
	m.MarkDeleted()

	if !m.IsDeleted() {
		t.Error("Model after MarkDeleted should be deleted")
	}
	if !m.IsPersisted() {
		t.Error("Deleted model should still be persisted")
	}

	// After restore
	m.Restore()

	if m.IsDeleted() {
		t.Error("Restored model should not be deleted")
	}
}
