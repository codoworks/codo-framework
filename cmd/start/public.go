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

var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Start the public API server",
	Long:  "Start the public API server (no authentication required)",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		// Get bootstrap options from registered initializer
		var opts app.BootstrapOptions
		if initializer := app.GetInitializer(); initializer != nil {
			var err error
			opts, err = initializer(cfg)
			if err != nil {
				return fmt.Errorf("initialization failed: %w", err)
			}
		}

		// Set mode and scope for this command
		scope := http.ScopePublic
		opts.Mode = app.HTTPRouter
		opts.RouterScope = &scope

		// Bootstrap application
		application, err := app.Bootstrap(cfg, opts)
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
			fmt.Fprintf(cmd.GetOutput(), "Starting Public API on http://localhost%s\n", cfg.Server.PublicAddr())
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
	cmd.AddStartCommand(publicCmd)
}
