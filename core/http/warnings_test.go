package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWarning(t *testing.T) {
	w := NewWarning("DEPRECATED", "This field is deprecated")

	assert.Equal(t, "DEPRECATED", w.Code)
	assert.Equal(t, "This field is deprecated", w.Message)
	assert.NotNil(t, w.Details)
	assert.Empty(t, w.Details)
}

func TestWarning_WithDetail(t *testing.T) {
	w := NewWarning("LIMIT_EXCEEDED", "Rate limit exceeded").
		WithDetail("limit", 100).
		WithDetail("current", 150)

	assert.Equal(t, "LIMIT_EXCEEDED", w.Code)
	assert.Equal(t, 100, w.Details["limit"])
	assert.Equal(t, 150, w.Details["current"])
}

func TestWarning_WithDetail_InitializesNilMap(t *testing.T) {
	// Create warning without using NewWarning (nil Details)
	w := Warning{
		Code:    "TEST",
		Message: "Test warning",
	}

	// WithDetail should initialize the map
	w = w.WithDetail("key", "value")

	assert.NotNil(t, w.Details)
	assert.Equal(t, "value", w.Details["key"])
}

func TestWarning_Chaining(t *testing.T) {
	w := NewWarning("CODE", "Message").
		WithDetail("key1", "value1").
		WithDetail("key2", "value2")

	// Both details should be present
	assert.Equal(t, "value1", w.Details["key1"])
	assert.Equal(t, "value2", w.Details["key2"])
}

func TestWarning_MultipleDetails(t *testing.T) {
	w := NewWarning("PARTIAL_SUCCESS", "Some items failed").
		WithDetail("successful", 8).
		WithDetail("failed", 2).
		WithDetail("total", 10)

	assert.Equal(t, 3, len(w.Details))
	assert.Equal(t, 8, w.Details["successful"])
	assert.Equal(t, 2, w.Details["failed"])
	assert.Equal(t, 10, w.Details["total"])
}
