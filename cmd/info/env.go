package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/codoworks/codo-framework/cmd"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show environment information",
	Long:  "Display environment and configuration information",
	Run: func(c *cobra.Command, args []string) {
		cfg := cmd.GetConfig()
		out := cmd.GetOutput()

		fmt.Fprintln(out, "=== Service ===")
		fmt.Fprintf(out, "Name:        %s\n", cmd.GetAppName())
		fmt.Fprintf(out, "Version:     %s\n", cmd.GetVersion())
		fmt.Fprintf(out, "Environment: %s\n", cfg.Service.Environment)
		fmt.Fprintf(out, "Dev Mode:    %v\n", cfg.IsDevMode())

		fmt.Fprintln(out, "\n=== Server ===")
		fmt.Fprintf(out, "Public Port:    %d\n", cfg.Server.PublicPort)
		fmt.Fprintf(out, "Protected Port: %d\n", cfg.Server.ProtectedPort)
		fmt.Fprintf(out, "Hidden Port:    %d\n", cfg.Server.HiddenPort)

		fmt.Fprintln(out, "\n=== Database ===")
		fmt.Fprintf(out, "Driver: %s\n", cfg.Database.Driver)
		// Show DSN if it's explicitly set, otherwise show individual fields
		if cfg.Database.DSNString != "" {
			fmt.Fprintf(out, "DSN:    %s\n", cfg.Database.DSNString)
		} else {
			fmt.Fprintf(out, "Host:   %s\n", cfg.Database.Host)
			fmt.Fprintf(out, "Port:   %d\n", cfg.Database.Port)
			fmt.Fprintf(out, "Name:   %s\n", cfg.Database.Name)
			fmt.Fprintf(out, "User:   %s\n", cfg.Database.User)
			// Password is masked
			if cfg.Database.Password != "" {
				fmt.Fprintf(out, "Password: %s\n", MaskPassword(cfg.Database.Password))
			}
		}

		fmt.Fprintln(out, "\n=== Environment Variables ===")
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "CODO_") {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 && IsSecret(parts[0]) {
					fmt.Fprintf(out, "%s=%s\n", parts[0], MaskPassword(parts[1]))
				} else {
					fmt.Fprintln(out, env)
				}
			}
		}

		// Display app-specific config sections from Extensions
		if len(cfg.Extensions) > 0 {
			fmt.Fprintln(out, "\n=== Application Config ===")
			for key, value := range cfg.Extensions {
				// Mask sensitive values before displaying
				masked := maskSensitiveFields(key, value)

				// Pretty-print as YAML
				yamlBytes, err := yaml.Marshal(map[string]interface{}{key: masked})
				if err == nil {
					// Print YAML output (already formatted)
					fmt.Fprint(out, string(yamlBytes))
				} else {
					// Fallback to simple display
					fmt.Fprintf(out, "%s: %v\n", key, masked)
				}
			}
		}
	},
}

// MaskPassword masks a password string, showing only first and last char
func MaskPassword(s string) string {
	if len(s) <= 2 {
		return "***"
	}
	return s[:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:]
}

// IsSecret checks if an environment variable name contains secret keywords
func IsSecret(name string) bool {
	secrets := []string{"PASSWORD", "SECRET", "KEY", "TOKEN"}
	upper := strings.ToUpper(name)
	for _, s := range secrets {
		if strings.Contains(upper, s) {
			return true
		}
	}
	return false
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
				if len(str) > 4 {
					return "***" + str[len(str)-4:]
				}
				return "***"
			}
			return "***"
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
	cmd.AddInfoCommand(envCmd)
}
