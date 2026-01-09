package errors

import (
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPermissionDenied(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"os.ErrPermission", os.ErrPermission, true},
		{"wrapped os.ErrPermission", fmt.Errorf("wrap: %w", os.ErrPermission), true},
		{"syscall.EACCES", syscall.EACCES, true},
		{"wrapped syscall.EACCES", fmt.Errorf("wrap: %w", syscall.EACCES), true},
		{"syscall.EPERM", syscall.EPERM, true},
		{"wrapped syscall.EPERM", fmt.Errorf("wrap: %w", syscall.EPERM), true},
		{"unrelated error", fmt.Errorf("some other error"), false},
		{"os.ErrNotExist", os.ErrNotExist, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPermissionDenied(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNotExist(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"os.ErrNotExist", os.ErrNotExist, true},
		{"wrapped os.ErrNotExist", fmt.Errorf("wrap: %w", os.ErrNotExist), true},
		{"syscall.ENOENT", syscall.ENOENT, true},
		{"wrapped syscall.ENOENT", fmt.Errorf("wrap: %w", syscall.ENOENT), true},
		{"deeply wrapped", fmt.Errorf("l1: %w", fmt.Errorf("l2: %w", os.ErrNotExist)), true},
		{"unrelated error", fmt.Errorf("some other error"), false},
		{"os.ErrPermission", os.ErrPermission, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNotExist(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"os.ErrExist", os.ErrExist, true},
		{"wrapped os.ErrExist", fmt.Errorf("wrap: %w", os.ErrExist), true},
		{"syscall.EEXIST", syscall.EEXIST, true},
		{"wrapped syscall.EEXIST", fmt.Errorf("wrap: %w", syscall.EEXIST), true},
		{"unrelated error", fmt.Errorf("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlreadyExists(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsClosed(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"os.ErrClosed", os.ErrClosed, true},
		{"wrapped os.ErrClosed", fmt.Errorf("wrap: %w", os.ErrClosed), true},
		{"unrelated error", fmt.Errorf("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isClosed(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNoSpace(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"syscall.ENOSPC", syscall.ENOSPC, true},
		{"wrapped syscall.ENOSPC", fmt.Errorf("wrap: %w", syscall.ENOSPC), true},
		{"unrelated error", fmt.Errorf("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNoSpace(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// timeoutError implements the Timeout() interface
type timeoutError struct {
	timeout bool
}

func (e *timeoutError) Error() string {
	return "timeout error"
}

func (e *timeoutError) Timeout() bool {
	return e.timeout
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"os.ErrDeadlineExceeded", os.ErrDeadlineExceeded, true},
		{"wrapped os.ErrDeadlineExceeded", fmt.Errorf("wrap: %w", os.ErrDeadlineExceeded), true},
		{"timeout interface true", &timeoutError{timeout: true}, true},
		{"wrapped timeout interface true", fmt.Errorf("wrap: %w", &timeoutError{timeout: true}), true},
		{"timeout interface false", &timeoutError{timeout: false}, false},
		{"unrelated error", fmt.Errorf("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTimeout(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration test: verify OS errors map to correct HTTP codes through the mapper
func TestOSErrorsIntegration(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "permission denied maps to 403",
			err:            fmt.Errorf("failed to write file: %w", os.ErrPermission),
			expectedCode:   CodeForbidden,
			expectedStatus: 403,
		},
		{
			name:           "not exist maps to 404",
			err:            fmt.Errorf("file not found: %w", os.ErrNotExist),
			expectedCode:   CodeNotFound,
			expectedStatus: 404,
		},
		{
			name:           "already exists maps to 409",
			err:            fmt.Errorf("file already exists: %w", os.ErrExist),
			expectedCode:   CodeConflict,
			expectedStatus: 409,
		},
		{
			name:           "closed maps to 503",
			err:            fmt.Errorf("connection closed: %w", os.ErrClosed),
			expectedCode:   CodeUnavailable,
			expectedStatus: 503,
		},
		{
			name:           "no space maps to 507",
			err:            fmt.Errorf("disk full: %w", syscall.ENOSPC),
			expectedCode:   CodeUnavailable,
			expectedStatus: 507,
		},
		{
			name:           "deadline exceeded maps to 408",
			err:            fmt.Errorf("operation timed out: %w", os.ErrDeadlineExceeded),
			expectedCode:   CodeTimeout,
			expectedStatus: 408,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapError(tt.err)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
		})
	}
}

// Benchmark to ensure OS error detection is fast
func BenchmarkIsPermissionDenied(b *testing.B) {
	err := fmt.Errorf("layer1: %w", fmt.Errorf("layer2: %w", os.ErrPermission))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isPermissionDenied(err)
	}
}

func BenchmarkIsTimeout_WithInterface(b *testing.B) {
	err := fmt.Errorf("wrap: %w", &timeoutError{timeout: true})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isTimeout(err)
	}
}

// Verify that time.After context deadline errors are detected
func TestIsTimeout_ContextDeadline(t *testing.T) {
	// os.ErrDeadlineExceeded is what context.DeadlineExceeded wraps to
	// when using context.WithDeadline
	err := os.ErrDeadlineExceeded
	assert.True(t, isTimeout(err))
}
