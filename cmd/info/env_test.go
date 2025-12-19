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
}

func TestEnvCmd_Properties(t *testing.T) {
	assert.Equal(t, "env", envCmd.Use)
	assert.Equal(t, "Show environment information", envCmd.Short)
}

func TestEnvCmd_ShowsConfig(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Service.Name = "test-service"
	cfg.Service.Version = "1.2.3"
	cfg.Database.Password = "secret123"
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the Run function
	envCmd.Run(envCmd, []string{})

	out := output.String()
	assert.Contains(t, out, "=== Service ===")
	assert.Contains(t, out, "test-service")
	assert.Contains(t, out, "1.2.3")
	assert.Contains(t, out, "=== Server ===")
	assert.Contains(t, out, "=== Database ===")
	assert.Contains(t, out, "postgres")
	// Password should be masked
	assert.NotContains(t, out, "secret123")
	assert.Contains(t, out, "s*******3")
}

func TestEnvCmd_ShowsCodoEnvVars(t *testing.T) {
	// Set a CODO_ env var
	os.Setenv("CODO_TEST_VAR", "test-value")
	defer os.Unsetenv("CODO_TEST_VAR")

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	envCmd.Run(envCmd, []string{})

	assert.Contains(t, output.String(), "=== Environment Variables ===")
	assert.Contains(t, output.String(), "CODO_TEST_VAR=test-value")
}

func TestMaskPassword(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "***"},
		{"a", "***"},
		{"ab", "***"},
		{"abc", "a*c"},
		{"password", "p******d"},
		{"secret123456", "s**********6"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MaskPassword(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskPassword_Short(t *testing.T) {
	assert.Equal(t, "***", MaskPassword(""))
	assert.Equal(t, "***", MaskPassword("a"))
	assert.Equal(t, "***", MaskPassword("ab"))
}

func TestIsSecret(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"CODO_PASSWORD", true},
		{"CODO_DB_PASSWORD", true},
		{"CODO_SECRET_KEY", true},
		{"CODO_API_KEY", true},
		{"CODO_AUTH_TOKEN", true},
		{"CODO_SERVICE_NAME", false},
		{"CODO_PORT", false},
		{"password", true},
		{"MY_PASSWORD_VAR", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSecret(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnvCmd_MasksSecretEnvVars(t *testing.T) {
	os.Setenv("CODO_DB_PASSWORD", "supersecret")
	defer os.Unsetenv("CODO_DB_PASSWORD")

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	envCmd.Run(envCmd, []string{})

	// Secret env var should be masked (11 chars - 2 = 9 asterisks)
	assert.NotContains(t, output.String(), "supersecret")
	assert.Contains(t, output.String(), "CODO_DB_PASSWORD=s*********t")
}
