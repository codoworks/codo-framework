package errors

import (
	"errors"
	"os"
	"syscall"
)

// OS Error Predicates
//
// These predicates detect common OS-level errors and allow the error mapper
// to automatically map them to appropriate HTTP status codes.
//
// This means applications using the framework don't need to explicitly handle
// common file system errors - they'll automatically get meaningful responses:
//   - Permission denied → 403 Forbidden
//   - File not found → 404 Not Found
//   - File exists → 409 Conflict
//   - Timeout → 408 Request Timeout

// isPermissionDenied checks if an error is a permission denied error.
// Maps to HTTP 403 Forbidden.
//
// Handles:
//   - os.ErrPermission
//   - syscall.EACCES (permission denied)
//   - syscall.EPERM (operation not permitted)
func isPermissionDenied(err error) bool {
	return errors.Is(err, os.ErrPermission) ||
		errors.Is(err, syscall.EACCES) ||
		errors.Is(err, syscall.EPERM)
}

// isNotExist checks if an error indicates a resource doesn't exist.
// Maps to HTTP 404 Not Found.
//
// Handles:
//   - os.ErrNotExist
//   - syscall.ENOENT (no such file or directory)
func isNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist) ||
		errors.Is(err, syscall.ENOENT)
}

// isAlreadyExists checks if an error indicates a resource already exists.
// Maps to HTTP 409 Conflict.
//
// Handles:
//   - os.ErrExist
//   - syscall.EEXIST (file exists)
func isAlreadyExists(err error) bool {
	return errors.Is(err, os.ErrExist) ||
		errors.Is(err, syscall.EEXIST)
}

// isTimeout checks if an error is a timeout error.
// Maps to HTTP 408 Request Timeout.
//
// Handles:
//   - os.ErrDeadlineExceeded
//   - Any error implementing Timeout() bool that returns true
func isTimeout(err error) bool {
	// Check for os.ErrDeadlineExceeded
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	// Check for net.Error timeout interface
	type timeoutError interface {
		Timeout() bool
	}

	// Walk the error chain looking for a timeout error
	current := err
	for current != nil {
		if te, ok := current.(timeoutError); ok && te.Timeout() {
			return true
		}
		current = errors.Unwrap(current)
	}

	return false
}

// isClosed checks if an error indicates a closed resource.
// Maps to HTTP 503 Service Unavailable (resource temporarily unavailable).
//
// Handles:
//   - os.ErrClosed
func isClosed(err error) bool {
	return errors.Is(err, os.ErrClosed)
}

// isNoSpace checks if an error indicates insufficient disk space.
// Maps to HTTP 507 Insufficient Storage.
//
// Handles:
//   - syscall.ENOSPC (no space left on device)
func isNoSpace(err error) bool {
	return errors.Is(err, syscall.ENOSPC)
}
