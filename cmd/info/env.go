package info

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/app"
	"github.com/codoworks/codo-framework/core/config"
)

var showSecrets bool

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show environment and configuration",
	Long:  "Display comprehensive environment variables and configuration information",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		out := cmd.GetOutput()

		// Get bootstrap options from registered initializer
		var opts app.BootstrapOptions
		if initializer := app.GetInitializer(); initializer != nil {
			var err error
			opts, err = initializer(cfg)
			if err != nil {
				return fmt.Errorf("initialization failed: %w", err)
			}
		}

		// Set mode for this command
		opts.Mode = app.ConfigInspector

		// Bootstrap application (config + env vars only, no clients)
		application, err := app.Bootstrap(cfg, opts)
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}
		defer application.Shutdown(context.Background())

		// Print framework environment variables
		printFrameworkEnvVars(out, cfg)

		// Print all configuration sections
		printServiceConfig(out, cfg)
		printLoggerConfig(out, cfg)
		printServerConfig(out, cfg)
		printDatabaseConfig(out, cfg)
		printAuthConfig(out, cfg)
		printRabbitMQConfig(out, cfg)
		printMiddlewareConfig(out, cfg)
		printErrorsConfig(out, cfg)
	
		// Print consumer environment variables if available
		if configApp, ok := application.(app.ConfigApp); ok {
			if registry := configApp.EnvRegistry(); registry != nil {
				printConsumerEnvVars(out, registry)
			}
		}

		// Print extensions if any
		if len(cfg.Extensions) > 0 {
			printExtensions(out, cfg)
		}

		return nil
	},
}

// printFrameworkEnvVars prints the framework CODO_* environment variables table
func printFrameworkEnvVars(out io.Writer, cfg *config.Config) {
	fmt.Fprintf(out, "=== Framework Environment Variables (prefix: %s) ===\n", config.FrameworkEnvPrefix)
	fmt.Fprintf(out, "%-25s %-25s %-20s %s\n", "VARIABLE", "VALUE", "DEFAULT", "SOURCE")
	fmt.Fprintln(out, strings.Repeat("-", 85))

	for _, v := range config.FrameworkEnvVars() {
		// Get actual value from config using reflection (NOT from os.Getenv)
		actualValue := v.GetActualValue(cfg)
		displayValue := actualValue

		// Mask sensitive values unless --show-secrets
		if v.Sensitive && !showSecrets && actualValue != "" {
			displayValue = "***MASKED***"
		}

		// Format default display
		defaultDisplay := v.Default
		if defaultDisplay == "" {
			defaultDisplay = "(not set)"
		}

		// Detect source: env, yaml, or default
		source := v.DetectSource(cfg)

		fmt.Fprintf(out, "%-25s %-25s %-20s %s\n",
			v.Name,
			truncate(displayValue, 25),
			truncate(defaultDisplay, 20),
			source,
		)
	}
	fmt.Fprintln(out)
}

// printServiceConfig prints service configuration
func printServiceConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Service Configuration ===")
	fmt.Fprintf(out, "Environment:        %s\n", cfg.Service.Environment)
	fmt.Fprintf(out, "Dev Mode:           %v\n", cfg.DevMode)
	fmt.Fprintf(out, "Strict Response:    %v\n", cfg.Response.Strict)
	if len(cfg.Features.DisabledFeatures) == 0 {
		fmt.Fprintln(out, "Disabled Features:  (none)")
	} else {
		fmt.Fprintf(out, "Disabled Features:  %s\n", strings.Join(cfg.Features.DisabledFeatures, ", "))
	}
	fmt.Fprintln(out)
}

// printLoggerConfig prints logger configuration
func printLoggerConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Logger Configuration ===")
	fmt.Fprintf(out, "Level:   %s\n", cfg.Logger.Level)
	fmt.Fprintf(out, "Format:  %s\n", cfg.Logger.Format)
	fmt.Fprintln(out)
}

// printServerConfig prints server configuration
func printServerConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Server Configuration ===")
	fmt.Fprintf(out, "Public Port:      %d\n", cfg.Server.PublicPort)
	fmt.Fprintf(out, "Protected Port:   %d\n", cfg.Server.ProtectedPort)
	fmt.Fprintf(out, "Hidden Port:      %d\n", cfg.Server.HiddenPort)
	fmt.Fprintf(out, "Read Timeout:     %s\n", cfg.Server.ReadTimeout.Duration())
	fmt.Fprintf(out, "Write Timeout:    %s\n", cfg.Server.WriteTimeout.Duration())
	fmt.Fprintf(out, "Idle Timeout:     %s\n", cfg.Server.IdleTimeout.Duration())
	fmt.Fprintf(out, "Shutdown Grace:   %s\n", cfg.Server.ShutdownGrace.Duration())
	fmt.Fprintln(out)
}

