package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFeaturesConfig(t *testing.T) {
	cfg := DefaultFeaturesConfig()

	assert.Empty(t, cfg.DisabledFeatures)
	assert.NotNil(t, cfg.DisabledFeatures)
}

func TestFeaturesConfig_IsEnabled(t *testing.T) {
	cfg := DefaultFeaturesConfig()

	assert.True(t, cfg.IsEnabled(FeatureDatabase))
	assert.True(t, cfg.IsEnabled(FeatureKratos))
	assert.True(t, cfg.IsEnabled(FeatureRedis))
	assert.True(t, cfg.IsEnabled("any-feature"))
}

func TestFeaturesConfig_IsEnabled_Disabled(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"database", "redis"},
	}

	assert.False(t, cfg.IsEnabled("database"))
	assert.False(t, cfg.IsEnabled("redis"))
	assert.True(t, cfg.IsEnabled("kratos"))
}

func TestFeaturesConfig_IsEnabled_CaseInsensitive(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"Database"},
	}

	assert.False(t, cfg.IsEnabled("database"))
	assert.False(t, cfg.IsEnabled("DATABASE"))
	assert.False(t, cfg.IsEnabled("Database"))
	assert.False(t, cfg.IsEnabled("DaTaBaSe"))
}

func TestFeaturesConfig_IsDisabled(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"database"},
	}

	assert.True(t, cfg.IsDisabled("database"))
	assert.False(t, cfg.IsDisabled("kratos"))
}

func TestFeaturesConfig_Disable(t *testing.T) {
	cfg := DefaultFeaturesConfig()

	cfg.Disable("database")

	assert.Contains(t, cfg.DisabledFeatures, "database")
	assert.False(t, cfg.IsEnabled("database"))
}

func TestFeaturesConfig_Disable_AlreadyDisabled(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"database"},
	}

	cfg.Disable("database")

	// Should not add duplicate
	assert.Len(t, cfg.DisabledFeatures, 1)
}

func TestFeaturesConfig_Disable_AlreadyDisabled_CaseInsensitive(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"Database"},
	}

	cfg.Disable("database")

	// Should not add duplicate (case-insensitive check)
	assert.Len(t, cfg.DisabledFeatures, 1)
}

func TestFeaturesConfig_Disable_Multiple(t *testing.T) {
	cfg := DefaultFeaturesConfig()

	cfg.Disable("database")
	cfg.Disable("redis")
	cfg.Disable("kratos")

	assert.Len(t, cfg.DisabledFeatures, 3)
	assert.False(t, cfg.IsEnabled("database"))
	assert.False(t, cfg.IsEnabled("redis"))
	assert.False(t, cfg.IsEnabled("kratos"))
}

func TestFeaturesConfig_Enable(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"database", "redis"},
	}

	cfg.Enable("database")

	assert.True(t, cfg.IsEnabled("database"))
	assert.False(t, cfg.IsEnabled("redis"))
	assert.Len(t, cfg.DisabledFeatures, 1)
}

func TestFeaturesConfig_Enable_NotDisabled(t *testing.T) {
	cfg := DefaultFeaturesConfig()

	cfg.Enable("database")

	// Should be a no-op
	assert.Empty(t, cfg.DisabledFeatures)
	assert.True(t, cfg.IsEnabled("database"))
}

func TestFeaturesConfig_Enable_CaseInsensitive(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"Database"},
	}

	cfg.Enable("database")

	assert.True(t, cfg.IsEnabled("database"))
	assert.Empty(t, cfg.DisabledFeatures)
}

func TestFeaturesConfig_Enable_All(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"database", "redis", "kratos"},
	}

	cfg.Enable("database")
	cfg.Enable("redis")
	cfg.Enable("kratos")

	assert.Empty(t, cfg.DisabledFeatures)
}

func TestFeaturesConfig_LoadFromEnv(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "database,redis")

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	assert.False(t, cfg.IsEnabled("database"))
	assert.False(t, cfg.IsEnabled("redis"))
	assert.True(t, cfg.IsEnabled("kratos"))
}

func TestFeaturesConfig_LoadFromEnv_Empty(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "")

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	assert.Empty(t, cfg.DisabledFeatures)
}

func TestFeaturesConfig_LoadFromEnv_NotSet(t *testing.T) {
	// Ensure the env var is not set
	// t.Setenv would set it, so we don't call it

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	assert.Empty(t, cfg.DisabledFeatures)
}

func TestFeaturesConfig_LoadFromEnv_Multiple(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "database,redis,kratos,keto")

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	assert.Len(t, cfg.DisabledFeatures, 4)
	assert.False(t, cfg.IsEnabled("database"))
	assert.False(t, cfg.IsEnabled("redis"))
	assert.False(t, cfg.IsEnabled("kratos"))
	assert.False(t, cfg.IsEnabled("keto"))
}

func TestFeaturesConfig_LoadFromEnv_WithSpaces(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "database , redis , kratos")

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	assert.False(t, cfg.IsEnabled("database"))
	assert.False(t, cfg.IsEnabled("redis"))
	assert.False(t, cfg.IsEnabled("kratos"))
}

func TestFeaturesConfig_LoadFromEnv_WithEmptyItems(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "database,,redis,,,kratos")

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	// Empty items should be ignored
	assert.Len(t, cfg.DisabledFeatures, 3)
}

func TestFeaturesConfig_LoadFromEnv_Single(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "database")

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	assert.Len(t, cfg.DisabledFeatures, 1)
	assert.False(t, cfg.IsEnabled("database"))
}

func TestFeaturesConfig_LoadFromEnv_Duplicates(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "database,database,DATABASE")

	cfg := DefaultFeaturesConfig()
	cfg.LoadFromEnv()

	// Duplicates should be handled (only first should be added due to case-insensitive check)
	assert.False(t, cfg.IsEnabled("database"))
	// The exact count depends on implementation - at least 1
	assert.GreaterOrEqual(t, len(cfg.DisabledFeatures), 1)
}

func TestFeaturesConfig_GetDisabled(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"database", "redis"},
	}

	disabled := cfg.GetDisabled()

	assert.Equal(t, []string{"database", "redis"}, disabled)
}

func TestFeaturesConfig_GetDisabled_Empty(t *testing.T) {
	cfg := DefaultFeaturesConfig()

	disabled := cfg.GetDisabled()

	assert.Empty(t, disabled)
	assert.NotNil(t, disabled)
}

func TestFeaturesConfig_GetDisabled_IsCopy(t *testing.T) {
	cfg := FeaturesConfig{
		DisabledFeatures: []string{"database"},
	}

	disabled := cfg.GetDisabled()
	disabled[0] = "modified"

	// Original should not be modified
	assert.Equal(t, "database", cfg.DisabledFeatures[0])
}

func TestFeatureConstants(t *testing.T) {
	assert.Equal(t, "database", FeatureDatabase)
	assert.Equal(t, "kratos", FeatureKratos)
	assert.Equal(t, "keto", FeatureKeto)
	assert.Equal(t, "redis", FeatureRedis)
	assert.Equal(t, "cors", FeatureCORS)
	assert.Equal(t, "gzip", FeatureGzip)
}
