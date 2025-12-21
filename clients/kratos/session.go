package kratos

import (
	"time"
)

// Session represents a Kratos session response
type Session struct {
	ID        string    `json:"id"`
	Active    bool      `json:"active"`
	ExpiresAt time.Time `json:"expires_at"`
	Identity  struct {
		ID     string         `json:"id"`
		Traits map[string]any `json:"traits"`
	} `json:"identity"`
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
