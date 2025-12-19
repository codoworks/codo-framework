package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()

	assert.Equal(t, 8081, cfg.PublicPort)
	assert.Equal(t, 8080, cfg.ProtectedPort)
	assert.Equal(t, 8079, cfg.HiddenPort)
	assert.Equal(t, 30*time.Second, cfg.ReadTimeout.Duration())
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout.Duration())
	assert.Equal(t, 60*time.Second, cfg.IdleTimeout.Duration())
	assert.Equal(t, 20*time.Second, cfg.ShutdownGrace.Duration())
}

func TestServerConfig_Validate_Valid(t *testing.T) {
	cfg := ServerConfig{
		PublicPort:    9000,
		ProtectedPort: 9001,
		HiddenPort:    9002,
		ReadTimeout:   Duration(10 * time.Second),
		WriteTimeout:  Duration(10 * time.Second),
		IdleTimeout:   Duration(30 * time.Second),
		ShutdownGrace: Duration(5 * time.Second),
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestServerConfig_Validate_DefaultsValid(t *testing.T) {
	cfg := DefaultServerConfig()

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestServerConfig_Validate_InvalidPublicPort_Zero(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.PublicPort = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.public_port must be between 1 and 65535")
}

func TestServerConfig_Validate_InvalidPublicPort_TooHigh(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.PublicPort = 65536

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.public_port must be between 1 and 65535")
}

func TestServerConfig_Validate_InvalidPublicPort_Negative(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.PublicPort = -1

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.public_port must be between 1 and 65535")
}

func TestServerConfig_Validate_InvalidProtectedPort_Zero(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ProtectedPort = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.protected_port must be between 1 and 65535")
}

func TestServerConfig_Validate_InvalidProtectedPort_TooHigh(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ProtectedPort = 70000

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.protected_port must be between 1 and 65535")
}

func TestServerConfig_Validate_InvalidHiddenPort_Zero(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.HiddenPort = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.hidden_port must be between 1 and 65535")
}

func TestServerConfig_Validate_InvalidHiddenPort_TooHigh(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.HiddenPort = 100000

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.hidden_port must be between 1 and 65535")
}

func TestServerConfig_Validate_DuplicatePorts_PublicProtected(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.PublicPort = 8080
	cfg.ProtectedPort = 8080

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server ports must be unique")
	assert.Contains(t, err.Error(), "public_port and protected_port")
}

func TestServerConfig_Validate_DuplicatePorts_PublicHidden(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.PublicPort = 8079
	cfg.HiddenPort = 8079

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server ports must be unique")
	assert.Contains(t, err.Error(), "public_port and hidden_port")
}

func TestServerConfig_Validate_DuplicatePorts_ProtectedHidden(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ProtectedPort = 8079
	cfg.HiddenPort = 8079

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server ports must be unique")
	assert.Contains(t, err.Error(), "protected_port and hidden_port")
}

func TestServerConfig_Validate_ZeroReadTimeout(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ReadTimeout = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.read_timeout must be positive")
}

func TestServerConfig_Validate_NegativeReadTimeout(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ReadTimeout = Duration(-1 * time.Second)

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.read_timeout must be positive")
}

func TestServerConfig_Validate_ZeroWriteTimeout(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.WriteTimeout = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.write_timeout must be positive")
}

func TestServerConfig_Validate_NegativeWriteTimeout(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.WriteTimeout = Duration(-5 * time.Second)

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.write_timeout must be positive")
}

func TestServerConfig_Validate_ZeroIdleTimeout(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.IdleTimeout = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.idle_timeout must be positive")
}

func TestServerConfig_Validate_NegativeIdleTimeout(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.IdleTimeout = Duration(-1 * time.Minute)

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.idle_timeout must be positive")
}

func TestServerConfig_Validate_ZeroShutdownGrace(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ShutdownGrace = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.shutdown_grace must be positive")
}

func TestServerConfig_Validate_NegativeShutdownGrace(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ShutdownGrace = Duration(-10 * time.Second)

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.shutdown_grace must be positive")
}

func TestServerConfig_PublicAddr(t *testing.T) {
	cfg := DefaultServerConfig()

	addr := cfg.PublicAddr()

	assert.Equal(t, ":8081", addr)
}

func TestServerConfig_PublicAddr_Custom(t *testing.T) {
	cfg := ServerConfig{PublicPort: 3000}

	addr := cfg.PublicAddr()

	assert.Equal(t, ":3000", addr)
}

func TestServerConfig_ProtectedAddr(t *testing.T) {
	cfg := DefaultServerConfig()

	addr := cfg.ProtectedAddr()

	assert.Equal(t, ":8080", addr)
}

func TestServerConfig_ProtectedAddr_Custom(t *testing.T) {
	cfg := ServerConfig{ProtectedPort: 4000}

	addr := cfg.ProtectedAddr()

	assert.Equal(t, ":4000", addr)
}

func TestServerConfig_HiddenAddr(t *testing.T) {
	cfg := DefaultServerConfig()

	addr := cfg.HiddenAddr()

	assert.Equal(t, ":8079", addr)
}

func TestServerConfig_HiddenAddr_Custom(t *testing.T) {
	cfg := ServerConfig{HiddenPort: 5000}

	addr := cfg.HiddenAddr()

	assert.Equal(t, ":5000", addr)
}

func TestServerConfig_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*ServerConfig)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid minimum port",
			modify:  func(c *ServerConfig) { c.PublicPort = 1 },
			wantErr: false,
		},
		{
			name:    "valid maximum port",
			modify:  func(c *ServerConfig) { c.PublicPort = 65535 },
			wantErr: false,
		},
		{
			name:    "all ports same",
			modify:  func(c *ServerConfig) { c.PublicPort = 8080; c.ProtectedPort = 8080; c.HiddenPort = 8080 },
			wantErr: true,
			errMsg:  "server ports must be unique",
		},
		{
			name:    "minimum valid timeout",
			modify:  func(c *ServerConfig) { c.ReadTimeout = Duration(1 * time.Nanosecond) },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultServerConfig()
			tt.modify(&cfg)

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
