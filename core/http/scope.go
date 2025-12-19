package http

import "fmt"

// RouterScope represents which API router a handler belongs to
type RouterScope int

const (
	// ScopePublic is for publicly accessible endpoints (no auth)
	ScopePublic RouterScope = iota
	// ScopeProtected is for authenticated endpoints (Kratos session)
	ScopeProtected
	// ScopeHidden is for admin/internal endpoints
	ScopeHidden
)

// String returns the string representation of the scope
func (s RouterScope) String() string {
	switch s {
	case ScopePublic:
		return "public"
	case ScopeProtected:
		return "protected"
	case ScopeHidden:
		return "hidden"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// ParseScope parses a string into a RouterScope
func ParseScope(s string) (RouterScope, error) {
	switch s {
	case "public":
		return ScopePublic, nil
	case "protected":
		return ScopeProtected, nil
	case "hidden":
		return ScopeHidden, nil
	default:
		return 0, fmt.Errorf("unknown scope: %s", s)
	}
}

// DefaultPort returns the default port for the scope
func (s RouterScope) DefaultPort() int {
	switch s {
	case ScopePublic:
		return 8081
	case ScopeProtected:
		return 8080
	case ScopeHidden:
		return 8079
	default:
		return 8080
	}
}
