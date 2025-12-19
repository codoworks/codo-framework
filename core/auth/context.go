package auth

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
)

// Context keys
type contextKey string

const (
	identityKey contextKey = "identity"
)

// ErrNoIdentity is returned when no identity is in context
var ErrNoIdentity = fmt.Errorf("no identity in context")

// ContextWithIdentity adds an identity to the context
func ContextWithIdentity(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, identityKey, identity)
}

// IdentityFromContext retrieves the identity from context
func IdentityFromContext(ctx context.Context) (*Identity, bool) {
	identity, ok := ctx.Value(identityKey).(*Identity)
	return identity, ok
}

// SetIdentity sets the identity in the Echo context
func SetIdentity(c echo.Context, identity *Identity) {
	ctx := ContextWithIdentity(c.Request().Context(), identity)
	c.SetRequest(c.Request().WithContext(ctx))
}

// GetIdentity retrieves the identity from the Echo context
func GetIdentity(c echo.Context) (*Identity, error) {
	identity, ok := IdentityFromContext(c.Request().Context())
	if !ok {
		return nil, ErrNoIdentity
	}
	return identity, nil
}

// MustGetIdentity retrieves the identity or panics
func MustGetIdentity(c echo.Context) *Identity {
	identity, err := GetIdentity(c)
	if err != nil {
		panic(err)
	}
	return identity
}
