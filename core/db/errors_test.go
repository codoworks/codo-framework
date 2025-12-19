package db

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrNotFound", ErrNotFound, true},
		{"wrapped ErrNotFound", fmt.Errorf("wrapped: %w", ErrNotFound), true},
		{"different error", errors.New("other error"), false},
		{"nil error", nil, false},
		{"ErrDuplicateKey", ErrDuplicateKey, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDuplicateKey(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrDuplicateKey", ErrDuplicateKey, true},
		{"wrapped ErrDuplicateKey", fmt.Errorf("wrapped: %w", ErrDuplicateKey), true},
		{"different error", errors.New("other error"), false},
		{"nil error", nil, false},
		{"ErrNotFound", ErrNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDuplicateKey(tt.err); got != tt.want {
				t.Errorf("IsDuplicateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	if ErrNotFound.Error() != "record not found" {
		t.Errorf("ErrNotFound = %q, want 'record not found'", ErrNotFound.Error())
	}
	if ErrDuplicateKey.Error() != "duplicate key" {
		t.Errorf("ErrDuplicateKey = %q, want 'duplicate key'", ErrDuplicateKey.Error())
	}
	if ErrNotInitialized.Error() != "database not initialized" {
		t.Errorf("ErrNotInitialized = %q, want 'database not initialized'", ErrNotInitialized.Error())
	}
	if ErrInvalidModel.Error() != "invalid model" {
		t.Errorf("ErrInvalidModel = %q, want 'invalid model'", ErrInvalidModel.Error())
	}
	if ErrNotPersisted.Error() != "record not persisted" {
		t.Errorf("ErrNotPersisted = %q, want 'record not persisted'", ErrNotPersisted.Error())
	}
}