// printDatabaseConfig prints database configuration
func printDatabaseConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Database Configuration ===")
	fmt.Fprintf(out, "Driver:           %s\n", cfg.Database.Driver)

	if cfg.Database.DSNString != "" {
		// Mask DSN if contains password
		dsn := cfg.Database.DSNString
		if !showSecrets {
			dsn = maskDSN(dsn)
		}
		fmt.Fprintf(out, "DSN:              %s\n", dsn)
	} else {
		fmt.Fprintf(out, "Host:             %s\n", cfg.Database.Host)
		fmt.Fprintf(out, "Port:             %d\n", cfg.Database.Port)
		fmt.Fprintf(out, "Name:             %s\n", cfg.Database.Name)
		fmt.Fprintf(out, "User:             %s\n", cfg.Database.User)
		if cfg.Database.Password != "" {
			if showSecrets {
				fmt.Fprintf(out, "Password:         %s\n", cfg.Database.Password)
			} else {
				fmt.Fprintf(out, "Password:         ***MASKED***\n")
			}
		} else {
			fmt.Fprintf(out, "Password:         (not set)\n")
		}
		fmt.Fprintf(out, "SSL Mode:         %s\n", cfg.Database.SSLMode)
	}
	fmt.Fprintf(out, "Max Open Conns:   %d\n", cfg.Database.MaxOpenConns)
	fmt.Fprintf(out, "Max Idle Conns:   %d\n", cfg.Database.MaxIdleConns)
	fmt.Fprintf(out, "Conn Max Life:    %s\n", time.Duration(cfg.Database.ConnMaxLifetime))
	fmt.Fprintln(out)
}

// printAuthConfig prints auth configuration
func printAuthConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Auth Configuration ===")
	fmt.Fprintf(out, "Kratos Public URL:  %s\n", cfg.Auth.KratosPublicURL)
	fmt.Fprintf(out, "Kratos Admin URL:   %s\n", cfg.Auth.KratosAdminURL)
	fmt.Fprintf(out, "Keto Read URL:      %s\n", cfg.Auth.KetoReadURL)
	fmt.Fprintf(out, "Keto Write URL:     %s\n", cfg.Auth.KetoWriteURL)
	fmt.Fprintf(out, "Session Cookie:     %s\n", cfg.Auth.SessionCookie)
	fmt.Fprintln(out)
}

// printRabbitMQConfig prints RabbitMQ configuration
func printRabbitMQConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== RabbitMQ Configuration ===")
	fmt.Fprintf(out, "Feature Enabled:  %v\n", cfg.Features.IsEnabled(config.FeatureRabbitMQ))

	if cfg.RabbitMQ.URL != "" {
		url := cfg.RabbitMQ.URL
		if !showSecrets {
			url = maskURL(url)
		}
		fmt.Fprintf(out, "URL:              %s\n", url)
	} else {
		fmt.Fprintf(out, "Host:             %s\n", cfg.RabbitMQ.Host)
		fmt.Fprintf(out, "Port:             %d\n", cfg.RabbitMQ.Port)
		fmt.Fprintf(out, "User:             %s\n", cfg.RabbitMQ.User)
		if cfg.RabbitMQ.Password != "" {
			if showSecrets {
				fmt.Fprintf(out, "Password:         %s\n", cfg.RabbitMQ.Password)
			} else {
				fmt.Fprintf(out, "Password:         ***MASKED***\n")
			}
		}
		fmt.Fprintf(out, "VHost:            %s\n", cfg.RabbitMQ.VHost)
	}
	fmt.Fprintf(out, "Exchange:         %s\n", cfg.RabbitMQ.Exchange)
	fmt.Fprintf(out, "Exchange Type:    %s\n", cfg.RabbitMQ.ExchangeType)
	fmt.Fprintf(out, "Prefetch Count:   %d\n", cfg.RabbitMQ.PrefetchCount)
	fmt.Fprintln(out)
}

// printMiddlewareConfig prints middleware configuration
func printMiddlewareConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Middleware Configuration ===")
	fmt.Fprintf(out, "Logger:           %s\n", enabledStr(cfg.Middleware.Logger.Enabled))
	fmt.Fprintf(out, "CORS:             %s\n", enabledStr(cfg.Middleware.CORS.Enabled))
	fmt.Fprintf(out, "Timeout:          %s (%s)\n", enabledStr(cfg.Middleware.Timeout.Enabled), cfg.Middleware.Timeout.Duration)
	fmt.Fprintf(out, "Recover:          %s\n", enabledStr(cfg.Middleware.Recover.Enabled))
	fmt.Fprintf(out, "Gzip:             %s (level %d)\n", enabledStr(cfg.Middleware.Gzip.Enabled), cfg.Middleware.Gzip.Level)
	fmt.Fprintf(out, "XSS:              %s\n", enabledStr(cfg.Middleware.XSS.Enabled))
	fmt.Fprintf(out, "Auth:             %s\n", enabledStr(cfg.Middleware.Auth.Enabled))
	fmt.Fprintf(out, "Health:           %s\n", enabledStr(cfg.Middleware.Health.Enabled))
	fmt.Fprintln(out)
}

