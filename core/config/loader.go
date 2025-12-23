package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Default config file locations
var defaultConfigPaths = []string{
	"config/app.yaml",
	"app.yaml",
	"config.yaml",
}

// LoadFromReader loads config from an io.Reader
func LoadFromReader(r io.Reader) (*Config, error) {
	cfg := NewWithDefaults()

	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(cfg); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults to any missing values
	cfg.ApplyDefaults()

	if err := cfg.applyEnvOverrides(); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadFromFile loads config from a specific file path
func LoadFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	return LoadFromReader(f)
}

// Load loads configuration from default locations
func Load() (*Config, error) {
	cfg := NewWithDefaults()

	// Try default config locations
	var loaded bool
	for _, path := range defaultConfigPaths {
		if _, err := os.Stat(path); err == nil {
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			decoder := yaml.NewDecoder(f)
			if err := decoder.Decode(cfg); err != nil && err != io.EOF {
				f.Close()
				return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
			}
			f.Close()
			loaded = true
			break
		}
	}

	// It's OK if no config file exists - we use defaults
	_ = loaded

	// Apply defaults to any missing values
	cfg.ApplyDefaults()

	// Apply environment overrides
	if err := cfg.applyEnvOverrides(); err != nil {
		return nil, err
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides using Viper
func (c *Config) applyEnvOverrides() error {
	v := viper.New()
	v.SetEnvPrefix("CODO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Service
	if val := v.GetString("SERVICE_ENVIRONMENT"); val != "" {
		c.Service.Environment = val
	}

	// Server
	if val := v.GetInt("PUBLIC_PORT"); val != 0 {
		c.Server.PublicPort = val
	}
	if val := v.GetInt("PROTECTED_PORT"); val != 0 {
		c.Server.ProtectedPort = val
	}
	if val := v.GetInt("HIDDEN_PORT"); val != 0 {
		c.Server.HiddenPort = val
	}

	// Database
	if val := v.GetString("DB_DRIVER"); val != "" {
		c.Database.Driver = val
	}
	if val := v.GetString("DB_HOST"); val != "" {
		c.Database.Host = val
	}
	if val := v.GetInt("DB_PORT"); val != 0 {
		c.Database.Port = val
	}
	if val := v.GetString("DB_NAME"); val != "" {
		c.Database.Name = val
	}
	if val := v.GetString("DB_USER"); val != "" {
		c.Database.User = val
	}
	if val := v.GetString("DB_PASSWORD"); val != "" {
		c.Database.Password = val
	}
	if val := v.GetString("DB_SSL_MODE"); val != "" {
		c.Database.SSLMode = val
	}

	// Auth
	if val := v.GetString("KRATOS_PUBLIC_URL"); val != "" {
		c.Auth.KratosPublicURL = val
	}
	if val := v.GetString("KRATOS_ADMIN_URL"); val != "" {
		c.Auth.KratosAdminURL = val
	}
	if val := v.GetString("KETO_READ_URL"); val != "" {
		c.Auth.KetoReadURL = val
	}
	if val := v.GetString("KETO_WRITE_URL"); val != "" {
		c.Auth.KetoWriteURL = val
	}
	if val := v.GetString("SESSION_COOKIE"); val != "" {
		c.Auth.SessionCookie = val
	}

	// Dev mode - only override if explicitly set in environment
	if v.IsSet("DEV_MODE") {
		c.DevMode = v.GetBool("DEV_MODE")
	}

	// Load features from env
	c.Features.LoadFromEnv()

	return nil
}

// GetDefaultConfigPaths returns the list of default config file paths
func GetDefaultConfigPaths() []string {
	paths := make([]string, len(defaultConfigPaths))
	copy(paths, defaultConfigPaths)
	return paths
}
