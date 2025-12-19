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

func TestExecCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"task", "exec", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Execute")
}

func TestExecCmd_Properties(t *testing.T) {
	assert.Equal(t, "exec [name] [args...]", execCmd.Use)
	assert.Equal(t, "Execute a task", execCmd.Short)
}

func TestExecCmd_RequiresArgs(t *testing.T) {
	// The exec command requires at least 1 arg (task name)
	// Test that the command has Args set to MinimumNArgs(1)
	assert.NotNil(t, execCmd.Args)
}

func TestExecCmd_TaskNotFound(t *testing.T) {
	tasks.Clear()
	defer tasks.Clear()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the RunE function with a nonexistent task
	err := execCmd.RunE(execCmd, []string{"nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestExecCmd_TaskSuccess(t *testing.T) {
	tasks.Clear()
	defer tasks.Clear()

	// Register a test task
	executed := false
	tasks.Register(tasks.Task{
		Name:        "test-task",
		Description: "A test task",
		Run: func(ctx context.Context, args []string) error {
			executed = true
			return nil
		},
	})

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the RunE function
	err := execCmd.RunE(execCmd, []string{"test-task"})
	require.NoError(t, err)

	assert.True(t, executed)
	assert.Contains(t, output.String(), "Executing task: test-task")
	assert.Contains(t, output.String(), "Task completed successfully")
}
