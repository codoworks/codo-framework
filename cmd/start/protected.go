package start

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/http"
)

var protectedCmd = &cobra.Command{
	Use:   "protected",
	Short: "Start the protected API server",
	Long:  "Start the protected API server (Kratos session authentication required)",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		router := http.NewRouter(http.ScopeProtected, cfg.Server.ProtectedAddr())

		if err := router.RegisterHandlers(); err != nil {
			return fmt.Errorf("failed to register handlers: %w", err)
		}

		// Setup graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			fmt.Fprintf(cmd.GetOutput(), "Starting Protected API on http://localhost%s\n", cfg.Server.ProtectedAddr())
			if err := router.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				stop()
			}
		}()

		<-ctx.Done()

		fmt.Fprintln(cmd.GetOutput(), "\nShutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownGrace.Duration())
		defer cancel()

		return router.Shutdown(shutdownCtx)
	},
}

func init() {
	cmd.AddStartCommand(protectedCmd)
}
