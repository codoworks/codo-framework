package db

import (
	"errors"
)

// Common database errors
var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")

	// ErrDuplicateKey is returned when a unique constraint is violated
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrNotInitialized is returned when the database is not initialized
	ErrNotInitialized = errors.New("database not initialized")

	// ErrInvalidModel is returned when the model is invalid
	ErrInvalidModel = errors.New("invalid model")

	// ErrNotPersisted is returned when an operation requires a persisted record
	ErrNotPersisted = errors.New("record not persisted")
)

// IsNotFound returns true if the error is ErrNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDuplicateKey returns true if the error is ErrDuplicateKey
func IsDuplicateKey(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}
