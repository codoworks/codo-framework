package clients

import (
	"github.com/codoworks/codo-framework/core/config"
)

// EnvConfigurable is an optional interface that clients can implement
// to declare their environment variable requirements.
//
// When a client implements this interface, the framework will:
// 1. Automatically register these env vars during bootstrap
// 2. Validate them before client initialization
// 3. Make the resolved values available via the EnvVarRegistry
//
// Example implementation:
//
//	func (c *StripeClient) EnvVars() []config.EnvVarDescriptor {
//	    return []config.EnvVarDescriptor{
//	        {
//	            Name:      "STRIPE_API_KEY",
//	            Type:      config.EnvTypeString,
//	            Required:  true,
//	            Sensitive: true,
//	            Group:     "stripe",
//	        },
//	    }
//	}
type EnvConfigurable interface {
	// EnvVars returns the environment variable descriptors needed by this client
	EnvVars() []config.EnvVarDescriptor
}

// EnvConfigProvider is an optional interface that clients can implement
// to receive their configuration from resolved environment variables.
//
// When a client implements this interface, the framework will:
// 1. After resolving env vars, call ConfigFromEnv with the client's group values
// 2. Use the returned config for client initialization instead of the default config
//
// Example implementation:
//
//	func (c *StripeClient) ConfigFromEnv(values map[string]*config.EnvVarValue) (any, error) {
//	    return &StripeConfig{
//	        APIKey: values["STRIPE_API_KEY"].String(),
//	    }, nil
//	}
type EnvConfigProvider interface {
	// ConfigFromEnv creates client configuration from resolved environment variables
	// The values map is keyed by variable name and contains only this client's group values
	ConfigFromEnv(values map[string]*config.EnvVarValue) (any, error)
}

// CollectEnvVarsFromClients collects environment variable descriptors from all
// registered clients that implement EnvConfigurable
func CollectEnvVarsFromClients() []config.EnvVarDescriptor {
	var descs []config.EnvVarDescriptor

	for _, client := range All() {
		if envClient, ok := client.(EnvConfigurable); ok {
			descs = append(descs, envClient.EnvVars()...)
		}
	}

	return descs
}

// BuildClientConfigFromEnv builds client configuration from resolved env vars
// for clients that implement EnvConfigProvider
//
// Returns a map of client name to config, for clients that successfully built config
func BuildClientConfigFromEnv(registry *config.EnvVarRegistry) (map[string]any, error) {
	configs := make(map[string]any)

	for name, client := range All() {
		if envProvider, ok := client.(EnvConfigProvider); ok {
			// Get the group name - use client name as default group
			groupName := name
			if envConfigurable, hasEnvVars := client.(EnvConfigurable); hasEnvVars {
				envVars := envConfigurable.EnvVars()
				if len(envVars) > 0 && envVars[0].Group != "" {
					groupName = envVars[0].Group
				}
			}

			// Get resolved values for this client's group
			groupValues := registry.GetGroup(groupName)

			// Build config from env values
			cfg, err := envProvider.ConfigFromEnv(groupValues)
			if err != nil {
				return nil, err
			}

			configs[name] = cfg
		}
	}

	return configs, nil
}
