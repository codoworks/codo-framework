package http

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// mockHandler implements Handler interface for testing
type mockHandler struct {
	prefix      string
	scope       RouterScope
	middlewares []echo.MiddlewareFunc
	initErr     error
	initialized bool
	routes      func(g *echo.Group)
}

func (h *mockHandler) Prefix() string                    { return h.prefix }
func (h *mockHandler) Scope() RouterScope                { return h.scope }
func (h *mockHandler) Middlewares() []echo.MiddlewareFunc { return h.middlewares }
func (h *mockHandler) Initialize() error {
	h.initialized = true
	return h.initErr
}
func (h *mockHandler) Routes(g *echo.Group) {
	if h.routes != nil {
		h.routes(g)
	}
}

func TestRegisterHandler(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	h := &mockHandler{prefix: "/test", scope: ScopePublic}
	RegisterHandler(h)

	assert.Equal(t, 1, HandlerCount())
}

func TestGetHandlers(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	// Register handlers with different scopes
	h1 := &mockHandler{prefix: "/public1", scope: ScopePublic}
	h2 := &mockHandler{prefix: "/public2", scope: ScopePublic}
	h3 := &mockHandler{prefix: "/protected", scope: ScopeProtected}
	h4 := &mockHandler{prefix: "/hidden", scope: ScopeHidden}

	RegisterHandler(h1)
	RegisterHandler(h2)
	RegisterHandler(h3)
	RegisterHandler(h4)

	// Test getting handlers by scope
	publicHandlers := GetHandlers(ScopePublic)
	assert.Len(t, publicHandlers, 2)

	protectedHandlers := GetHandlers(ScopeProtected)
	assert.Len(t, protectedHandlers, 1)

	hiddenHandlers := GetHandlers(ScopeHidden)
	assert.Len(t, hiddenHandlers, 1)
}

func TestAllHandlers(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	h1 := &mockHandler{prefix: "/test1", scope: ScopePublic}
	h2 := &mockHandler{prefix: "/test2", scope: ScopeProtected}

	RegisterHandler(h1)
	RegisterHandler(h2)

	all := AllHandlers()
	assert.Len(t, all, 2)
}

func TestClearHandlers(t *testing.T) {
	h := &mockHandler{prefix: "/test", scope: ScopePublic}
	RegisterHandler(h)
	assert.True(t, HandlerCount() > 0)

	ClearHandlers()
	assert.Equal(t, 0, HandlerCount())
}

func TestHandlerCount(t *testing.T) {
	ClearHandlers()
	defer ClearHandlers()

	assert.Equal(t, 0, HandlerCount())

	RegisterHandler(&mockHandler{prefix: "/test1", scope: ScopePublic})
	assert.Equal(t, 1, HandlerCount())

	RegisterHandler(&mockHandler{prefix: "/test2", scope: ScopePublic})
	assert.Equal(t, 2, HandlerCount())
}
