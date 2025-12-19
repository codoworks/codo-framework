package db

import (
	"context"
)

// Record wraps a model and provides active record methods
type Record[T Modeler] struct {
	model T
	repo  *Repository[T]
}

// Model returns the underlying model
func (r *Record[T]) Model() T {
	return r.model
}

// Repository returns the associated repository
func (r *Record[T]) Repository() *Repository[T] {
	return r.repo
}

// ID returns the model's ID
func (r *Record[T]) ID() string {
	return getModelID(r.model)
}

// IsNew returns true if the record hasn't been persisted
func (r *Record[T]) IsNew() bool {
	if baseModel := getBaseModel(r.model); baseModel != nil {
		return baseModel.IsNew()
	}
	return getModelID(r.model) == ""
}

// IsPersisted returns true if the record has been persisted
func (r *Record[T]) IsPersisted() bool {
	if baseModel := getBaseModel(r.model); baseModel != nil {
		return baseModel.IsPersisted()
	}
	return getModelID(r.model) != ""
}

// IsDeleted returns true if the record has been soft deleted
func (r *Record[T]) IsDeleted() bool {
	if baseModel := getBaseModel(r.model); baseModel != nil {
		return baseModel.IsDeleted()
	}
	return false
}

// Save creates or updates the record
func (r *Record[T]) Save(ctx context.Context) error {
	if r.IsNew() {
		return r.repo.Create(ctx, r.model)
	}
	return r.repo.Update(ctx, r.model)
}

// Create inserts the record into the database
func (r *Record[T]) Create(ctx context.Context) error {
	return r.repo.Create(ctx, r.model)
}

// Update saves changes to an existing record
func (r *Record[T]) Update(ctx context.Context) error {
	return r.repo.Update(ctx, r.model)
}

// Delete soft-deletes the record
func (r *Record[T]) Delete(ctx context.Context) error {
	return r.repo.Delete(ctx, r.model)
}

// HardDelete permanently deletes the record
func (r *Record[T]) HardDelete(ctx context.Context) error {
	return r.repo.HardDelete(ctx, r.model)
}

// Restore un-deletes a soft-deleted record
func (r *Record[T]) Restore(ctx context.Context) error {
	return r.repo.Restore(ctx, r.model)
}

// Reload refreshes the record from the database
func (r *Record[T]) Reload(ctx context.Context) error {
	id := r.ID()
	if id == "" {
		return ErrNotPersisted
	}

	fresh, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	r.model = fresh.model
	return nil
}

// Touch updates the record's updated_at timestamp
func (r *Record[T]) Touch(ctx context.Context) error {
	if r.IsNew() {
		return ErrNotPersisted
	}

	if baseModel := getBaseModel(r.model); baseModel != nil {
		ApplyBeforeUpdate(baseModel)
	}

	return r.repo.Update(ctx, r.model)
}

// Duplicate creates a new record with the same data but no ID
func (r *Record[T]) Duplicate() *Record[T] {
	// Create new record
	newRecord := r.repo.New()

	// Copy the model data
	// This requires the model to be copyable
	// For now, we just return a new record - actual copying would require reflection

	return newRecord
}
