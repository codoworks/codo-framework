package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, "codo-app", cfg.Service.Name)
	assert.Equal(t, 8081, cfg.Server.PublicPort)
	assert.Equal(t, "postgres", cfg.Database.Driver)
}

func TestDefaultConfig_EqualsNewWithDefaults(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg2 := NewWithDefaults()

	assert.Equal(t, cfg1.Service, cfg2.Service)
	assert.Equal(t, cfg1.Server, cfg2.Server)
	assert.Equal(t, cfg1.Database, cfg2.Database)
	assert.Equal(t, cfg1.Auth, cfg2.Auth)
	assert.Equal(t, cfg1.Features, cfg2.Features)
	assert.Equal(t, cfg1.DevMode, cfg2.DevMode)
}

func TestConfig_ApplyDefaults_Empty(t *testing.T) {
	cfg := &Config{}

	cfg.ApplyDefaults()

	assert.Equal(t, "codo-app", cfg.Service.Name)
	assert.Equal(t, "0.0.0", cfg.Service.Version)
	assert.Equal(t, "development", cfg.Service.Environment)
	assert.Equal(t, 8081, cfg.Server.PublicPort)
	assert.Equal(t, 8080, cfg.Server.ProtectedPort)
	assert.Equal(t, 8079, cfg.Server.HiddenPort)
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "ory_kratos_session", cfg.Auth.SessionCookie)
}

func TestConfig_ApplyDefaults_Partial(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Name: "my-app",
		},
		Server: ServerConfig{
			PublicPort: 9000,
		},
	}

	cfg.ApplyDefaults()

	// Provided values should be preserved
	assert.Equal(t, "my-app", cfg.Service.Name)
	assert.Equal(t, 9000, cfg.Server.PublicPort)

	// Missing values should get defaults
	assert.Equal(t, "0.0.0", cfg.Service.Version)
	assert.Equal(t, 8080, cfg.Server.ProtectedPort)
	assert.Equal(t, "postgres", cfg.Database.Driver)
}

func TestConfig_ApplyDefaults_PreservesNonZeroValues(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Name:        "custom-app",
			Version:     "1.0.0",
			Environment: "production",
		},
		Server: ServerConfig{
			PublicPort:    3000,
			ProtectedPort: 3001,
			HiddenPort:    3002,
			ReadTimeout:   Duration(10 * time.Second),
			WriteTimeout:  Duration(10 * time.Second),
			IdleTimeout:   Duration(30 * time.Second),
			ShutdownGrace: Duration(5 * time.Second),
		},
		Database: DatabaseConfig{
			Driver:          "mysql",
			Host:            "db.example.com",
			Port:            3306,
			Name:            "mydb",
			User:            "admin",
			SSLMode:         "require",
			MaxOpenConns:    50,
			MaxIdleConns:    10,
			ConnMaxLifetime: Duration(10 * time.Minute),
		},
		Auth: AuthConfig{
			KratosPublicURL: "https://auth.example.com",
			KratosAdminURL:  "https://auth-admin.example.com",
			KetoReadURL:     "https://keto.example.com:4466",
			KetoWriteURL:    "https://keto.example.com:4467",
			SessionCookie:   "my_session",
		},
	}

	cfg.ApplyDefaults()

	// All values should be preserved
	assert.Equal(t, "custom-app", cfg.Service.Name)
	assert.Equal(t, "1.0.0", cfg.Service.Version)
	assert.Equal(t, "production", cfg.Service.Environment)
	assert.Equal(t, 3000, cfg.Server.PublicPort)
	assert.Equal(t, 3001, cfg.Server.ProtectedPort)
	assert.Equal(t, 3002, cfg.Server.HiddenPort)
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.Equal(t, "mydb", cfg.Database.Name)
	assert.Equal(t, "admin", cfg.Database.User)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, 50, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, 10*time.Minute, cfg.Database.ConnMaxLifetime.Duration())
	assert.Equal(t, "https://auth.example.com", cfg.Auth.KratosPublicURL)
	assert.Equal(t, "https://auth-admin.example.com", cfg.Auth.KratosAdminURL)
	assert.Equal(t, "my_session", cfg.Auth.SessionCookie)
}

func TestConfig_ApplyDefaults_FeaturesNil(t *testing.T) {
	cfg := &Config{
		Features: FeaturesConfig{
			DisabledFeatures: nil,
		},
	}

	cfg.ApplyDefaults()

	assert.NotNil(t, cfg.Features.DisabledFeatures)
	assert.Empty(t, cfg.Features.DisabledFeatures)
}

func TestConfig_ApplyDefaults_FeaturesPreserved(t *testing.T) {
	cfg := &Config{
		Features: FeaturesConfig{
			DisabledFeatures: []string{"database", "redis"},
		},
	}

	cfg.ApplyDefaults()

	// Existing disabled features should be preserved
	assert.Equal(t, []string{"database", "redis"}, cfg.Features.DisabledFeatures)
}

func TestConfig_ApplyDefaults_FeaturesEmptySlice(t *testing.T) {
	cfg := &Config{
		Features: FeaturesConfig{
			DisabledFeatures: []string{},
		},
	}

	cfg.ApplyDefaults()

	// Empty slice should be preserved (not nil)
	assert.NotNil(t, cfg.Features.DisabledFeatures)
	assert.Empty(t, cfg.Features.DisabledFeatures)
}

func TestConfig_ApplyDefaults_Timeouts(t *testing.T) {
	cfg := &Config{}

	cfg.ApplyDefaults()

	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout.Duration())
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout.Duration())
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout.Duration())
	assert.Equal(t, 20*time.Second, cfg.Server.ShutdownGrace.Duration())
}

func TestConfig_ApplyDefaults_DevModeNotChanged(t *testing.T) {
	cfg := &Config{
		DevMode: true,
	}

	cfg.ApplyDefaults()

	// DevMode should remain true (ApplyDefaults doesn't touch it)
	assert.True(t, cfg.DevMode)
}

func TestConfig_ApplyDefaults_CalledMultipleTimes(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Name: "my-app",
		},
	}

	cfg.ApplyDefaults()
	cfg.ApplyDefaults()
	cfg.ApplyDefaults()

	// Should be idempotent
	assert.Equal(t, "my-app", cfg.Service.Name)
	assert.Equal(t, "0.0.0", cfg.Service.Version)
}
