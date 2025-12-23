package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/core/app"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/errors"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start all API servers",
	Long:  "Start the public, protected, and hidden API servers concurrently",
	RunE: func(cmd *cobra.Command, args []string) error {
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
		opts.Mode = app.HTTPServer

		// Bootstrap application
		application, err := app.Bootstrap(cfg, opts)
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}

		// Auto-register CLI metadata (uses app's or framework defaults)
		meta := app.GetMetadata(application)
		SetAppInfo(meta.Name(), meta.Short(), meta.Long())

		// Type assert to HTTPApp and get server
		httpApp, ok := application.(app.HTTPApp)
		if !ok {
			return fmt.Errorf("expected HTTPApp, got %T", application)
		}
		server := httpApp.Server()

		// Get logger for startup messages
		log := clients.MustGetTyped[*logger.Logger](logger.ClientName)

		// Setup graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			log.Info("Starting servers...")
			log.Infof("Public API: http://localhost%s", cfg.Server.PublicAddr())
			log.Infof("Protected API: http://localhost%s", cfg.Server.ProtectedAddr())
			log.Infof("Hidden API: http://localhost%s", cfg.Server.HiddenAddr())

			if err := server.Start(); err != nil {
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

		// Shutdown app (which handles server + clients)
		return application.Shutdown(shutdownCtx)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
