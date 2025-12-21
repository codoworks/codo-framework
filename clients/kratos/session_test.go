package kratos

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession_IsExpired_True(t *testing.T) {
	session := &Session{
		ID:        "session-123",
		Active:    true,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired 1 hour ago
	}

	assert.True(t, session.IsExpired())
}

func TestSession_IsExpired_False(t *testing.T) {
	session := &Session{
		ID:        "session-123",
		Active:    true,
		ExpiresAt: time.Now().Add(1 * time.Hour), // expires in 1 hour
	}

	assert.False(t, session.IsExpired())
}

func TestSession_IsExpired_JustExpired(t *testing.T) {
	session := &Session{
		ID:        "session-123",
		Active:    true,
		ExpiresAt: time.Now().Add(-1 * time.Millisecond), // just expired
	}

	assert.True(t, session.IsExpired())
}

func TestSession_Fields(t *testing.T) {
	expiresAt := time.Now().Add(1 * time.Hour)
	session := &Session{
		ID:        "session-123",
		Active:    true,
		ExpiresAt: expiresAt,
	}
	session.Identity.ID = "user-456"
	session.Identity.Traits = map[string]any{
		"email": "test@example.com",
	}

	assert.Equal(t, "session-123", session.ID)
	assert.True(t, session.Active)
	assert.Equal(t, expiresAt, session.ExpiresAt)
	assert.Equal(t, "user-456", session.Identity.ID)
	assert.Equal(t, "test@example.com", session.Identity.Traits["email"])
}
