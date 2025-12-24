package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLoggerConfig(t *testing.T) {
	cfg := DefaultLoggerConfig()

	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
}

func TestLoggerConfig_CustomValues(t *testing.T) {
	cfg := LoggerConfig{
		Level:  "debug",
		Format: "text",
	}

	assert.Equal(t, "debug", cfg.Level)
	assert.Equal(t, "text", cfg.Format)
}
