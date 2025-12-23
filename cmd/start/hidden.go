package start

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/app"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/errors"
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
		scope := http.ScopeHidden
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

		// Get logger for startup messages
		log := clients.MustGetTyped[*logger.Logger](logger.ClientName)

		// Setup graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			log.Infof("Starting Hidden API on http://localhost%s", cfg.Server.HiddenAddr())
			if err := router.Start(); err != nil {
				frameworkErr := errors.Unavailable("Server failed to start").
					WithCause(err).
					WithPhase(errors.PhaseBootstrap)
				errors.RenderCLI(frameworkErr)
				stop()
			}
		}()

		<-ctx.Done()

		log.Info("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownGrace.Duration())
		defer cancel()

		// Shutdown app (which handles router + clients)
		return application.Shutdown(shutdownCtx)
	},
}

func init() {
	cmd.AddStartCommand(hiddenCmd)
}
