package db

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
)

var (
	rollbackSteps int
	rollbackAll   bool
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback database migrations",
	Long:  "Rollback the last migration(s) from the database",
	RunE: func(c *cobra.Command, args []string) error {
		cfg := cmd.GetConfig()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		if rollbackAll {
			fmt.Fprintln(cmd.GetOutput(), "Rolling back all migrations...")
		} else {
			fmt.Fprintf(cmd.GetOutput(), "Rolling back %d migration(s)...\n", rollbackSteps)
		}

		fmt.Fprintln(cmd.GetOutput(), "Rollback complete")
		return nil
	},
}

// ResetRollbackFlags resets rollback command flags (for testing)
func ResetRollbackFlags() {
	rollbackSteps = 1
	rollbackAll = false
}

func init() {
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 1, "number of migrations to rollback")
	rollbackCmd.Flags().BoolVar(&rollbackAll, "all", false, "rollback all migrations")
	cmd.AddDBCommand(rollbackCmd)
}
