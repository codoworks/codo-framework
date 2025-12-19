package task

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/tasks"
)

func TestListCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"task", "list", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "task")
}

func TestListCmd_Properties(t *testing.T) {
	assert.Equal(t, "list", listCmd.Use)
	assert.Equal(t, "List available tasks", listCmd.Short)
}

func TestListCmd_NoTasks(t *testing.T) {
	tasks.Clear()
	defer tasks.Clear()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	listCmd.Run(listCmd, []string{})

	assert.Contains(t, output.String(), "No tasks registered")
}

func TestListCmd_WithTasks(t *testing.T) {
	tasks.Clear()
	defer tasks.Clear()

	// Register some test tasks
	tasks.Register(tasks.Task{
		Name:        "alpha-task",
		Description: "First task",
		Run:         func(ctx context.Context, args []string) error { return nil },
	})
	tasks.Register(tasks.Task{
		Name:        "beta-task",
		Description: "Second task",
		Run:         func(ctx context.Context, args []string) error { return nil },
	})

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	listCmd.Run(listCmd, []string{})

	out := output.String()
	assert.Contains(t, out, "Available tasks:")
	assert.Contains(t, out, "alpha-task")
	assert.Contains(t, out, "beta-task")
	assert.Contains(t, out, "First task")
	assert.Contains(t, out, "Second task")
}

func TestListCmd_TaskWithNoDescription(t *testing.T) {
	tasks.Clear()
	defer tasks.Clear()

	tasks.Register(tasks.Task{
		Name: "no-desc-task",
		Run:  func(ctx context.Context, args []string) error { return nil },
	})

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	listCmd.Run(listCmd, []string{})

	assert.Contains(t, output.String(), "no-desc-task")
	assert.Contains(t, output.String(), "(no description)")
}
