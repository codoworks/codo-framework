package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigValueByPath(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.DevMode = true
	cfg.Database.Host = "custom-db.example.com"
	cfg.Database.Port = 5433
	cfg.Auth.KratosPublicURL = "https://auth.example.com"
	cfg.Server.PublicPort = 9000

	tests := []struct {
		path     string
		expected string
	}{
		{"dev_mode", "true"},
		{"database.host", "custom-db.example.com"},
		{"database.port", "5433"},
		{"auth.kratos_public_url", "https://auth.example.com"},
		{"server.public_port", "9000"},
		{"service.environment", "development"},
		{"database.driver", "postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := GetConfigValueByPath(cfg, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigValueByPath_InvalidPath(t *testing.T) {
	cfg := NewWithDefaults()

	result := GetConfigValueByPath(cfg, "nonexistent.field")
	assert.Equal(t, "", result)

	result = GetConfigValueByPath(cfg, "database.nonexistent")
	assert.Equal(t, "", result)
}

func TestFrameworkEnvVar_GetActualValue(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.DevMode = true
	cfg.Database.Host = "test-host.example.com"

	// Test dev_mode
	v := GetFrameworkEnvVarByName("DEV_MODE")
	assert.NotNil(t, v)
	assert.Equal(t, "true", v.GetActualValue(cfg))

	// Test database.host
	v = GetFrameworkEnvVarByName("DB_HOST")
	assert.NotNil(t, v)
	assert.Equal(t, "test-host.example.com", v.GetActualValue(cfg))
}

func TestFrameworkEnvVar_DetectSource_Default(t *testing.T) {
	cfg := NewWithDefaults()

	// SERVICE_ENVIRONMENT with default value
	v := GetFrameworkEnvVarByName("SERVICE_ENVIRONMENT")
	assert.NotNil(t, v)
	assert.Equal(t, "default", v.DetectSource(cfg))
}

func TestFrameworkEnvVar_DetectSource_YAML(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.DevMode = true // Set via yaml, different from default (false)

	v := GetFrameworkEnvVarByName("DEV_MODE")
	assert.NotNil(t, v)
	assert.Equal(t, "yaml", v.DetectSource(cfg))
}

func TestFrameworkEnvVar_DetectSource_Env(t *testing.T) {
	// Set env var
	os.Setenv("CODO_DEV_MODE", "true")
	defer os.Unsetenv("CODO_DEV_MODE")

	cfg := NewWithDefaults()
	cfg.DevMode = true // Config matches env var

	v := GetFrameworkEnvVarByName("DEV_MODE")
	assert.NotNil(t, v)
	assert.Equal(t, "env", v.DetectSource(cfg))
}

func TestFrameworkEnvVar_DetectSource_EnvWithDefault(t *testing.T) {
	// Set env var to a non-default value
	os.Setenv("CODO_DB_HOST", "env-host.example.com")
	defer os.Unsetenv("CODO_DB_HOST")

	cfg := NewWithDefaults()
	cfg.Database.Host = "env-host.example.com" // Config matches env var

	v := GetFrameworkEnvVarByName("DB_HOST")
	assert.NotNil(t, v)
	assert.Equal(t, "env", v.DetectSource(cfg))
}

func TestGetFrameworkEnvVarByName(t *testing.T) {
	v := GetFrameworkEnvVarByName("DEV_MODE")
	assert.NotNil(t, v)
	assert.Equal(t, "DEV_MODE", v.Name)
	assert.Equal(t, "dev_mode", v.ConfigPath)
	assert.Equal(t, "bool", v.Type)

	v = GetFrameworkEnvVarByName("DB_PASSWORD")
	assert.NotNil(t, v)
	assert.True(t, v.Sensitive)

	v = GetFrameworkEnvVarByName("NONEXISTENT")
	assert.Nil(t, v)
}

func TestFrameworkEnvVars_AllHaveConfigPath(t *testing.T) {
	vars := FrameworkEnvVars()

	for _, v := range vars {
		t.Run(v.Name, func(t *testing.T) {
			assert.NotEmpty(t, v.ConfigPath, "ConfigPath should not be empty for %s", v.Name)

			// Verify the path resolves (may be empty for optional fields like DB_PASSWORD)
			cfg := NewWithDefaults()
			value := GetConfigValueByPath(cfg, v.ConfigPath)

			// For fields with non-empty defaults, verify the resolved value matches
			if v.Default != "" {
				assert.Equal(t, v.Default, value, "ConfigPath %s should resolve to default value for %s", v.ConfigPath, v.Name)
			}
		})
	}
}

func TestFrameworkEnvVar_FullName(t *testing.T) {
	v := GetFrameworkEnvVarByName("DEV_MODE")
	assert.NotNil(t, v)
	assert.Equal(t, "CODO_DEV_MODE", v.FullName())
}

func TestFrameworkEnvVar_IsSet(t *testing.T) {
	v := GetFrameworkEnvVarByName("DEV_MODE")
	assert.NotNil(t, v)

	// Not set initially
	os.Unsetenv("CODO_DEV_MODE")
	assert.False(t, v.IsSet())

	// Set the env var
	os.Setenv("CODO_DEV_MODE", "true")
	defer os.Unsetenv("CODO_DEV_MODE")
	assert.True(t, v.IsSet())
}
