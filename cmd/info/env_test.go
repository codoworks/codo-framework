package info

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
)

func TestEnvCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"info", "env", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "environment")
	assert.Contains(t, output.String(), "--show-secrets")
}

func TestEnvCmd_Properties(t *testing.T) {
	assert.Equal(t, "env", envCmd.Use)
	assert.Contains(t, envCmd.Short, "environment")
}

func TestEnvCmd_ShowsFrameworkEnvVars(t *testing.T) {
	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Run the command
	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// Check for framework env vars section
	assert.Contains(t, out, "Framework Environment Variables")
	assert.Contains(t, out, "SERVICE_ENVIRONMENT")
	assert.Contains(t, out, "PUBLIC_PORT")
	assert.Contains(t, out, "DB_HOST")
}

func TestEnvCmd_ShowsAllConfigSections(t *testing.T) {
	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// Check for all config sections
	assert.Contains(t, out, "=== Service Configuration ===")
	assert.Contains(t, out, "=== Server Configuration ===")
	assert.Contains(t, out, "=== Database Configuration ===")
	assert.Contains(t, out, "=== Auth Configuration ===")
	assert.Contains(t, out, "=== RabbitMQ Configuration ===")
	assert.Contains(t, out, "=== Middleware Configuration ===")
	assert.Contains(t, out, "=== Errors Configuration ===")
}

func TestEnvCmd_ShowsServerConfig(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Server.PublicPort = 9081
	cfg.Server.ProtectedPort = 9080
	cfg.Server.HiddenPort = 9079
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	assert.Contains(t, out, "9081")
	assert.Contains(t, out, "9080")
	assert.Contains(t, out, "9079")
}

func TestEnvCmd_ShowsDatabaseConfig(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Database.Host = "testdb.example.com"
	cfg.Database.Name = "testdb"
	cfg.Database.User = "testuser"
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	assert.Contains(t, out, "testdb.example.com")
	assert.Contains(t, out, "testdb")
	assert.Contains(t, out, "testuser")
}

func TestEnvCmd_MasksDatabasePassword(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Database.Password = "supersecretpassword"
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	// Reset the show-secrets flag
	showSecrets = false

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// Password should be masked
	assert.NotContains(t, out, "supersecretpassword")
	assert.Contains(t, out, "***MASKED***")
}

func TestEnvCmd_ShowSecretsFlag(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Database.Password = "mysecretpw"
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	// Enable show-secrets
	showSecrets = true
	defer func() { showSecrets = false }()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// Password should be visible
	assert.Contains(t, out, "mysecretpw")
}

func TestEnvCmd_ShowsMiddlewareConfig(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Middleware.Logger.Enabled = true
	cfg.Middleware.CORS.Enabled = true
	cfg.Middleware.Auth.Enabled = false
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	assert.Contains(t, out, "Logger:")
	assert.Contains(t, out, "CORS:")
	assert.Contains(t, out, "Auth:")
}

func TestEnvCmd_ShowsServiceConfigDetails(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Features.DisabledFeatures = []string{"rabbitmq", "kratos"}
	cfg.DevMode = true
	cfg.Response.Strict = true
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// All these fields should be in Service Configuration section
	assert.Contains(t, out, "=== Service Configuration ===")
	assert.Contains(t, out, "Dev Mode:")
	assert.Contains(t, out, "Strict Response:")
	assert.Contains(t, out, "Disabled Features:")
	assert.Contains(t, out, "rabbitmq")
}

func TestEnvCmd_FrameworkEnvVarFromEnvironment(t *testing.T) {
	// Set a CODO_ env var
	os.Setenv("CODO_DB_HOST", "env-db-host.example.com")
	defer os.Unsetenv("CODO_DB_HOST")

	cfg := config.NewWithDefaults()
	// Simulate what config loader does: apply env override to config
	cfg.Database.Host = "env-db-host.example.com"
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// Should show the env value in the framework env vars table
	assert.Contains(t, out, "env-db-host.example.com")
	assert.Contains(t, out, "env") // Source should be "env"
}

func TestEnvCmd_FrameworkEnvVarFromYAML(t *testing.T) {
	// Ensure the env var is NOT set
	os.Unsetenv("CODO_DEV_MODE")

	cfg := config.NewWithDefaults()
	// Simulate value loaded from YAML (different from default)
	cfg.DevMode = true
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// DEV_MODE should show "true" with source "yaml"
	assert.Contains(t, out, "DEV_MODE")
	assert.Contains(t, out, "true")
	assert.Contains(t, out, "yaml") // Source should be "yaml"
}

func TestEnvCmd_MasksRabbitMQPassword(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.RabbitMQ.Password = "rabbitsecret"
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	// Reset the show-secrets flag
	showSecrets = false

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	err := envCmd.RunE(envCmd, []string{})
	require.NoError(t, err)

	out := output.String()

	// Password should be masked
	assert.NotContains(t, out, "rabbitsecret")
	assert.Contains(t, out, "***MASKED***")
}

// Test helper functions

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskDSN(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "postgres with password",
			input:    "postgres://user:password@localhost:5432/db",
			expected: "postgres://user:***@localhost:5432/db",
		},
		{
			name:     "amqp with password",
			input:    "amqp://guest:guest@localhost:5672/",
			expected: "amqp://guest:***@localhost:5672/",
		},
		{
			name:     "no password",
			input:    "postgres://user@localhost:5432/db",
			expected: "postgres://user@localhost:5432/db",
		},
		{
			name:     "no auth",
			input:    "postgres://localhost:5432/db",
			expected: "postgres://localhost:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskDSN(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnabledStr(t *testing.T) {
	assert.Equal(t, "enabled", enabledStr(true))
	assert.Equal(t, "disabled", enabledStr(false))
}
