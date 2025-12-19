package db

import (
	"time"
)

// Modeler is the interface all models must implement
type Modeler interface {
	TableName() string
	PrimaryKey() string
}

// IDGetter is an optional interface for models to provide their ID
type IDGetter interface {
	GetID() string
}

// Model is the base struct for all models with common fields
type Model struct {
	ID        string     `db:"id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// PrimaryKey returns the primary key column name
func (m *Model) PrimaryKey() string {
	return "id"
}

// GetID returns the model's ID
func (m *Model) GetID() string {
	return m.ID
}

// SetID sets the model's ID
func (m *Model) SetID(id string) {
	m.ID = id
}

// IsNew returns true if the model has not been persisted
func (m *Model) IsNew() bool {
	return m.ID == ""
}

// IsPersisted returns true if the model has been persisted
func (m *Model) IsPersisted() bool {
	return m.ID != "" && !m.CreatedAt.IsZero()
}

// IsDeleted returns true if the model has been soft deleted
func (m *Model) IsDeleted() bool {
	return m.DeletedAt != nil
}

// MarkDeleted sets the DeletedAt timestamp
func (m *Model) MarkDeleted() {
	now := time.Now()
	m.DeletedAt = &now
}

// Restore clears the DeletedAt timestamp
func (m *Model) Restore() {
	m.DeletedAt = nil
}
