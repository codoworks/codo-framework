package db

import (
	"time"

	"github.com/google/uuid"
)

// BeforeCreateHook is called before creating a record
type BeforeCreateHook interface {
	BeforeCreate() error
}

// AfterCreateHook is called after creating a record
type AfterCreateHook interface {
	AfterCreate() error
}

// BeforeUpdateHook is called before updating a record
type BeforeUpdateHook interface {
	BeforeUpdate() error
}

// AfterUpdateHook is called after updating a record
type AfterUpdateHook interface {
	AfterUpdate() error
}

// BeforeDeleteHook is called before deleting a record
type BeforeDeleteHook interface {
	BeforeDelete() error
}

// AfterDeleteHook is called after deleting a record
type AfterDeleteHook interface {
	AfterDelete() error
}

// BeforeSaveHook is called before create or update
type BeforeSaveHook interface {
	BeforeSave() error
}

// AfterSaveHook is called after create or update
type AfterSaveHook interface {
	AfterSave() error
}

// AfterFindHook is called after finding a record
type AfterFindHook interface {
	AfterFind() error
}

// ValidateHook is called to validate a record before saving
type ValidateHook interface {
	Validate() error
}

// ApplyBeforeCreate sets ID and timestamps for new records
func ApplyBeforeCreate(model *Model) {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
}

// ApplyBeforeCreateWithTime sets ID and timestamps with a specific time
func ApplyBeforeCreateWithTime(model *Model, t time.Time) {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	model.CreatedAt = t
	model.UpdatedAt = t
}

// ApplyBeforeUpdate sets UpdatedAt timestamp
func ApplyBeforeUpdate(model *Model) {
	model.UpdatedAt = time.Now()
}

// ApplyBeforeUpdateWithTime sets UpdatedAt with a specific time
func ApplyBeforeUpdateWithTime(model *Model, t time.Time) {
	model.UpdatedAt = t
}

// ApplyBeforeDelete sets DeletedAt timestamp for soft delete
func ApplyBeforeDelete(model *Model) {
	now := time.Now()
	model.DeletedAt = &now
}

// ApplyBeforeDeleteWithTime sets DeletedAt with a specific time
func ApplyBeforeDeleteWithTime(model *Model, t time.Time) {
	model.DeletedAt = &t
}

// RunBeforeCreateHooks runs all before create hooks on the model
func RunBeforeCreateHooks(model any) error {
	if hook, ok := model.(ValidateHook); ok {
		if err := hook.Validate(); err != nil {
			return err
		}
	}
	if hook, ok := model.(BeforeSaveHook); ok {
		if err := hook.BeforeSave(); err != nil {
			return err
		}
	}
	if hook, ok := model.(BeforeCreateHook); ok {
		if err := hook.BeforeCreate(); err != nil {
			return err
		}
	}
	return nil
}

// RunAfterCreateHooks runs all after create hooks on the model
func RunAfterCreateHooks(model any) error {
	if hook, ok := model.(AfterCreateHook); ok {
		if err := hook.AfterCreate(); err != nil {
			return err
		}
	}
	if hook, ok := model.(AfterSaveHook); ok {
		if err := hook.AfterSave(); err != nil {
			return err
		}
	}
	return nil
}

// RunBeforeUpdateHooks runs all before update hooks on the model
func RunBeforeUpdateHooks(model any) error {
	if hook, ok := model.(ValidateHook); ok {
		if err := hook.Validate(); err != nil {
			return err
		}
	}
	if hook, ok := model.(BeforeSaveHook); ok {
		if err := hook.BeforeSave(); err != nil {
			return err
		}
	}
	if hook, ok := model.(BeforeUpdateHook); ok {
		if err := hook.BeforeUpdate(); err != nil {
			return err
		}
	}
	return nil
}

// RunAfterUpdateHooks runs all after update hooks on the model
func RunAfterUpdateHooks(model any) error {
	if hook, ok := model.(AfterUpdateHook); ok {
		if err := hook.AfterUpdate(); err != nil {
			return err
		}
	}
	if hook, ok := model.(AfterSaveHook); ok {
		if err := hook.AfterSave(); err != nil {
			return err
		}
	}
	return nil
}

// RunBeforeDeleteHooks runs all before delete hooks on the model
func RunBeforeDeleteHooks(model any) error {
	if hook, ok := model.(BeforeDeleteHook); ok {
		if err := hook.BeforeDelete(); err != nil {
			return err
		}
	}
	return nil
}

// RunAfterDeleteHooks runs all after delete hooks on the model
func RunAfterDeleteHooks(model any) error {
	if hook, ok := model.(AfterDeleteHook); ok {
		if err := hook.AfterDelete(); err != nil {
			return err
		}
	}
	return nil
}

// RunAfterFindHooks runs all after find hooks on the model
func RunAfterFindHooks(model any) error {
	if hook, ok := model.(AfterFindHook); ok {
		if err := hook.AfterFind(); err != nil {
			return err
		}
	}
	return nil
}
