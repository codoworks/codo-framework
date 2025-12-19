package config

import (
	"fmt"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	KratosPublicURL string `yaml:"kratos_public_url"`
	KratosAdminURL  string `yaml:"kratos_admin_url"`
	KetoReadURL     string `yaml:"keto_read_url"`
	KetoWriteURL    string `yaml:"keto_write_url"`
	SessionCookie   string `yaml:"session_cookie"`
}

// DefaultAuthConfig returns default auth configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		KratosPublicURL: "http://localhost:4433",
		KratosAdminURL:  "http://localhost:4434",
		KetoReadURL:     "http://localhost:4466",
		KetoWriteURL:    "http://localhost:4467",
		SessionCookie:   "ory_kratos_session",
	}
}

// Validate validates auth configuration
func (c *AuthConfig) Validate() error {
	if c.SessionCookie == "" {
		return fmt.Errorf("auth.session_cookie is required")
	}
	return nil
}
