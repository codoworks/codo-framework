package db

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create the database",
	Long:  "Create the database specified in the configuration",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		// Build connection string without database name
		var dsn string
		switch cfg.Database.Driver {
		case "postgres":
			sslMode := cfg.Database.SSLMode
			if sslMode == "" {
				sslMode = "disable"
			}
			dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s",
				cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, sslMode)
		case "mysql":
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/",
				cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port)
		case "sqlite":
			// SQLite doesn't need explicit creation
			fmt.Fprintf(cmd.GetOutput(), "SQLite database will be created automatically: %s\n", cfg.Database.Name)
			return nil
		default:
			return fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
		}

		db, err := sql.Open(cfg.Database.Driver, dsn)
		if err != nil {
			return fmt.Errorf("failed to connect to database server: %w", err)
		}
		defer db.Close()

		var createQuery string
		switch cfg.Database.Driver {
		case "postgres":
			createQuery = fmt.Sprintf("CREATE DATABASE %s", cfg.Database.Name)
		case "mysql":
			createQuery = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", cfg.Database.Name)
		}

		fmt.Fprintf(cmd.GetOutput(), "Creating database %s...\n", cfg.Database.Name)
		if _, err := db.Exec(createQuery); err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

		fmt.Fprintln(cmd.GetOutput(), "Database created successfully")
		return nil
	},
}

func init() {
	cmd.AddDBCommand(createCmd)
}
