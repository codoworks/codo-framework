package config

import "time"

// BaseMiddlewareConfig provides common configuration fields for all middleware
type BaseMiddlewareConfig struct {
	Enabled          bool `yaml:"enabled" default:"true"`
	DisableInDevMode bool `yaml:"disableInDevMode" default:"false"`
}

// IsEnabled checks if middleware should be enabled based on current mode
func (b *BaseMiddlewareConfig) IsEnabled(isDevMode bool) bool {
	if !b.Enabled {
		return false
	}

	if b.DisableInDevMode && isDevMode {
		return false
	}

	return true
}

// MiddlewareConfig holds configuration for all middleware
type MiddlewareConfig struct {
	Logger  LoggerMiddlewareConfig  `yaml:"logger"`
	CORS    CORSMiddlewareConfig    `yaml:"cors"`
	Timeout TimeoutMiddlewareConfig `yaml:"timeout"`
	Recover RecoverMiddlewareConfig `yaml:"recover"`
	Gzip    GzipMiddlewareConfig    `yaml:"gzip"`
	XSS     XSSMiddlewareConfig     `yaml:"xss"`
	Auth    AuthMiddlewareConfig    `yaml:"auth"`
	Health  HealthConfig            `yaml:"health"`
}

// LoggerMiddlewareConfig holds configuration for the logger middleware
type LoggerMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
	SkipPaths            []string `yaml:"skip_paths"`
}

// CORSMiddlewareConfig holds configuration for the CORS middleware
type CORSMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
	AllowOrigins         []string `yaml:"allow_origins"`
	AllowMethods         []string `yaml:"allow_methods"`
	AllowHeaders         []string `yaml:"allow_headers"`
	ExposeHeaders        []string `yaml:"expose_headers"`
	AllowCredentials     bool     `yaml:"allow_credentials"`
	MaxAge               int      `yaml:"max_age"`
}

// TimeoutMiddlewareConfig holds configuration for the timeout middleware
type TimeoutMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
	Duration             time.Duration `yaml:"duration"`
}

// RecoverMiddlewareConfig holds configuration for the recover middleware
type RecoverMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
}

// GzipMiddlewareConfig holds configuration for the gzip middleware
type GzipMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
	Level                int `yaml:"level"`    // 1-9, default 5
	MinSize              int `yaml:"min_size"` // minimum size in bytes, default 1024
}

// XSSMiddlewareConfig holds configuration for the XSS protection middleware
type XSSMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
	XSSProtection        string `yaml:"xss_protection"`       // X-XSS-Protection header
	ContentTypeNosniff   string `yaml:"content_type_nosniff"` // X-Content-Type-Options header
	XFrameOptions        string `yaml:"x_frame_options"`      // X-Frame-Options header
	HSTSMaxAge           int    `yaml:"hsts_max_age"`         // Strict-Transport-Security max-age
}

// AuthMiddlewareConfig holds configuration for the authentication middleware
type AuthMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
	SkipPaths            []string          `yaml:"skip_paths"`
	DevMode              bool              `yaml:"dev_mode"`
	DevIdentity          *DevIdentityConfig `yaml:"dev_identity"`
}

// DevIdentityConfig holds configuration for dev mode identity bypass
type DevIdentityConfig struct {
	ID     string         `yaml:"id"`
	Traits map[string]any `yaml:"traits"`
}

// HealthConfig holds configuration for health check endpoints
type HealthConfig struct {
	Enabled           bool `yaml:"enabled" default:"true"`
	ShowDetailsInProd bool `yaml:"show_details_in_prod" default:"false"`
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		Logger: LoggerMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          true,
				DisableInDevMode: false,
			},
			SkipPaths: []string{"/health", "/health/live", "/health/ready"},
		},
		CORS: CORSMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          true,
				DisableInDevMode: false,
			},
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
			ExposeHeaders:    []string{"X-Request-ID"},
			AllowCredentials: false,
			MaxAge:           86400, // 24 hours
		},
		Timeout: TimeoutMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          true,
				DisableInDevMode: false,
			},
			Duration: 60 * time.Second,
		},
		Recover: RecoverMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          true,
				DisableInDevMode: false,
			},
		},
		Gzip: GzipMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          true,
				DisableInDevMode: false,
			},
			Level:   5,
			MinSize: 1024,
		},
		XSS: XSSMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          true,
				DisableInDevMode: false,
			},
			XSSProtection:      "1; mode=block",
			ContentTypeNosniff: "nosniff",
			XFrameOptions:      "SAMEORIGIN",
			HSTSMaxAge:         31536000, // 1 year
		},
		Auth: AuthMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          true, // ENABLED BY DEFAULT
				DisableInDevMode: false,
			},
			SkipPaths: []string{"/health"},
			DevMode:   false,
		},
		Health: HealthConfig{
			Enabled:           true,  // ENABLED BY DEFAULT
			ShowDetailsInProd: false, // Details only in dev mode
		},
	}
}
