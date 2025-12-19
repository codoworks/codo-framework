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

var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Start the public API server",
	Long:  "Start the public API server (no authentication required)",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		router := http.NewRouter(http.ScopePublic, cfg.Server.PublicAddr())

		if err := router.RegisterHandlers(); err != nil {
			return fmt.Errorf("failed to register handlers: %w", err)
		}

		// Setup graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			fmt.Fprintf(cmd.GetOutput(), "Starting Public API on http://localhost%s\n", cfg.Server.PublicAddr())
			if err := router.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				stop()
			}
		}()

		<-ctx.Done()

		fmt.Fprintln(cmd.GetOutput(), "\nShutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownGrace)
		defer cancel()

		return router.Shutdown(shutdownCtx)
	},
}

func init() {
	cmd.AddStartCommand(publicCmd)
}
