package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
)

var forceDropDB bool

var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drop the database",
	Long:  "Drop the database specified in the configuration (DESTRUCTIVE)",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		if !forceDropDB {
			fmt.Fprintf(cmd.GetOutput(), "WARNING: This will permanently delete the database '%s'.\n", cfg.Database.Name)
			fmt.Fprintln(cmd.GetOutput(), "Use --force to confirm.")
			return nil
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
			fmt.Fprintf(cmd.GetOutput(), "Dropping SQLite database: %s\n", cfg.Database.Name)
			if err := os.Remove(cfg.Database.Name); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to drop database: %w", err)
			}
			fmt.Fprintln(cmd.GetOutput(), "Database dropped successfully")
			return nil
		default:
			return fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
		}

		db, err := sql.Open(cfg.Database.Driver, dsn)
		if err != nil {
			return fmt.Errorf("failed to connect to database server: %w", err)
		}
		defer db.Close()

		var dropQuery string
		switch cfg.Database.Driver {
		case "postgres":
			dropQuery = fmt.Sprintf("DROP DATABASE IF EXISTS %s", cfg.Database.Name)
		case "mysql":
			dropQuery = fmt.Sprintf("DROP DATABASE IF EXISTS %s", cfg.Database.Name)
		}

		fmt.Fprintf(cmd.GetOutput(), "Dropping database %s...\n", cfg.Database.Name)
		if _, err := db.Exec(dropQuery); err != nil {
			return fmt.Errorf("failed to drop database: %w", err)
		}

		fmt.Fprintln(cmd.GetOutput(), "Database dropped successfully")
		return nil
	},
}

// ResetDropFlags resets drop command flags (for testing)
func ResetDropFlags() {
	forceDropDB = false
}

func init() {
	dropCmd.Flags().BoolVar(&forceDropDB, "force", false, "force drop without confirmation")
	cmd.AddDBCommand(dropCmd)
}
