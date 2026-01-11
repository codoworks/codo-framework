package db

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/app"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/core/db/seeds"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run database seeds",
	Long:  "Run all database seeds to populate the database with initial data",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		// Get bootstrap options from registered initializer
		initializer := app.GetInitializer()
		if initializer == nil {
			return fmt.Errorf("no application initializer registered")
		}

		opts, err := initializer(cfg)
		if err != nil {
			return fmt.Errorf("initialization failed: %w", err)
		}

		if opts.SeedAdder == nil {
			fmt.Fprintln(cmd.GetOutput(), "No seeds registered")
			return nil
		}

		// Initialize database client
		dbClient := db.NewClient(nil)
		clients.MustRegister(dbClient)

		dbConfig := &db.ClientConfig{
			Driver:          cfg.Database.Driver,
			DSN:             cfg.Database.DSN(),
			MaxOpenConns:    cfg.Database.MaxOpenConns,
			MaxIdleConns:    cfg.Database.MaxIdleConns,
			ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
		}

		if err := dbClient.Initialize(dbConfig); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer dbClient.Shutdown()

		// Create seeder and add seeds
		seeder := seeds.NewSeeder(dbClient.DB())
		opts.SeedAdder(seeder)

		if seeder.Count() == 0 {
			fmt.Fprintln(cmd.GetOutput(), "No seeds to run")
			return nil
		}

		fmt.Fprintf(cmd.GetOutput(), "Running %d seed(s)...\n", seeder.Count())

		if err := seeder.Run(c.Context()); err != nil {
			return fmt.Errorf("seed failed: %w", err)
		}

		fmt.Fprintln(cmd.GetOutput(), "Seeds complete")
		return nil
	},
}

func init() {
	cmd.AddDBCommand(seedCmd)
}
