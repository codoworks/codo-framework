package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouterScope_String(t *testing.T) {
	tests := []struct {
		scope    RouterScope
		expected string
	}{
		{ScopePublic, "public"},
		{ScopeProtected, "protected"},
		{ScopeHidden, "hidden"},
		{RouterScope(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.scope.String())
		})
	}
}

func TestParseScope(t *testing.T) {
	tests := []struct {
		input       string
		expected    RouterScope
		expectError bool
	}{
		{"public", ScopePublic, false},
		{"protected", ScopeProtected, false},
		{"hidden", ScopeHidden, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			scope, err := ParseScope(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, scope)
			}
		})
	}
}

func TestRouterScope_DefaultPort(t *testing.T) {
	tests := []struct {
		scope    RouterScope
		expected int
	}{
		{ScopePublic, 8081},
		{ScopeProtected, 8080},
		{ScopeHidden, 8079},
		{RouterScope(99), 8080},
	}

	for _, tt := range tests {
		t.Run(tt.scope.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.scope.DefaultPort())
		})
	}
}
