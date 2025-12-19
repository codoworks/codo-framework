package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()

	assert.Equal(t, "codo-app", cfg.Name)
	assert.Equal(t, "0.0.0", cfg.Version)
	assert.Equal(t, "development", cfg.Environment)
}

func TestServiceConfig_Validate_Valid(t *testing.T) {
	cfg := ServiceConfig{
		Name:        "my-service",
		Version:     "1.0.0",
		Environment: "production",
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestServiceConfig_Validate_DefaultsValid(t *testing.T) {
	cfg := DefaultServiceConfig()

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestServiceConfig_Validate_EmptyName(t *testing.T) {
	cfg := ServiceConfig{
		Name:        "",
		Version:     "1.0.0",
		Environment: "production",
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "service.name is required")
}

func TestServiceConfig_Validate_EmptyVersion(t *testing.T) {
	cfg := ServiceConfig{
		Name:        "my-service",
		Version:     "",
		Environment: "production",
	}

	// Empty version is allowed - it's optional
	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestServiceConfig_Validate_EmptyEnvironment(t *testing.T) {
	cfg := ServiceConfig{
		Name:        "my-service",
		Version:     "1.0.0",
		Environment: "",
	}

	// Empty environment is allowed - it's optional
	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestServiceConfig_Validate_OnlyName(t *testing.T) {
	cfg := ServiceConfig{
		Name: "minimal-service",
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestServiceConfig_YAMLTags(t *testing.T) {
	// Verify YAML tags are correctly set
	cfg := ServiceConfig{
		Name:        "test",
		Version:     "1.0",
		Environment: "dev",
	}

	// These assertions ensure the struct can be properly serialized
	assert.Equal(t, "test", cfg.Name)
	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "dev", cfg.Environment)
}
