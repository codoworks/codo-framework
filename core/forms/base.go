package forms

import (
	"time"
)

// FormBase provides common fields for response forms
type FormBase struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewFormBase creates a FormBase from model fields
func NewFormBase(id string, createdAt, updatedAt time.Time) FormBase {
	return FormBase{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// FormWithID provides just an ID field for minimal responses
type FormWithID struct {
	ID string `json:"id"`
}

// NewFormWithID creates a FormWithID
func NewFormWithID(id string) FormWithID {
	return FormWithID{ID: id}
}

// TimestampFields provides only timestamp fields
type TimestampFields struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTimestampFields creates TimestampFields
func NewTimestampFields(createdAt, updatedAt time.Time) TimestampFields {
	return TimestampFields{
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// DeletedFormBase includes soft delete timestamp
type DeletedFormBase struct {
	FormBase
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewDeletedFormBase creates a DeletedFormBase
func NewDeletedFormBase(id string, createdAt, updatedAt time.Time, deletedAt *time.Time) DeletedFormBase {
	return DeletedFormBase{
		FormBase:  NewFormBase(id, createdAt, updatedAt),
		DeletedAt: deletedAt,
	}
}
