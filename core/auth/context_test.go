package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextWithIdentity(t *testing.T) {
	ctx := context.Background()
	identity := &Identity{ID: "user-123"}

	newCtx := ContextWithIdentity(ctx, identity)

	assert.NotEqual(t, ctx, newCtx)
	retrieved := newCtx.Value(identityKey)
	assert.Equal(t, identity, retrieved)
}

func TestIdentityFromContext(t *testing.T) {
	ctx := context.Background()
	identity := &Identity{ID: "user-123"}
	ctx = ContextWithIdentity(ctx, identity)

	retrieved, ok := IdentityFromContext(ctx)

	assert.True(t, ok)
	assert.Equal(t, identity, retrieved)
	assert.Equal(t, "user-123", retrieved.ID)
}

func TestIdentityFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	retrieved, ok := IdentityFromContext(ctx)

	assert.False(t, ok)
	assert.Nil(t, retrieved)
}

func TestIdentityFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), identityKey, "not an identity")

	retrieved, ok := IdentityFromContext(ctx)

	assert.False(t, ok)
	assert.Nil(t, retrieved)
}

func TestSetIdentity(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	identity := &Identity{ID: "user-123"}
	SetIdentity(c, identity)

	retrieved, ok := IdentityFromContext(c.Request().Context())
	assert.True(t, ok)
	assert.Equal(t, identity, retrieved)
}

func TestGetIdentity(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	identity := &Identity{ID: "user-123"}
	SetIdentity(c, identity)

	retrieved, err := GetIdentity(c)
	require.NoError(t, err)
	assert.Equal(t, identity, retrieved)
	assert.Equal(t, "user-123", retrieved.ID)
}

func TestGetIdentity_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	retrieved, err := GetIdentity(c)
	assert.ErrorIs(t, err, ErrNoIdentity)
	assert.Nil(t, retrieved)
}

func TestMustGetIdentity(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	identity := &Identity{ID: "user-123"}
	SetIdentity(c, identity)

	retrieved := MustGetIdentity(c)
	assert.Equal(t, identity, retrieved)
}

func TestMustGetIdentity_Panics(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.Panics(t, func() {
		MustGetIdentity(c)
	})
}

func TestErrNoIdentity(t *testing.T) {
	assert.NotNil(t, ErrNoIdentity)
	assert.Equal(t, "no identity in context", ErrNoIdentity.Error())
}
