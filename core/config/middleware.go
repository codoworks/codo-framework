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
	Logger     LoggerMiddlewareConfig     `yaml:"logger"`
	CORS       CORSMiddlewareConfig       `yaml:"cors"`
	Timeout    TimeoutMiddlewareConfig    `yaml:"timeout"`
	Recover    RecoverMiddlewareConfig    `yaml:"recover"`
	Gzip       GzipMiddlewareConfig       `yaml:"gzip"`
	XSS        XSSMiddlewareConfig        `yaml:"xss"`
	Auth       AuthMiddlewareConfig       `yaml:"auth"`
	Health     HealthConfig               `yaml:"health"`
	Pagination PaginationMiddlewareConfig `yaml:"pagination"`
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
	SkipPaths            []string           `yaml:"skip_paths"`
	DevMode              bool               `yaml:"dev_mode"`         // Enables verbose logging (user ID/name)
	DevBypassAuth        bool               `yaml:"dev_bypass_auth"`  // Skip real auth, use DevIdentity
	DevIdentity          *DevIdentityConfig `yaml:"dev_identity"`
	CacheEnabled         bool               `yaml:"cache_enabled"`    // Enable session caching
	CacheTTL             time.Duration      `yaml:"cache_ttl"`        // Cache time-to-live
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

// PaginationMiddlewareConfig holds configuration for the pagination middleware
type PaginationMiddlewareConfig struct {
	BaseMiddlewareConfig `yaml:",inline"`
	DefaultPageSize      int                    `yaml:"default_page_size"` // Default items per page (default: 20)
	MaxPageSize          int                    `yaml:"max_page_size"`     // Maximum allowed items per page (default: 100)
	DefaultType          string                 `yaml:"default_type"`      // Default pagination type: "offset" or "cursor" (default: "offset")
	LogDetails           bool                   `yaml:"log_details"`       // Log pagination params on each request (default: false)
	ParamNames           PaginationParamNames   `yaml:"param_names"`       // Query parameter names
}

// PaginationParamNames configures the query parameter names used for pagination
type PaginationParamNames struct {
	Page      string `yaml:"page"`      // Page number param (default: "page")
	PerPage   string `yaml:"per_page"`  // Items per page param (default: "per_page")
	Cursor    string `yaml:"cursor"`    // Cursor param for cursor-based pagination (default: "cursor")
	Direction string `yaml:"direction"` // Direction param for cursor pagination: "next" or "prev" (default: "direction")
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
			SkipPaths:     []string{"/health"},
			DevMode:       false,
			DevBypassAuth: false,
			CacheEnabled:  true,
			CacheTTL:      15 * time.Minute,
		},
		Health: HealthConfig{
			Enabled:           true,  // ENABLED BY DEFAULT
			ShowDetailsInProd: false, // Details only in dev mode
		},
		Pagination: PaginationMiddlewareConfig{
			BaseMiddlewareConfig: BaseMiddlewareConfig{
				Enabled:          false, // DISABLED BY DEFAULT - consumers must opt-in
				DisableInDevMode: false,
			},
			DefaultPageSize: 20,
			MaxPageSize:     100,
			DefaultType:     "offset",
			LogDetails:      false,
			ParamNames: PaginationParamNames{
				Page:      "page",
				PerPage:   "per_page",
				Cursor:    "cursor",
				Direction: "direction",
			},
		},
	}
}
