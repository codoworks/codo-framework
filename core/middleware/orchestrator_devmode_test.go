package middleware

import (
	"testing"

	"github.com/codoworks/codo-framework/core/config"
)

func TestOrchestrator_shouldEnableBasedOnMode(t *testing.T) {
	tests := []struct {
		name      string
		cfg       any
		devMode   bool
		want      bool
	}{
		{
			name: "nil config returns true",
			cfg:  nil,
			devMode: false,
			want: true,
		},
		{
			name: "enabled=true, disableInDevMode=false, devMode=false -> true",
			cfg: &config.LoggerMiddlewareConfig{
				BaseMiddlewareConfig: config.BaseMiddlewareConfig{
					Enabled:          true,
					DisableInDevMode: false,
				},
			},
			devMode: false,
			want:    true,
		},
		{
			name: "enabled=true, disableInDevMode=false, devMode=true -> true",
			cfg: &config.LoggerMiddlewareConfig{
				BaseMiddlewareConfig: config.BaseMiddlewareConfig{
					Enabled:          true,
					DisableInDevMode: false,
				},
			},
			devMode: true,
			want:    true,
		},
		{
			name: "enabled=true, disableInDevMode=true, devMode=false -> true",
			cfg: &config.CORSMiddlewareConfig{
				BaseMiddlewareConfig: config.BaseMiddlewareConfig{
					Enabled:          true,
					DisableInDevMode: true,
				},
			},
			devMode: false,
			want:    true,
		},
		{
			name: "enabled=true, disableInDevMode=true, devMode=true -> false",
			cfg: &config.CORSMiddlewareConfig{
				BaseMiddlewareConfig: config.BaseMiddlewareConfig{
					Enabled:          true,
					DisableInDevMode: true,
				},
			},
			devMode: true,
			want:    false,
		},
		{
			name: "enabled=false, disableInDevMode=false, devMode=false -> false",
			cfg: &config.GzipMiddlewareConfig{
				BaseMiddlewareConfig: config.BaseMiddlewareConfig{
					Enabled:          false,
					DisableInDevMode: false,
				},
			},
			devMode: false,
			want:    false,
		},
		{
			name: "enabled=false, disableInDevMode=true, devMode=true -> false",
			cfg: &config.TimeoutMiddlewareConfig{
				BaseMiddlewareConfig: config.BaseMiddlewareConfig{
					Enabled:          false,
					DisableInDevMode: true,
				},
			},
			devMode: true,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Orchestrator{
				config: &config.Config{
					DevMode: tt.devMode,
				},
			}

			got := o.shouldEnableBasedOnMode(tt.cfg)
			if got != tt.want {
				t.Errorf("shouldEnableBasedOnMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
