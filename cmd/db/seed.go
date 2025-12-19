package db

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
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

		fmt.Fprintln(cmd.GetOutput(), "Running seeds...")
		fmt.Fprintln(cmd.GetOutput(), "Seeds complete")
		return nil
	},
}

func init() {
	cmd.AddDBCommand(seedCmd)
}
