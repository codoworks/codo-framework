package db

import (
	"errors"
	"testing"
	"time"
)

func TestApplyBeforeCreate(t *testing.T) {
	t.Run("sets ID and timestamps", func(t *testing.T) {
		m := &Model{}

		before := time.Now()
		ApplyBeforeCreate(m)
		after := time.Now()

		if m.ID == "" {
			t.Error("ApplyBeforeCreate should set ID")
		}
		if len(m.ID) != 36 { // UUID length
			t.Errorf("ID should be a UUID, got length %d", len(m.ID))
		}
		if m.CreatedAt.Before(before) || m.CreatedAt.After(after) {
			t.Error("ApplyBeforeCreate should set CreatedAt to current time")
		}
		if m.UpdatedAt.Before(before) || m.UpdatedAt.After(after) {
			t.Error("ApplyBeforeCreate should set UpdatedAt to current time")
		}
		if m.CreatedAt != m.UpdatedAt {
			t.Error("CreatedAt and UpdatedAt should be equal on create")
		}
	})

	t.Run("preserves existing ID", func(t *testing.T) {
		existingID := "existing-id-123"
		m := &Model{ID: existingID}

		ApplyBeforeCreate(m)

		if m.ID != existingID {
			t.Errorf("ApplyBeforeCreate should preserve existing ID, got %s", m.ID)
		}
	})
}

func TestApplyBeforeCreateWithTime(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	m := &Model{}

	ApplyBeforeCreateWithTime(m, fixedTime)

	if m.ID == "" {
		t.Error("should set ID")
	}
	if m.CreatedAt != fixedTime {
		t.Errorf("CreatedAt = %v, want %v", m.CreatedAt, fixedTime)
	}
	if m.UpdatedAt != fixedTime {
		t.Errorf("UpdatedAt = %v, want %v", m.UpdatedAt, fixedTime)
	}
}

func TestApplyBeforeUpdate(t *testing.T) {
	originalCreatedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	m := &Model{
		ID:        "test-id",
		CreatedAt: originalCreatedAt,
		UpdatedAt: originalCreatedAt,
	}

	before := time.Now()
	ApplyBeforeUpdate(m)
	after := time.Now()

	if m.CreatedAt != originalCreatedAt {
		t.Error("ApplyBeforeUpdate should not change CreatedAt")
	}
	if m.UpdatedAt.Before(before) || m.UpdatedAt.After(after) {
		t.Error("ApplyBeforeUpdate should set UpdatedAt to current time")
	}
	if m.ID != "test-id" {
		t.Error("ApplyBeforeUpdate should not change ID")
	}
}

func TestApplyBeforeUpdateWithTime(t *testing.T) {
	originalTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updateTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	m := &Model{
		ID:        "test-id",
		CreatedAt: originalTime,
		UpdatedAt: originalTime,
	}

	ApplyBeforeUpdateWithTime(m, updateTime)

	if m.UpdatedAt != updateTime {
		t.Errorf("UpdatedAt = %v, want %v", m.UpdatedAt, updateTime)
	}
	if m.CreatedAt != originalTime {
		t.Error("CreatedAt should not change")
	}
}

