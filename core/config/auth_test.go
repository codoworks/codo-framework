package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAuthConfig(t *testing.T) {
	cfg := DefaultAuthConfig()

	assert.Equal(t, "http://localhost:4433", cfg.KratosPublicURL)
	assert.Equal(t, "http://localhost:4434", cfg.KratosAdminURL)
	assert.Equal(t, "http://localhost:4466", cfg.KetoReadURL)
	assert.Equal(t, "http://localhost:4467", cfg.KetoWriteURL)
	assert.Equal(t, "ory_kratos_session", cfg.SessionCookie)
}

func TestAuthConfig_Validate_Valid(t *testing.T) {
	cfg := DefaultAuthConfig()

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestAuthConfig_Validate_CustomValues(t *testing.T) {
	cfg := AuthConfig{
		KratosPublicURL: "https://auth.example.com",
		KratosAdminURL:  "https://auth-admin.example.com",
		KetoReadURL:     "https://keto.example.com:4466",
		KetoWriteURL:    "https://keto.example.com:4467",
		SessionCookie:   "my_session",
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestAuthConfig_Validate_EmptySessionCookie(t *testing.T) {
	cfg := DefaultAuthConfig()
	cfg.SessionCookie = ""

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth.session_cookie is required")
}

func TestAuthConfig_Validate_EmptyURLs(t *testing.T) {
	// Empty URLs are allowed - they may be set via env vars or not used
	cfg := AuthConfig{
		KratosPublicURL: "",
		KratosAdminURL:  "",
		KetoReadURL:     "",
		KetoWriteURL:    "",
		SessionCookie:   "session",
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestAuthConfig_Validate_OnlySessionCookie(t *testing.T) {
	cfg := AuthConfig{
		SessionCookie: "my_cookie",
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestAuthConfig_Validate_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AuthConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid defaults",
			cfg:     DefaultAuthConfig(),
			wantErr: false,
		},
		{
			name: "valid custom session cookie",
			cfg: AuthConfig{
				SessionCookie: "custom_session_cookie",
			},
			wantErr: false,
		},
		{
			name: "empty session cookie",
			cfg: AuthConfig{
				KratosPublicURL: "http://localhost:4433",
				SessionCookie:   "",
			},
			wantErr: true,
			errMsg:  "auth.session_cookie is required",
		},
		{
			name: "whitespace only session cookie is valid",
			cfg: AuthConfig{
				SessionCookie: "   ",
			},
			wantErr: false, // Whitespace is technically valid, validation doesn't trim
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

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
