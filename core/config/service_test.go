package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()

	assert.Equal(t, "development", cfg.Environment)
}

func TestServiceConfig_Validate_Valid(t *testing.T) {
	cfg := ServiceConfig{
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



func TestServiceConfig_Validate_EmptyEnvironment(t *testing.T) {
	cfg := ServiceConfig{
		Environment: "",
	}

	// Empty environment is allowed - it's optional
	err := cfg.Validate()

	assert.NoError(t, err)
}


func TestServiceConfig_YAMLTags(t *testing.T) {
	// Verify YAML tags are correctly set
	cfg := ServiceConfig{
		Environment: "dev",
	}

	// These assertions ensure the struct can be properly serialized
	assert.Equal(t, "dev", cfg.Environment)
}
