package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Ensure no config files exist in default locations
	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "development", cfg.Service.Environment)
	assert.Equal(t, 8081, cfg.Server.PublicPort)
	assert.Equal(t, 8080, cfg.Server.ProtectedPort)
	assert.Equal(t, 8079, cfg.Server.HiddenPort)
	assert.Equal(t, "postgres", cfg.Database.Driver)
}

func TestLoad_YAMLFile(t *testing.T) {
	yamlContent := `
service:
  environment: staging
server:
  public_port: 9000
  protected_port: 9001
  hidden_port: 9002
database:
  driver: mysql
  host: db.example.com
  port: 3306
  name: myapp
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	assert.Equal(t, "staging", cfg.Service.Environment)
	assert.Equal(t, 9000, cfg.Server.PublicPort)
	assert.Equal(t, 9001, cfg.Server.ProtectedPort)
	assert.Equal(t, 9002, cfg.Server.HiddenPort)
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, "myapp", cfg.Database.Name)
}

func TestLoad_EnvOverride(t *testing.T) {
	yamlContent := `
service:
  environment: staging
server:
  public_port: 9000
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Set env var to override
	t.Setenv("CODO_SERVICE_ENVIRONMENT", "production")

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	// Env should override YAML
	assert.Equal(t, "production", cfg.Service.Environment)
	// YAML should override defaults
	assert.Equal(t, 9000, cfg.Server.PublicPort)
	// Defaults should be used for unset values
	assert.Equal(t, 8080, cfg.Server.ProtectedPort)
}

func TestLoad_EnvOverride_AllFields(t *testing.T) {
	t.Setenv("CODO_SERVICE_ENVIRONMENT", "production")
	t.Setenv("CODO_PUBLIC_PORT", "3000")
	t.Setenv("CODO_PROTECTED_PORT", "3001")
	t.Setenv("CODO_HIDDEN_PORT", "3002")
	t.Setenv("CODO_DB_DRIVER", "mysql")
	t.Setenv("CODO_DB_HOST", "mysql.local")
	t.Setenv("CODO_DB_PORT", "3306")
	t.Setenv("CODO_DB_NAME", "envdb")
	t.Setenv("CODO_DB_USER", "envuser")
	t.Setenv("CODO_DB_PASSWORD", "envpass")
	t.Setenv("CODO_DB_SSL_MODE", "require")
	t.Setenv("CODO_KRATOS_PUBLIC_URL", "https://kratos.local")
	t.Setenv("CODO_KRATOS_ADMIN_URL", "https://kratos-admin.local")
	t.Setenv("CODO_KETO_READ_URL", "https://keto-read.local")
	t.Setenv("CODO_KETO_WRITE_URL", "https://keto-write.local")
	t.Setenv("CODO_SESSION_COOKIE", "env_session")
	t.Setenv("CODO_DEV_MODE", "true")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "production", cfg.Service.Environment)
	assert.Equal(t, 3000, cfg.Server.PublicPort)
	assert.Equal(t, 3001, cfg.Server.ProtectedPort)
	assert.Equal(t, 3002, cfg.Server.HiddenPort)
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.Equal(t, "mysql.local", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, "envdb", cfg.Database.Name)
	assert.Equal(t, "envuser", cfg.Database.User)
	assert.Equal(t, "envpass", cfg.Database.Password)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "https://kratos.local", cfg.Auth.KratosPublicURL)
	assert.Equal(t, "https://kratos-admin.local", cfg.Auth.KratosAdminURL)
	assert.Equal(t, "https://keto-read.local", cfg.Auth.KetoReadURL)
	assert.Equal(t, "https://keto-write.local", cfg.Auth.KetoWriteURL)
	assert.Equal(t, "env_session", cfg.Auth.SessionCookie)
	assert.True(t, cfg.DevMode)
}

