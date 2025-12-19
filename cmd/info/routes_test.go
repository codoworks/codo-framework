package info

import (
	"bytes"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/http"
)

// mockHandler is a test handler implementation
type mockHandler struct {
	scope  http.RouterScope
	prefix string
}

func (h *mockHandler) Scope() http.RouterScope {
	return h.scope
}

func (h *mockHandler) Prefix() string {
	return h.prefix
}

func (h *mockHandler) Middlewares() []echo.MiddlewareFunc {
	return nil
}

func (h *mockHandler) Initialize() error {
	return nil
}

func (h *mockHandler) Routes(g *echo.Group) {}

func TestRoutesCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	ResetRoutesFlags()
	defer func() {
		cmd.ResetFlags()
		ResetRoutesFlags()
	}()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"info", "routes", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "routes")
	assert.Contains(t, output.String(), "--scope")
}

func TestRoutesCmd_Properties(t *testing.T) {
	assert.Equal(t, "routes", routesCmd.Use)
	assert.Equal(t, "Show registered routes", routesCmd.Short)
}

func TestRoutesCmd_NoHandlersForScope(t *testing.T) {
	ResetRoutesFlags()
	defer ResetRoutesFlags()

	http.ClearHandlers()
	defer http.ClearHandlers()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	// Filter by public scope - default handlers are all protected
	scopeFilter = "public"

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := routesCmd.RunE(routesCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "No handlers registered")
}

func TestRoutesCmd_WithHandlers(t *testing.T) {
	ResetRoutesFlags()
	defer ResetRoutesFlags()

	http.ClearHandlers()
	defer http.ClearHandlers()

	// Register some test handlers
	http.RegisterHandler(&mockHandler{scope: http.ScopePublic, prefix: "/api/v1/health"})
	http.RegisterHandler(&mockHandler{scope: http.ScopeProtected, prefix: "/api/v1/users"})
	http.RegisterHandler(&mockHandler{scope: http.ScopeHidden, prefix: "/api/v1/admin"})

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := routesCmd.RunE(routesCmd, []string{})
	require.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "SCOPE")
	assert.Contains(t, out, "PREFIX")
	assert.Contains(t, out, "public")
	assert.Contains(t, out, "/api/v1/health")
	assert.Contains(t, out, "protected")
	assert.Contains(t, out, "/api/v1/users")
	assert.Contains(t, out, "hidden")
	assert.Contains(t, out, "/api/v1/admin")
}

func TestRoutesCmd_ScopeFilter(t *testing.T) {
	ResetRoutesFlags()
	defer ResetRoutesFlags()

	http.ClearHandlers()
	defer http.ClearHandlers()

	http.RegisterHandler(&mockHandler{scope: http.ScopePublic, prefix: "/api/v1/health"})
	http.RegisterHandler(&mockHandler{scope: http.ScopeProtected, prefix: "/api/v1/users"})

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	// Set the scope filter
	scopeFilter = "public"

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := routesCmd.RunE(routesCmd, []string{})
	require.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "public")
	assert.Contains(t, out, "/api/v1/health")
	assert.NotContains(t, out, "protected")
	assert.NotContains(t, out, "/api/v1/users")
}

func TestRoutesCmd_InvalidScope(t *testing.T) {
	ResetRoutesFlags()
	defer ResetRoutesFlags()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	// Set invalid scope filter
	scopeFilter = "invalid"

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := routesCmd.RunE(routesCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scope")
}

func TestResetRoutesFlags(t *testing.T) {
	scopeFilter = "public"
	ResetRoutesFlags()
	assert.Equal(t, "", scopeFilter)
}
