package config

import (
	"os"
	"strings"
)

// Known feature names
const (
	FeatureDatabase = "database"
	FeatureKratos   = "kratos"
	FeatureKeto     = "keto"
	FeatureRedis    = "redis"
	FeatureCORS     = "cors"
	FeatureGzip     = "gzip"
)

// FeaturesConfig holds feature toggle configuration
type FeaturesConfig struct {
	DisabledFeatures []string `yaml:"disabled"`
}

// DefaultFeaturesConfig returns default features configuration
func DefaultFeaturesConfig() FeaturesConfig {
	return FeaturesConfig{
		DisabledFeatures: []string{},
	}
}

// IsEnabled returns true if a feature is enabled
func (c *FeaturesConfig) IsEnabled(feature string) bool {
	for _, disabled := range c.DisabledFeatures {
		if strings.EqualFold(disabled, feature) {
			return false
		}
	}
	return true
}

// IsDisabled returns true if a feature is disabled
func (c *FeaturesConfig) IsDisabled(feature string) bool {
	return !c.IsEnabled(feature)
}

// Disable disables a feature
func (c *FeaturesConfig) Disable(feature string) {
	if !c.IsEnabled(feature) {
		return // Already disabled
	}
	c.DisabledFeatures = append(c.DisabledFeatures, feature)
}

// Enable enables a feature (removes from disabled list)
func (c *FeaturesConfig) Enable(feature string) {
	newList := make([]string, 0, len(c.DisabledFeatures))
	for _, f := range c.DisabledFeatures {
		if !strings.EqualFold(f, feature) {
			newList = append(newList, f)
		}
	}
	c.DisabledFeatures = newList
}

// LoadFromEnv loads disabled features from DISABLE_FEATURES env var
func (c *FeaturesConfig) LoadFromEnv() {
	envVal := os.Getenv("DISABLE_FEATURES")
	if envVal == "" {
		return
	}

	features := strings.Split(envVal, ",")
	for _, f := range features {
		f = strings.TrimSpace(f)
		if f != "" {
			c.Disable(f)
		}
	}
}

// GetDisabled returns a copy of the disabled features list
func (c *FeaturesConfig) GetDisabled() []string {
	result := make([]string, len(c.DisabledFeatures))
	copy(result, c.DisabledFeatures)
	return result
}
