package start

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/app"
	"github.com/codoworks/codo-framework/core/http"
)

var hiddenCmd = &cobra.Command{
	Use:   "hidden",
	Short: "Start the hidden API server",
	Long:  "Start the hidden API server (admin/internal access only)",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		// Bootstrap in HTTPRouter mode (creates only hidden router with middleware)
		scope := http.ScopeHidden
		application, err := app.Bootstrap(cfg, app.BootstrapOptions{
			Mode:        app.HTTPRouter,
			RouterScope: &scope,
			// HandlerRegistrar: nil, // Handlers auto-registered via init()
			// MigrationAdder: nil,   // Consumers can provide via app.RegisterMigrations()
		})
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}

		// Auto-register CLI metadata (uses app's or framework defaults)
		meta := app.GetMetadata(application)
		cmd.SetAppInfo(meta.Name(), meta.Short(), meta.Long())

		// Type assert to SingleRouterApp and get router
		routerApp, ok := application.(app.SingleRouterApp)
		if !ok {
			return fmt.Errorf("expected SingleRouterApp, got %T", application)
		}
		router := routerApp.Router()

		// Setup graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			fmt.Fprintf(cmd.GetOutput(), "Starting Hidden API on http://localhost%s\n", cfg.Server.HiddenAddr())
			if err := router.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				stop()
			}
		}()

		<-ctx.Done()

		fmt.Fprintln(cmd.GetOutput(), "\nShutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownGrace.Duration())
		defer cancel()

		// Shutdown app (which handles router + clients)
		return application.Shutdown(shutdownCtx)
	},
}

func init() {
	cmd.AddStartCommand(hiddenCmd)
}
