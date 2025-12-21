package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithDefaults(t *testing.T) {
	cfg := NewWithDefaults()

	assert.NotNil(t, cfg)
	assert.Equal(t, "development", cfg.Service.Environment)
	assert.Equal(t, 8081, cfg.Server.PublicPort)
	assert.Equal(t, 8080, cfg.Server.ProtectedPort)
	assert.Equal(t, 8079, cfg.Server.HiddenPort)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout.Duration())
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "http://localhost:4433", cfg.Auth.KratosPublicURL)
	assert.Equal(t, "ory_kratos_session", cfg.Auth.SessionCookie)
	assert.Empty(t, cfg.Features.DisabledFeatures)
	assert.False(t, cfg.DevMode)
}

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := NewWithDefaults()

	err := cfg.Validate()

	assert.NoError(t, err)
}


func TestConfig_Validate_Invalid_Server(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.Server.PublicPort = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.public_port")
}

func TestConfig_Validate_Invalid_Database(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.Database.Name = ""

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.name is required")
}

func TestConfig_Validate_Invalid_Auth(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.Auth.SessionCookie = ""

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth.session_cookie is required")
}

func TestConfig_IsDevMode(t *testing.T) {
	cfg := NewWithDefaults()
	assert.False(t, cfg.IsDevMode())

	cfg.DevMode = true
	assert.True(t, cfg.IsDevMode())
}

func TestConfig_IsProd(t *testing.T) {
	cfg := NewWithDefaults()
	assert.True(t, cfg.IsProd())

	cfg.DevMode = true
	assert.False(t, cfg.IsProd())
}

func TestConfig_Clone(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.Service.Environment = "production"
	cfg.DevMode = true
	cfg.Features.Disable("database")

	clone := cfg.Clone()

	// Clone should have same values
	assert.Equal(t, "production", clone.Service.Environment)
	assert.True(t, clone.DevMode)
	assert.False(t, clone.Features.IsEnabled("database"))

	// Modifying clone should not affect original
	clone.Service.Environment = "staging"
	clone.DevMode = false
	clone.Features.Enable("database")

	assert.Equal(t, "production", cfg.Service.Environment)
	assert.True(t, cfg.DevMode)
	assert.False(t, cfg.Features.IsEnabled("database"))
}

func TestConfig_Clone_FeaturesDeepCopy(t *testing.T) {
	cfg := NewWithDefaults()
	cfg.Features.DisabledFeatures = []string{"a", "b", "c"}

	clone := cfg.Clone()

	// Modify clone's features
	clone.Features.DisabledFeatures[0] = "modified"

	// Original should be unchanged
	assert.Equal(t, "a", cfg.Features.DisabledFeatures[0])
}

func TestConfig_Clone_EmptyFeatures(t *testing.T) {
	cfg := NewWithDefaults()

	clone := cfg.Clone()

	assert.NotNil(t, clone.Features.DisabledFeatures)
	assert.Empty(t, clone.Features.DisabledFeatures)
}

func TestConfig_Validate_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid defaults",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name:    "invalid port",
			modify:  func(c *Config) { c.Server.PublicPort = 0 },
			wantErr: true,
			errMsg:  "server.public_port",
		},
		{
			name:    "missing db name",
			modify:  func(c *Config) { c.Database.Name = "" },
			wantErr: true,
			errMsg:  "database.name",
		},
		{
			name:    "duplicate ports",
			modify:  func(c *Config) { c.Server.PublicPort = 8080; c.Server.ProtectedPort = 8080 },
			wantErr: true,
			errMsg:  "server ports must be unique",
		},
		{
			name:    "invalid driver",
			modify:  func(c *Config) { c.Database.Driver = "oracle" },
			wantErr: true,
			errMsg:  "database.driver",
		},
		{
			name:    "missing session cookie",
			modify:  func(c *Config) { c.Auth.SessionCookie = "" },
			wantErr: true,
			errMsg:  "auth.session_cookie",
		},
		{
			name:    "zero read timeout",
			modify:  func(c *Config) { c.Server.ReadTimeout = 0 },
			wantErr: true,
			errMsg:  "server.read_timeout",
		},
		{
			name:    "negative max open conns",
			modify:  func(c *Config) { c.Database.MaxOpenConns = -1 },
			wantErr: true,
			errMsg:  "database.max_open_conns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewWithDefaults()
			tt.modify(cfg)

			err := cfg.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_FullIntegration(t *testing.T) {
	cfg := NewWithDefaults()

	// Modify various settings
	cfg.Service.Environment = "production"
	cfg.Server.PublicPort = 3000
	cfg.Server.ProtectedPort = 3001
	cfg.Server.HiddenPort = 3002
	cfg.Database.Driver = "mysql"
	cfg.Database.Host = "db.example.com"
	cfg.Database.Port = 3306
	cfg.Database.Name = "myapp"
	cfg.Auth.SessionCookie = "my_session"
	cfg.Features.Disable("redis")
	cfg.SetDevMode(true)

	// Validate
	err := cfg.Validate()
	require.NoError(t, err)

	// Check values
	assert.Equal(t, "production", cfg.Service.Environment)
	assert.Equal(t, 3000, cfg.Server.PublicPort)
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.False(t, cfg.Features.IsEnabled("redis"))
	assert.True(t, cfg.IsDevMode())
	assert.False(t, cfg.IsProd())

	// Clone and verify
	clone := cfg.Clone()
	assert.Equal(t, cfg.Service.Environment, clone.Service.Environment)
	assert.Equal(t, cfg.DevMode, clone.DevMode)
}

func TestNewWithDefaults_MultipleCallsIndependent(t *testing.T) {
	cfg1 := NewWithDefaults()
	cfg2 := NewWithDefaults()

	cfg1.Service.Environment = "production"

	assert.Equal(t, "production", cfg1.Service.Environment)
	assert.Equal(t, "development", cfg2.Service.Environment)
}

func TestConfig_ValidateOrder(t *testing.T) {
	// Test that validation runs in a specific order
	// Service first, then Server, then Database, then Auth
	cfg := NewWithDefaults()
	cfg.Server.PublicPort = 0   // Will fail first
	cfg.Database.Name = ""      // Would fail second
	cfg.Auth.SessionCookie = "" // Would fail third

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.public_port") // Should be the first error
}