func TestApplyBeforeDelete(t *testing.T) {
	m := &Model{
		ID:        "test-id",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if m.DeletedAt != nil {
		t.Error("DeletedAt should be nil initially")
	}

	before := time.Now()
	ApplyBeforeDelete(m)
	after := time.Now()

	if m.DeletedAt == nil {
		t.Error("ApplyBeforeDelete should set DeletedAt")
	}
	if m.DeletedAt.Before(before) || m.DeletedAt.After(after) {
		t.Error("ApplyBeforeDelete should set DeletedAt to current time")
	}
}

func TestApplyBeforeDeleteWithTime(t *testing.T) {
	deleteTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	m := &Model{ID: "test-id"}

	ApplyBeforeDeleteWithTime(m, deleteTime)

	if m.DeletedAt == nil {
		t.Fatal("DeletedAt should be set")
	}
	if *m.DeletedAt != deleteTime {
		t.Errorf("DeletedAt = %v, want %v", *m.DeletedAt, deleteTime)
	}
}

// Test models for hook testing
type hookTestModel struct {
	Model
	validateCalled      bool
	beforeSaveCalled    bool
	afterSaveCalled     bool
	beforeCreateCalled  bool
	afterCreateCalled   bool
	beforeUpdateCalled  bool
	afterUpdateCalled   bool
	beforeDeleteCalled  bool
	afterDeleteCalled   bool
	afterFindCalled     bool
	shouldFailValidate  bool
	shouldFailBeforeSave bool
}

func (m *hookTestModel) TableName() string { return "hook_tests" }

func (m *hookTestModel) Validate() error {
	m.validateCalled = true
	if m.shouldFailValidate {
		return errors.New("validation failed")
	}
	return nil
}

func (m *hookTestModel) BeforeSave() error {
	m.beforeSaveCalled = true
	if m.shouldFailBeforeSave {
		return errors.New("before save failed")
	}
	return nil
}

func (m *hookTestModel) AfterSave() error {
	m.afterSaveCalled = true
	return nil
}

func (m *hookTestModel) BeforeCreate() error {
	m.beforeCreateCalled = true
	return nil
}

func (m *hookTestModel) AfterCreate() error {
	m.afterCreateCalled = true
	return nil
}

func (m *hookTestModel) BeforeUpdate() error {
	m.beforeUpdateCalled = true
	return nil
}

func (m *hookTestModel) AfterUpdate() error {
	m.afterUpdateCalled = true
	return nil
}

func (m *hookTestModel) BeforeDelete() error {
	m.beforeDeleteCalled = true
	return nil
}

func (m *hookTestModel) AfterDelete() error {
	m.afterDeleteCalled = true
	return nil
}

func (m *hookTestModel) AfterFind() error {
	m.afterFindCalled = true
	return nil
}

func TestRunBeforeCreateHooks(t *testing.T) {
	t.Run("calls all hooks in order", func(t *testing.T) {
		m := &hookTestModel{}

		err := RunBeforeCreateHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !m.validateCalled {
			t.Error("Validate should be called")
		}
		if !m.beforeSaveCalled {
			t.Error("BeforeSave should be called")
		}
		if !m.beforeCreateCalled {
			t.Error("BeforeCreate should be called")
		}
	})

	t.Run("stops on validation error", func(t *testing.T) {
		m := &hookTestModel{shouldFailValidate: true}

		err := RunBeforeCreateHooks(m)

		if err == nil {
			t.Error("expected validation error")
		}
		if !m.validateCalled {
			t.Error("Validate should be called")
		}
		if m.beforeSaveCalled {
			t.Error("BeforeSave should not be called after validation fails")
		}
	})

	t.Run("stops on BeforeSave error", func(t *testing.T) {
		m := &hookTestModel{shouldFailBeforeSave: true}

		err := RunBeforeCreateHooks(m)

		if err == nil {
			t.Error("expected before save error")
		}
		if !m.beforeSaveCalled {
			t.Error("BeforeSave should be called")
		}
		if m.beforeCreateCalled {
			t.Error("BeforeCreate should not be called after BeforeSave fails")
		}
	})

	t.Run("works with model without hooks", func(t *testing.T) {
		m := &Model{}

		err := RunBeforeCreateHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRunAfterCreateHooks(t *testing.T) {
	t.Run("calls all hooks", func(t *testing.T) {
		m := &hookTestModel{}

		err := RunAfterCreateHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !m.afterCreateCalled {
			t.Error("AfterCreate should be called")
		}
		if !m.afterSaveCalled {
			t.Error("AfterSave should be called")
		}
	})

	t.Run("works with model without hooks", func(t *testing.T) {
		m := &Model{}

		err := RunAfterCreateHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRunBeforeUpdateHooks(t *testing.T) {
	t.Run("calls all hooks in order", func(t *testing.T) {
		m := &hookTestModel{}

		err := RunBeforeUpdateHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !m.validateCalled {
			t.Error("Validate should be called")
		}
		if !m.beforeSaveCalled {
			t.Error("BeforeSave should be called")
		}
		if !m.beforeUpdateCalled {
			t.Error("BeforeUpdate should be called")
		}
	})

	t.Run("stops on validation error", func(t *testing.T) {
		m := &hookTestModel{shouldFailValidate: true}

		err := RunBeforeUpdateHooks(m)

		if err == nil {
			t.Error("expected validation error")
		}
	})
}

func TestRunAfterUpdateHooks(t *testing.T) {
	t.Run("calls all hooks", func(t *testing.T) {
		m := &hookTestModel{}

		err := RunAfterUpdateHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !m.afterUpdateCalled {
			t.Error("AfterUpdate should be called")
		}
		if !m.afterSaveCalled {
			t.Error("AfterSave should be called")
		}
	})
}

func TestRunBeforeDeleteHooks(t *testing.T) {
	t.Run("calls hook", func(t *testing.T) {
		m := &hookTestModel{}

		err := RunBeforeDeleteHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !m.beforeDeleteCalled {
			t.Error("BeforeDelete should be called")
		}
	})

	t.Run("works with model without hooks", func(t *testing.T) {
		m := &Model{}

		err := RunBeforeDeleteHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRunAfterDeleteHooks(t *testing.T) {
	t.Run("calls hook", func(t *testing.T) {
		m := &hookTestModel{}

		err := RunAfterDeleteHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !m.afterDeleteCalled {
			t.Error("AfterDelete should be called")
		}
	})
}

func TestRunAfterFindHooks(t *testing.T) {
	t.Run("calls hook", func(t *testing.T) {
		m := &hookTestModel{}

		err := RunAfterFindHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !m.afterFindCalled {
			t.Error("AfterFind should be called")
		}
	})

	t.Run("works with model without hooks", func(t *testing.T) {
		m := &Model{}

		err := RunAfterFindHooks(m)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// Test error handling in hooks
type errorHookModel struct {
	Model
}

func (m *errorHookModel) TableName() string { return "errors" }

func (m *errorHookModel) AfterCreate() error {
	return errors.New("after create error")
}

func (m *errorHookModel) AfterUpdate() error {
	return errors.New("after update error")
}

func (m *errorHookModel) BeforeDelete() error {
	return errors.New("before delete error")
}

func (m *errorHookModel) AfterDelete() error {
	return errors.New("after delete error")
}

func (m *errorHookModel) AfterFind() error {
	return errors.New("after find error")
}

func TestHookErrors(t *testing.T) {
	m := &errorHookModel{}

	if err := RunAfterCreateHooks(m); err == nil {
		t.Error("AfterCreate error should propagate")
	}

	if err := RunAfterUpdateHooks(m); err == nil {
		t.Error("AfterUpdate error should propagate")
	}

	if err := RunBeforeDeleteHooks(m); err == nil {
		t.Error("BeforeDelete error should propagate")
	}

	if err := RunAfterDeleteHooks(m); err == nil {
		t.Error("AfterDelete error should propagate")
	}

	if err := RunAfterFindHooks(m); err == nil {
		t.Error("AfterFind error should propagate")
	}
}

// Test that hooks are interface assertions
func TestHookInterfaces(t *testing.T) {
	m := &hookTestModel{}

	// These should compile if the interfaces are correct
	var _ ValidateHook = m
	var _ BeforeSaveHook = m
	var _ AfterSaveHook = m
	var _ BeforeCreateHook = m
	var _ AfterCreateHook = m
	var _ BeforeUpdateHook = m
	var _ AfterUpdateHook = m
	var _ BeforeDeleteHook = m
	var _ AfterDeleteHook = m
	var _ AfterFindHook = m
}

// Models for testing individual hook errors
type beforeCreateErrorModel struct {
	Model
}

func (m *beforeCreateErrorModel) TableName() string { return "before_create_error" }
func (m *beforeCreateErrorModel) BeforeCreate() error {
	return errors.New("before create failed")
}

type afterSaveErrorModel struct {
	Model
}

func (m *afterSaveErrorModel) TableName() string { return "after_save_error" }
func (m *afterSaveErrorModel) AfterSave() error {
	return errors.New("after save failed")
}

type beforeUpdateErrorModel struct {
	Model
}

func (m *beforeUpdateErrorModel) TableName() string { return "before_update_error" }
func (m *beforeUpdateErrorModel) BeforeUpdate() error {
	return errors.New("before update failed")
}

type beforeSaveErrorModel struct {
	Model
}

func (m *beforeSaveErrorModel) TableName() string { return "before_save_error" }
func (m *beforeSaveErrorModel) BeforeSave() error {
	return errors.New("before save failed")
}

func TestRunBeforeCreateHooks_BeforeCreateError(t *testing.T) {
	m := &beforeCreateErrorModel{}

	err := RunBeforeCreateHooks(m)

	if err == nil {
		t.Error("expected BeforeCreate error")
	}
	if err != nil && err.Error() != "before create failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunAfterCreateHooks_AfterSaveError(t *testing.T) {
	m := &afterSaveErrorModel{}

	err := RunAfterCreateHooks(m)

	if err == nil {
		t.Error("expected AfterSave error")
	}
	if err != nil && err.Error() != "after save failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunBeforeUpdateHooks_BeforeSaveError(t *testing.T) {
	m := &beforeSaveErrorModel{}

	err := RunBeforeUpdateHooks(m)

	if err == nil {
		t.Error("expected BeforeSave error")
	}
	if err != nil && err.Error() != "before save failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunBeforeUpdateHooks_BeforeUpdateError(t *testing.T) {
	m := &beforeUpdateErrorModel{}

	err := RunBeforeUpdateHooks(m)

	if err == nil {
		t.Error("expected BeforeUpdate error")
	}
	if err != nil && err.Error() != "before update failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunAfterUpdateHooks_AfterSaveError(t *testing.T) {
	m := &afterSaveErrorModel{}

	err := RunAfterUpdateHooks(m)

	if err == nil {
		t.Error("expected AfterSave error")
	}
	if err != nil && err.Error() != "after save failed" {
		t.Errorf("unexpected error: %v", err)
	}
}
