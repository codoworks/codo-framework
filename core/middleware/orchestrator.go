package middleware

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/http"
)

// Orchestrator manages middleware lifecycle and application to routers
type Orchestrator struct {
	registry *Registry
	config   *config.Config
	active   []Middleware // Sorted by priority
}

// NewOrchestrator creates a new middleware orchestrator
func NewOrchestrator(cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		registry: GetGlobalRegistry(),
		config:   cfg,
		active:   []Middleware{},
	}
}

// Initialize configures all enabled middleware and sorts them by priority
func (o *Orchestrator) Initialize() error {
	allMiddlewares := o.registry.All()

	for _, m := range allMiddlewares {
		// Get config section for this middleware
		cfg := o.getConfigSection(m.ConfigKey())

		// Check base middleware config (dev mode aware)
		if !o.shouldEnableBasedOnMode(cfg) {
			continue // Skip based on dev mode rules
		}

		// Let middleware do additional custom checks
		if !m.Enabled(cfg) {
			continue // Skip disabled middleware
		}

		// Configure
		if err := m.Configure(cfg); err != nil {
			return fmt.Errorf("configure middleware %s: %w", m.Name(), err)
		}

		o.active = append(o.active, m)
	}

	// Sort by priority (stable sort preserves registration order for ties)
	sort.SliceStable(o.active, func(i, j int) bool {
		return o.active[i].Priority() < o.active[j].Priority()
	})

	return nil
}

// shouldEnableBasedOnMode checks BaseMiddlewareConfig fields using reflection
func (o *Orchestrator) shouldEnableBasedOnMode(cfg any) bool {
	if cfg == nil {
		return true // No config = enabled by default
	}

	// Get underlying value if pointer
	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return true // Not a struct, can't check fields
	}

	// Check Enabled field
	enabledField := val.FieldByName("Enabled")
	if enabledField.IsValid() && enabledField.Kind() == reflect.Bool {
		if !enabledField.Bool() {
			return false // Explicitly disabled
		}
	}

	// Check DisableInDevMode field
	disableInDevField := val.FieldByName("DisableInDevMode")
	if disableInDevField.IsValid() && disableInDevField.Kind() == reflect.Bool {
		if disableInDevField.Bool() && o.config.IsDevMode() {
			return false // Disabled in dev mode
		}
	}

	return true
}

// Apply applies middleware to the given router based on router type
func (o *Orchestrator) Apply(router *http.Router, routerType Router) {
	for _, m := range o.active {
		// Check if this middleware applies to this router type
		if m.Routers().Includes(routerType) {
			router.Use(m.Handler())
		}
	}
}

// List returns all active middleware for a given router type
func (o *Orchestrator) List(routerType Router) []Middleware {
	result := []Middleware{}
	for _, m := range o.active {
		if m.Routers().Includes(routerType) {
			result = append(result, m)
		}
	}
	return result
}

// ListAll returns all active middleware regardless of router type
func (o *Orchestrator) ListAll() []Middleware {
	return o.active
}

// getConfigSection retrieves the configuration section for a middleware
// Returns nil if no config key is specified or if the section doesn't exist
func (o *Orchestrator) getConfigSection(configKey string) any {
	if configKey == "" {
		return nil
	}

	// Use reflection to access the config field by the config key
	// Config key format: "middleware.logger" -> access Config.Middleware.Logger

	// For now, we'll try to access via Extensions map since Middleware field may not exist yet
	// Once we add the Middleware field to config.Config, this will be updated to use direct access

	// Try to get from Extensions first (during transition)
	if o.config.Extensions != nil {
		if middleware, ok := o.config.Extensions["middleware"]; ok {
			if middlewareMap, ok := middleware.(map[string]interface{}); ok {
				// Extract the specific middleware config
				// e.g., "middleware.logger" -> get middlewareMap["logger"]
				parts := parseConfigKey(configKey)
				if len(parts) >= 2 && parts[0] == "middleware" {
					if cfg, ok := middlewareMap[parts[1]]; ok {
						return cfg
					}
				}
			}
		}
	}

	// Try to access via reflection on Config struct
	// This will work once we add the Middleware field
	val := reflect.ValueOf(o.config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Parse config key: "middleware.logger" -> ["middleware", "logger"]
	parts := parseConfigKey(configKey)

	for _, part := range parts {
		// Find field by name (case-insensitive match for struct field)
		field := val.FieldByNameFunc(func(name string) bool {
			return equalFoldASCII(name, part)
		})

		if !field.IsValid() {
			return nil
		}

		// Move to next level
		val = field
	}

	// Return the config value if found
	// If it's a struct, return a pointer to it (middleware expect pointers)
	if val.IsValid() && val.CanInterface() {
		if val.Kind() == reflect.Struct {
			ptr := reflect.New(val.Type())
			ptr.Elem().Set(val)
			return ptr.Interface()
		}
		return val.Interface()
	}

	return nil
}

// parseConfigKey splits a config key by dots
// e.g., "middleware.logger" -> ["middleware", "logger"]
func parseConfigKey(key string) []string {
	if key == "" {
		return nil
	}

	var parts []string
	current := ""

	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(key[i])
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// equalFoldASCII compares two strings case-insensitively (ASCII only)
func equalFoldASCII(s, t string) bool {
	if len(s) != len(t) {
		return false
	}

	for i := 0; i < len(s); i++ {
		c1 := s[i]
		c2 := t[i]

		// Convert to lowercase if uppercase ASCII letter
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}

		if c1 != c2 {
			return false
		}
	}

	return true
}
