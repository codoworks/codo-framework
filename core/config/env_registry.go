package config

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// EnvVarRegistry holds all declared environment variables and their resolved values
type EnvVarRegistry struct {
	mu          sync.RWMutex
	descriptors map[string]*EnvVarDescriptor   // keyed by Name
	groups      map[string][]*EnvVarDescriptor // keyed by Group
	values      map[string]*EnvVarValue        // resolved values, keyed by Name
	errors      EnvValidationErrors            // accumulated validation errors
	resolved    bool                           // whether Resolve() has been called
}

// NewEnvVarRegistry creates a new environment variable registry
func NewEnvVarRegistry() *EnvVarRegistry {
	return &EnvVarRegistry{
		descriptors: make(map[string]*EnvVarDescriptor),
		groups:      make(map[string][]*EnvVarDescriptor),
		values:      make(map[string]*EnvVarValue),
		errors:      nil,
	}
}

// Register registers a single environment variable descriptor
func (r *EnvVarRegistry) Register(desc EnvVarDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if desc.Name == "" {
		return fmt.Errorf("environment variable name cannot be empty")
	}

	if _, exists := r.descriptors[desc.Name]; exists {
		return fmt.Errorf("environment variable %q already registered", desc.Name)
	}

	// Store descriptor
	descCopy := desc
	r.descriptors[desc.Name] = &descCopy

	// Add to group index
	if desc.Group != "" {
		r.groups[desc.Group] = append(r.groups[desc.Group], &descCopy)
	}

	return nil
}

// RegisterMany registers multiple environment variable descriptors
func (r *EnvVarRegistry) RegisterMany(descs []EnvVarDescriptor) error {
	for _, desc := range descs {
		if err := r.Register(desc); err != nil {
			return err
		}
	}
	return nil
}

// MustRegister registers a descriptor and panics on error
func (r *EnvVarRegistry) MustRegister(desc EnvVarDescriptor) {
	if err := r.Register(desc); err != nil {
		panic(fmt.Sprintf("failed to register env var %q: %v", desc.Name, err))
	}
}

// Resolve reads all registered environment variables, converts types, and validates
// This should be called once during bootstrap after all descriptors are registered
func (r *EnvVarRegistry) Resolve() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.resolved {
		return fmt.Errorf("environment variables already resolved")
	}

	// Reset errors
	r.errors = nil

	// Create Viper instance for env var reading
	v := viper.New()
	v.AutomaticEnv() // No prefix - consumers control full env var names

	for name, desc := range r.descriptors {
		rawVal := v.GetString(name)
		isSet := rawVal != ""

		// Handle required check
		if !isSet && desc.Required {
			r.errors.Add(name, "required variable not set")
			continue
		}

		// Apply default for optional unset variables
		if !isSet && desc.Default != nil {
			value := &EnvVarValue{
				Descriptor: desc,
				RawValue:   "",
				Value:      desc.Default,
				IsSet:      false,
			}
			r.values[name] = value
			continue
		}

		// Skip if not set and no default (optional with no default)
		if !isSet {
			continue
		}

		// Type conversion
		convertedVal, err := r.convertValue(rawVal, desc.Type)
		if err != nil {
			r.errors.Add(name, fmt.Sprintf("type conversion failed: %s", err.Error()))
			continue
		}

		// Custom validation
		if desc.Validator != nil {
			if err := desc.Validator(convertedVal); err != nil {
				r.errors.Add(name, fmt.Sprintf("validation failed: %s", err.Error()))
				continue
			}
		}

		// Store resolved value
		r.values[name] = &EnvVarValue{
			Descriptor: desc,
			RawValue:   rawVal,
			Value:      convertedVal,
			IsSet:      true,
		}
	}

	r.resolved = true
	return r.errors.ToError()
}

// convertValue converts a raw string value to the specified type
func (r *EnvVarRegistry) convertValue(rawVal string, varType EnvVarType) (any, error) {
	switch varType {
	case EnvTypeString:
		return rawVal, nil

	case EnvTypeInt:
		return strconv.Atoi(rawVal)

	case EnvTypeBool:
		return strconv.ParseBool(rawVal)

	case EnvTypeDuration:
		return time.ParseDuration(rawVal)

	case EnvTypeFloat:
		return strconv.ParseFloat(rawVal, 64)

	case EnvTypeURL:
		u, err := url.Parse(rawVal)
		if err != nil {
			return nil, err
		}
		// Validate URL has scheme and host
		if u.Scheme == "" {
			return nil, fmt.Errorf("URL missing scheme")
		}
		if u.Host == "" {
			return nil, fmt.Errorf("URL missing host")
		}
		return u, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", varType)
	}
}

