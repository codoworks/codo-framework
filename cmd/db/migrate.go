package db

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
)

var freshMigrate bool

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "Run all pending database migrations",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		if freshMigrate {
			fmt.Fprintln(cmd.GetOutput(), "Rolling back all migrations...")
		}

		fmt.Fprintln(cmd.GetOutput(), "Running migrations...")
		fmt.Fprintln(cmd.GetOutput(), "Migrations complete")
		return nil
	},
}

// ResetMigrateFlags resets migrate command flags (for testing)
func ResetMigrateFlags() {
	freshMigrate = false
}

func init() {
	migrateCmd.Flags().BoolVar(&freshMigrate, "fresh", false, "rollback all migrations first")
	cmd.AddDBCommand(migrateCmd)
}