func TestLoad_InvalidYAML(t *testing.T) {
	yamlContent := `
service:
  name: [invalid yaml
  version: 1.0
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	_, err = LoadFromFile(tmpFile.Name())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open config file")
}

func TestLoad_PartialConfig(t *testing.T) {
	yamlContent := `
service:
  environment: staging
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	// Provided value should be used
	assert.Equal(t, "staging", cfg.Service.Environment)
	// Defaults should fill in the rest
	assert.Equal(t, 8081, cfg.Server.PublicPort)
	assert.Equal(t, "postgres", cfg.Database.Driver)
}

func TestLoadFromFile(t *testing.T) {
	yamlContent := `
service:
  environment: production
database:
  driver: sqlite
  name: test.db
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	assert.Equal(t, "production", cfg.Service.Environment)
	assert.Equal(t, "sqlite", cfg.Database.Driver)
	assert.Equal(t, "test.db", cfg.Database.Name)
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/does/not/exist.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open config file")
}

func TestLoadFromReader(t *testing.T) {
	yamlContent := `
service:
  environment: production
server:
  public_port: 5000
`
	reader := strings.NewReader(yamlContent)

	cfg, err := LoadFromReader(reader)

	require.NoError(t, err)
	assert.Equal(t, "production", cfg.Service.Environment)
	assert.Equal(t, 5000, cfg.Server.PublicPort)
}

func TestLoadFromReader_Invalid(t *testing.T) {
	yamlContent := `
service:
  name: [broken
`
	reader := strings.NewReader(yamlContent)

	_, err := LoadFromReader(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestLoadFromReader_Empty(t *testing.T) {
	reader := strings.NewReader("")

	cfg, err := LoadFromReader(reader)

	require.NoError(t, err)
	// Should return defaults
	assert.Equal(t, "development", cfg.Service.Environment)
}

func TestLoadFromReader_ValidationFailure(t *testing.T) {
	// Use duplicate ports which can't be fixed by defaults
	yamlContent := `
server:
  public_port: 8080
  protected_port: 8080
  hidden_port: 8079
`
	reader := strings.NewReader(yamlContent)

	_, err := LoadFromReader(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server ports must be unique")
}

func TestLoad_DevMode_FromEnv(t *testing.T) {
	t.Setenv("CODO_DEV_MODE", "true")

	cfg, err := Load()

	require.NoError(t, err)
	assert.True(t, cfg.DevMode)
	assert.True(t, cfg.IsDevMode())
}

func TestLoad_DevMode_False(t *testing.T) {
	t.Setenv("CODO_DEV_MODE", "false")

	cfg, err := Load()

	require.NoError(t, err)
	assert.False(t, cfg.DevMode)
}

func TestLoad_DevMode_NotSet(t *testing.T) {
	cfg, err := Load()

	require.NoError(t, err)
	assert.False(t, cfg.DevMode)
}

func TestLoad_Features_FromEnv(t *testing.T) {
	t.Setenv("DISABLE_FEATURES", "database,redis")

	cfg, err := Load()

	require.NoError(t, err)
	assert.False(t, cfg.Features.IsEnabled("database"))
	assert.False(t, cfg.Features.IsEnabled("redis"))
	assert.True(t, cfg.Features.IsEnabled("kratos"))
}

func TestGetDefaultConfigPaths(t *testing.T) {
	paths := GetDefaultConfigPaths()

	assert.Contains(t, paths, "config/app.yaml")
	assert.Contains(t, paths, "app.yaml")
	assert.Contains(t, paths, "config.yaml")
	assert.Len(t, paths, 3)
}

func TestGetDefaultConfigPaths_IsCopy(t *testing.T) {
	paths := GetDefaultConfigPaths()
	paths[0] = "modified"

	original := GetDefaultConfigPaths()
	assert.Equal(t, "config/app.yaml", original[0])
}

func TestLoad_YAMLWithValidation(t *testing.T) {
	// Valid config
	yamlContent := `
service:
  environment: production
server:
  public_port: 8081
  protected_port: 8080
  hidden_port: 8079
database:
  driver: postgres
  host: localhost
  port: 5432
  name: testdb
auth:
  session_cookie: test_session
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	assert.Equal(t, "production", cfg.Service.Environment)
}

func TestLoad_YAMLWithInvalidPorts(t *testing.T) {
	yamlContent := `
service:
  environment: production
server:
  public_port: 8080
  protected_port: 8080
  hidden_port: 8079
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	_, err = LoadFromFile(tmpFile.Name())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server ports must be unique")
}

func TestLoad_EnvOverridesYAMLBeforeValidation(t *testing.T) {
	// YAML has invalid duplicate ports, but env fixes it
	yamlContent := `
service:
  environment: production
server:
  public_port: 8080
  protected_port: 8080
  hidden_port: 8079
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Fix the duplicate port via env
	t.Setenv("CODO_PROTECTED_PORT", "8081")

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.PublicPort)
	assert.Equal(t, 8081, cfg.Server.ProtectedPort)
}