// Get retrieves a resolved environment variable value by name
func (r *EnvVarRegistry) Get(name string) (*EnvVarValue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.resolved {
		return nil, fmt.Errorf("environment variables not yet resolved; call Resolve() first")
	}

	val, ok := r.values[name]
	if !ok {
		// Check if it was registered but not set
		if _, registered := r.descriptors[name]; registered {
			return nil, fmt.Errorf("environment variable %q not set", name)
		}
		return nil, fmt.Errorf("environment variable %q not registered", name)
	}
	return val, nil
}

// GetString retrieves a string environment variable value
func (r *EnvVarRegistry) GetString(name string) (string, error) {
	val, err := r.Get(name)
	if err != nil {
		return "", err
	}
	if s, ok := val.Value.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("environment variable %q is not a string (type: %s)", name, val.Descriptor.Type)
}

// GetInt retrieves an integer environment variable value
func (r *EnvVarRegistry) GetInt(name string) (int, error) {
	val, err := r.Get(name)
	if err != nil {
		return 0, err
	}
	if i, ok := val.Value.(int); ok {
		return i, nil
	}
	return 0, fmt.Errorf("environment variable %q is not an int (type: %s)", name, val.Descriptor.Type)
}

// GetBool retrieves a boolean environment variable value
func (r *EnvVarRegistry) GetBool(name string) (bool, error) {
	val, err := r.Get(name)
	if err != nil {
		return false, err
	}
	if b, ok := val.Value.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("environment variable %q is not a bool (type: %s)", name, val.Descriptor.Type)
}

// GetDuration retrieves a duration environment variable value
func (r *EnvVarRegistry) GetDuration(name string) (time.Duration, error) {
	val, err := r.Get(name)
	if err != nil {
		return 0, err
	}
	if d, ok := val.Value.(time.Duration); ok {
		return d, nil
	}
	return 0, fmt.Errorf("environment variable %q is not a duration (type: %s)", name, val.Descriptor.Type)
}

// GetFloat retrieves a float64 environment variable value
func (r *EnvVarRegistry) GetFloat(name string) (float64, error) {
	val, err := r.Get(name)
	if err != nil {
		return 0, err
	}
	if f, ok := val.Value.(float64); ok {
		return f, nil
	}
	return 0, fmt.Errorf("environment variable %q is not a float (type: %s)", name, val.Descriptor.Type)
}

// GetURL retrieves a URL environment variable value
func (r *EnvVarRegistry) GetURL(name string) (*url.URL, error) {
	val, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	if u, ok := val.Value.(*url.URL); ok {
		return u, nil
	}
	return nil, fmt.Errorf("environment variable %q is not a URL (type: %s)", name, val.Descriptor.Type)
}

// GetGroup retrieves all resolved values for a specific group
func (r *EnvVarRegistry) GetGroup(group string) map[string]*EnvVarValue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*EnvVarValue)
	descs, ok := r.groups[group]
	if !ok {
		return result
	}

	for _, desc := range descs {
		if val, exists := r.values[desc.Name]; exists {
			result[desc.Name] = val
		}
	}
	return result
}

// ValidationErrors returns any validation errors that occurred during Resolve()
func (r *EnvVarRegistry) ValidationErrors() EnvValidationErrors {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.errors
}

// HasErrors returns true if there are validation errors
func (r *EnvVarRegistry) HasErrors() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.errors) > 0
}

// IsResolved returns true if Resolve() has been called
func (r *EnvVarRegistry) IsResolved() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resolved
}

// All returns all registered descriptors
func (r *EnvVarRegistry) All() map[string]*EnvVarDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*EnvVarDescriptor, len(r.descriptors))
	for k, v := range r.descriptors {
		result[k] = v
	}
	return result
}

// AllResolved returns all resolved values
func (r *EnvVarRegistry) AllResolved() map[string]*EnvVarValue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*EnvVarValue, len(r.values))
	for k, v := range r.values {
		result[k] = v
	}
	return result
}

// MaskedValues returns all resolved values with sensitive values masked
// Useful for logging and debugging
func (r *EnvVarRegistry) MaskedValues() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]string, len(r.values))
	for name, val := range r.values {
		result[name] = val.MaskedValue()
	}
	return result
}

// Groups returns all group names
func (r *EnvVarRegistry) Groups() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	groups := make([]string, 0, len(r.groups))
	for g := range r.groups {
		groups = append(groups, g)
	}
	return groups
}

// Reset clears all registered descriptors and values
// Primarily useful for testing
func (r *EnvVarRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.descriptors = make(map[string]*EnvVarDescriptor)
	r.groups = make(map[string][]*EnvVarDescriptor)
	r.values = make(map[string]*EnvVarValue)
	r.errors = nil
	r.resolved = false
}
