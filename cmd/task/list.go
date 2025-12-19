package task

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/tasks"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tasks",
	Long:  "List all registered tasks with their descriptions",
	Run: func(c *cobra.Command, args []string) {
		allTasks := tasks.All()

		if len(allTasks) == 0 {
			fmt.Fprintln(cmd.GetOutput(), "No tasks registered")
			return
		}

		// Sort tasks by name
		sort.Slice(allTasks, func(i, j int) bool {
			return allTasks[i].Name < allTasks[j].Name
		})

		fmt.Fprintln(cmd.GetOutput(), "Available tasks:")
		fmt.Fprintln(cmd.GetOutput(), strings.Repeat("-", 60))
		fmt.Fprintf(cmd.GetOutput(), "%-20s %s\n", "NAME", "DESCRIPTION")
		fmt.Fprintln(cmd.GetOutput(), strings.Repeat("-", 60))

		for _, t := range allTasks {
			desc := t.Description
			if desc == "" {
				desc = "(no description)"
			}
			fmt.Fprintf(cmd.GetOutput(), "%-20s %s\n", t.Name, desc)
		}
	},
}

func init() {
	cmd.AddTaskCommand(listCmd)
}
