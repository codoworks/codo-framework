package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

// IsConstraintViolation checks if the error indicates a unique constraint violation
// from the underlying database driver. Supports MySQL, PostgreSQL, and SQLite.
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// MySQL: 1062 "Duplicate entry"
	// PostgreSQL: 23505 "unique_violation"
	// SQLite: "UNIQUE constraint failed"
	return strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "1062") ||
		strings.Contains(errStr, "23505")
}

// WrapDBError wraps database errors with appropriate sentinel errors.
// It detects constraint violations and no-rows errors, returning the
// corresponding framework error sentinels.
func WrapDBError(err error, op string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if IsConstraintViolation(err) {
		return fmt.Errorf("%w: %v", ErrDuplicateKey, err)
	}
	return fmt.Errorf("%s failed: %w", op, err)
}
