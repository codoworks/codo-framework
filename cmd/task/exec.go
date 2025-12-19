package task

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/tasks"
)

var execCmd = &cobra.Command{
	Use:   "exec [name] [args...]",
	Short: "Execute a task",
	Long:  "Execute a registered task by name with optional arguments",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		taskName := args[0]
		taskArgs := args[1:]

		fmt.Fprintf(cmd.GetOutput(), "Executing task: %s\n", taskName)

		ctx := context.Background()
		if err := tasks.Execute(ctx, taskName, taskArgs); err != nil {
			return fmt.Errorf("task failed: %w", err)
		}

		fmt.Fprintln(cmd.GetOutput(), "Task completed successfully")
		return nil
	},
}

func init() {
	cmd.AddTaskCommand(execCmd)
}
