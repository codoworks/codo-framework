package config

import (
	"os"
)

// FrameworkEnvPrefix is the prefix for all framework environment variables
const FrameworkEnvPrefix = "CODO_"

// FrameworkEnvVar describes a framework environment variable
type FrameworkEnvVar struct {
	// Name is the variable name without prefix (e.g., "DB_HOST" not "CODO_DB_HOST")
	Name string
	// Type is the data type ("string", "int", "bool")
	Type string
	// Default is the default value as a string
	Default string
	// Description provides documentation
	Description string
	// Sensitive indicates if the value should be masked unless --show-secrets is used
	Sensitive bool
	// ConfigPath is the dot-notation path in the config struct (e.g., "database.host")
	ConfigPath string
}

// FullName returns the full environment variable name with prefix
func (v FrameworkEnvVar) FullName() string {
	return FrameworkEnvPrefix + v.Name
}

// IsSet returns true if the environment variable is set
func (v FrameworkEnvVar) IsSet() bool {
	_, exists := os.LookupEnv(v.FullName())
	return exists
}

// GetValue returns the current value from the environment, or the default
func (v FrameworkEnvVar) GetValue() string {
	if val, exists := os.LookupEnv(v.FullName()); exists {
		return val
	}
	return v.Default
}

// Source returns where the value is coming from: "env", "yaml", or "default"
// Note: This is a simplified check - it only distinguishes env from default
// For full source detection including yaml, the caller needs access to the config
func (v FrameworkEnvVar) Source() string {
	if v.IsSet() {
		return "env"
	}
	return "default"
}

// FrameworkEnvVars returns all framework environment variables
// These are all variables read in applyEnvOverrides() in loader.go
func FrameworkEnvVars() []FrameworkEnvVar {
	return []FrameworkEnvVar{
		// Service
		{
			Name:        "SERVICE_ENVIRONMENT",
			Type:        "string",
			Default:     "development",
			Description: "Runtime environment name",
			ConfigPath:  "service.environment",
		},

		// Server
		{
			Name:        "PUBLIC_PORT",
			Type:        "int",
			Default:     "8081",
			Description: "Public API server port",
			ConfigPath:  "server.public_port",
		},
		{
			Name:        "PROTECTED_PORT",
			Type:        "int",
			Default:     "8080",
			Description: "Protected API server port (auth required)",
			ConfigPath:  "server.protected_port",
		},
		{
			Name:        "HIDDEN_PORT",
			Type:        "int",
			Default:     "8079",
			Description: "Hidden API server port (admin only)",
			ConfigPath:  "server.hidden_port",
		},

		// Database
		{
			Name:        "DB_DRIVER",
			Type:        "string",
			Default:     "postgres",
			Description: "Database driver (postgres, mysql, sqlite)",
			ConfigPath:  "database.driver",
		},
		{
			Name:        "DB_HOST",
			Type:        "string",
			Default:     "localhost",
			Description: "Database server hostname",
			ConfigPath:  "database.host",
		},
		{
			Name:        "DB_PORT",
			Type:        "int",
			Default:     "5432",
			Description: "Database server port",
			ConfigPath:  "database.port",
		},
		{
			Name:        "DB_NAME",
			Type:        "string",
			Default:     "codo",
			Description: "Database name",
			ConfigPath:  "database.name",
		},
		{
			Name:        "DB_USER",
			Type:        "string",
			Default:     "codo",
			Description: "Database username",
			ConfigPath:  "database.user",
		},
		{
			Name:        "DB_PASSWORD",
			Type:        "string",
			Default:     "",
			Description: "Database password",
			Sensitive:   true,
			ConfigPath:  "database.password",
		},
		{
			Name:        "DB_SSL_MODE",
			Type:        "string",
			Default:     "disable",
			Description: "Database SSL mode",
			ConfigPath:  "database.ssl_mode",
		},

		// Auth
		{
			Name:        "KRATOS_PUBLIC_URL",
			Type:        "string",
			Default:     "http://localhost:4433",
			Description: "Ory Kratos public API URL",
			ConfigPath:  "auth.kratos_public_url",
		},
		{
			Name:        "KRATOS_ADMIN_URL",
			Type:        "string",
			Default:     "http://localhost:4434",
			Description: "Ory Kratos admin API URL",
			ConfigPath:  "auth.kratos_admin_url",
		},
		{
			Name:        "KETO_READ_URL",
			Type:        "string",
			Default:     "http://localhost:4466",
			Description: "Ory Keto read API URL",
			ConfigPath:  "auth.keto_read_url",
		},
		{
			Name:        "KETO_WRITE_URL",
			Type:        "string",
			Default:     "http://localhost:4467",
			Description: "Ory Keto write API URL",
			ConfigPath:  "auth.keto_write_url",
		},
		{
			Name:        "SESSION_COOKIE",
			Type:        "string",
			Default:     "ory_kratos_session",
			Description: "Session cookie name",
			ConfigPath:  "auth.session_cookie",
		},

		// Dev mode
		{
			Name:        "DEV_MODE",
			Type:        "bool",
			Default:     "false",
			Description: "Enable development mode",
			ConfigPath:  "dev_mode",
		},
	}
}

// GetFrameworkEnvVarByName returns a framework env var by its name (without prefix)
func GetFrameworkEnvVarByName(name string) *FrameworkEnvVar {
	for _, v := range FrameworkEnvVars() {
		if v.Name == name {
			return &v
		}
	}
	return nil
}