// printErrorsConfig prints errors configuration
func printErrorsConfig(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Errors Configuration ===")
	fmt.Fprintf(out, "Expose Details:       %v\n", cfg.Errors.Handler.ExposeDetails)
	fmt.Fprintf(out, "Expose Stack Traces:  %v\n", cfg.Errors.Handler.ExposeStackTraces)
	fmt.Fprintf(out, "Stack Trace on 5xx:   %v\n", cfg.Errors.Capture.StackTraceOn5xx)
	fmt.Fprintf(out, "Stack Trace Depth:    %d\n", cfg.Errors.Capture.StackTraceDepth)
	fmt.Fprintln(out)
}

// printConsumerEnvVars prints consumer-registered environment variables
func printConsumerEnvVars(out io.Writer, registry *config.EnvVarRegistry) {
	if !registry.IsResolved() {
		return
	}

	resolved := registry.AllResolved()
	if len(resolved) == 0 {
		return
	}

	fmt.Fprintln(out, "=== Consumer Environment Variables ===")
	fmt.Fprintf(out, "%-12s %-25s %-10s %-20s %s\n", "GROUP", "VARIABLE", "TYPE", "VALUE", "REQUIRED")
	fmt.Fprintln(out, strings.Repeat("-", 85))

	// Sort by group then name
	type envEntry struct {
		group    string
		name     string
		value    *config.EnvVarValue
	}
	var entries []envEntry
	for name, val := range resolved {
		entries = append(entries, envEntry{
			group: val.Descriptor.Group,
			name:  name,
			value: val,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].group != entries[j].group {
			return entries[i].group < entries[j].group
		}
		return entries[i].name < entries[j].name
	})

	for _, e := range entries {
		v := e.value
		displayValue := v.MaskedValue()
		if showSecrets && v.Descriptor.Sensitive {
			displayValue = v.RawValue
		}

		required := "no"
		if v.Descriptor.Required {
			required = "yes"
		}

		group := v.Descriptor.Group
		if group == "" {
			group = "-"
		}

		fmt.Fprintf(out, "%-12s %-25s %-10s %-20s %s\n",
			group,
			v.Descriptor.Name,
			string(v.Descriptor.Type),
			truncate(displayValue, 20),
			required,
		)
	}
	fmt.Fprintln(out)
}

// printExtensions prints extension configurations
func printExtensions(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out, "=== Extensions ===")
	for key, value := range cfg.Extensions {
		masked := maskSensitiveFields(key, value)
		fmt.Fprintf(out, "%s: %v\n", key, masked)
	}
	fmt.Fprintln(out)
}

// Helper functions

func enabledStr(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func maskDSN(dsn string) string {
	// Simple DSN masking - find password patterns
	// Example: postgres://user:password@host:port/db
	// Try to mask the password portion
	if idx := strings.Index(dsn, "://"); idx != -1 {
		// Found protocol, look for @
		rest := dsn[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			// Found @ sign, check for :password before it
			userPass := rest[:atIdx]
			if colonIdx := strings.Index(userPass, ":"); colonIdx != -1 {
				// Has password
				user := userPass[:colonIdx]
				return dsn[:idx+3] + user + ":***@" + rest[atIdx+1:]
			}
		}
	}
	return dsn
}

func maskURL(url string) string {
	return maskDSN(url) // Same logic works for AMQP URLs
}

// maskSensitiveFields recursively masks sensitive values in config structures
func maskSensitiveFields(key string, value interface{}) interface{} {
	lowerKey := strings.ToLower(key)

	// Check if key contains sensitive terms
	sensitiveTerms := []string{"key", "secret", "password", "token", "credential"}
	for _, term := range sensitiveTerms {
		if strings.Contains(lowerKey, term) {
			// Mask the value if it's a string
			if str, ok := value.(string); ok && str != "" {
				if showSecrets {
					return str
				}
				return "***MASKED***"
			}
			if !showSecrets {
				return "***MASKED***"
			}
		}
	}

	// Recursively mask nested maps
	if m, ok := value.(map[string]interface{}); ok {
		masked := make(map[string]interface{})
		for k, v := range m {
			masked[k] = maskSensitiveFields(k, v)
		}
		return masked
	}

	// Return value as-is if not sensitive
	return value
}

func init() {
	envCmd.Flags().BoolVar(&showSecrets, "show-secrets", false, "Reveal masked values (passwords, keys, tokens)")
	cmd.AddInfoCommand(envCmd)
}