func TestLoad_EmptyYAMLFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	// Should use all defaults
	assert.Equal(t, "development", cfg.Service.Environment)
	assert.Equal(t, 8081, cfg.Server.PublicPort)
}

func TestLoad_YAMLWithComments(t *testing.T) {
	yamlContent := `
# This is a comment
service:
  environment: staging  # inline comment
  # Another comment
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadFromFile(tmpFile.Name())

	require.NoError(t, err)
	assert.Equal(t, "staging", cfg.Service.Environment)
}

func TestLoad_FromDefaultLocation_ConfigYAML(t *testing.T) {
	// Create a config.yaml file in current directory (one of the default locations)
	yamlContent := `
service:
  environment: production
`
	err := os.WriteFile("config.yaml", []byte(yamlContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "production", cfg.Service.Environment)
}

func TestLoad_FromDefaultLocation_AppYAML(t *testing.T) {
	// Create app.yaml file in current directory
	yamlContent := `
service:
  environment: staging
`
	err := os.WriteFile("app.yaml", []byte(yamlContent), 0644)
	require.NoError(t, err)
	defer os.Remove("app.yaml")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "staging", cfg.Service.Environment)
}

func TestLoad_FromDefaultLocation_InvalidYAML(t *testing.T) {
	// Create a config.yaml file with invalid YAML
	yamlContent := `
service:
  name: [broken yaml
`
	err := os.WriteFile("config.yaml", []byte(yamlContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	_, err = Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestLoad_FromDefaultLocation_ValidationFailure(t *testing.T) {
	// Create a config.yaml file with validation errors
	yamlContent := `
server:
  public_port: 8080
  protected_port: 8080
  hidden_port: 8079
`
	err := os.WriteFile("config.yaml", []byte(yamlContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	_, err = Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server ports must be unique")
}

func TestLoad_PriorityOrder(t *testing.T) {
	// config/app.yaml should be checked first, then app.yaml, then config.yaml
	// Create config.yaml
	err := os.WriteFile("config.yaml", []byte("service:\n  environment: production"), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Create app.yaml (should take precedence)
	err = os.WriteFile("app.yaml", []byte("service:\n  environment: staging"), 0644)
	require.NoError(t, err)
	defer os.Remove("app.yaml")

	cfg, err := Load()

	require.NoError(t, err)
	// app.yaml should be loaded (comes before config.yaml in priority)
	assert.Equal(t, "staging", cfg.Service.Environment)
}

func TestLoad_ConfigAppYAML_Priority(t *testing.T) {
	// Create config/app.yaml (highest priority)
	err := os.MkdirAll("config", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("config")

	err = os.WriteFile("config/app.yaml", []byte("service:\n  environment: production"), 0644)
	require.NoError(t, err)

	// Also create app.yaml (lower priority)
	err = os.WriteFile("app.yaml", []byte("service:\n  environment: staging"), 0644)
	require.NoError(t, err)
	defer os.Remove("app.yaml")

	cfg, err := Load()

	require.NoError(t, err)
	// config/app.yaml should be loaded (highest priority)
	assert.Equal(t, "production", cfg.Service.Environment)
}
