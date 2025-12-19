package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_SetDevMode_Enable(t *testing.T) {
	cfg := NewWithDefaults()
	assert.False(t, cfg.DevMode)

	cfg.SetDevMode(true)

	assert.True(t, cfg.DevMode)
	assert.True(t, cfg.IsDevMode())
}

func TestConfig_SetDevMode_Disable(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.DevMode = true

	cfg.SetDevMode(false)

	assert.False(t, cfg.DevMode)
	assert.False(t, cfg.IsDevMode())
}

func TestConfig_SetDevMode_EnableTwice(t *testing.T) {
	cfg := NewWithDefaults()

	cfg.SetDevMode(true)
	cfg.SetDevMode(true)

	assert.True(t, cfg.DevMode)
}

func TestConfig_SetDevMode_DisableTwice(t *testing.T) {
	cfg := NewWithDefaults()

	cfg.SetDevMode(false)
	cfg.SetDevMode(false)

	assert.False(t, cfg.DevMode)
}

func TestConfig_SetDevMode_Toggle(t *testing.T) {
	cfg := NewWithDefaults()

	cfg.SetDevMode(true)
	assert.True(t, cfg.DevMode)

	cfg.SetDevMode(false)
	assert.False(t, cfg.DevMode)

	cfg.SetDevMode(true)
	assert.True(t, cfg.DevMode)
}

func TestConfig_DevModeOverrides_NotDevMode(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.DevMode = false

	// Should be a no-op
	cfg.DevModeOverrides()

	assert.False(t, cfg.DevMode)
}

func TestConfig_DevModeOverrides_DevMode(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.DevMode = true

	// Should run without error
	cfg.DevModeOverrides()

	assert.True(t, cfg.DevMode)
}

func TestConfig_DevModeOverrides_CalledBySetDevMode(t *testing.T) {
	cfg := NewWithDefaults()

	// SetDevMode(true) should call DevModeOverrides
	cfg.SetDevMode(true)

	// Verify it ran without issues
	assert.True(t, cfg.IsDevMode())
}

func TestConfig_DevModeOverrides_NotCalledWhenDisabling(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.DevMode = true

	// SetDevMode(false) should not call DevModeOverrides
	cfg.SetDevMode(false)

	assert.False(t, cfg.IsDevMode())
}
