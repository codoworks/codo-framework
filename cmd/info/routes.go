package info

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/pkg/handlers"
)

var scopeFilter string

var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Show registered routes",
	Long:  "Display all registered HTTP handlers and their routes",
	RunE: func(c *cobra.Command, args []string) error {
		// Register handlers for route display (nil db is safe - Prefix/Scope are static)
		handlers.RegisterAll(nil)

		out := cmd.GetOutput()
		httpHandlers := http.AllHandlers()

		if scopeFilter != "" {
			scope, err := http.ParseScope(scopeFilter)
			if err != nil {
				return fmt.Errorf("invalid scope: %s (must be public, protected, or hidden)", scopeFilter)
			}
			httpHandlers = http.GetHandlers(scope)
		}

		if len(httpHandlers) == 0 {
			fmt.Fprintln(out, "No handlers registered")
			return nil
		}

		fmt.Fprintf(out, "%-12s %-40s\n", "SCOPE", "PREFIX")
		fmt.Fprintln(out, strings.Repeat("-", 54))

		for _, h := range httpHandlers {
			fmt.Fprintf(out, "%-12s %-40s\n", h.Scope(), h.Prefix())
		}

		return nil
	},
}

// ResetRoutesFlags resets routes command flags (for testing)
func ResetRoutesFlags() {
	scopeFilter = ""
}

func init() {
	routesCmd.Flags().StringVar(&scopeFilter, "scope", "", "filter by scope (public, protected, hidden)")
	cmd.AddInfoCommand(routesCmd)
}
