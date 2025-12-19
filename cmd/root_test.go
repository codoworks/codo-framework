package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/core/config"
)

func TestExecute(t *testing.T) {
	// Reset state
	ResetFlags()
	defer ResetFlags()

	// Capture output
	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	// Execute help
	rootCmd.SetArgs([]string{"--help"})
	err := Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Codo Framework")
}

func TestExecute_Help(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	rootCmd.SetArgs([]string{"help"})
	err := Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Usage:")
	assert.Contains(t, output.String(), "Available Commands:")
}

func TestRootCmd_DevFlag(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	// Set args with dev flag
	rootCmd.SetArgs([]string{"--dev", "version"})

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	err := Execute()
	require.NoError(t, err)

	// After executing, dev mode should be set on the config
	// Note: version command skips config loading, so we test by checking the flag was parsed
	assert.True(t, devMode)
}

func TestRootCmd_VerboseFlag(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	rootCmd.SetArgs([]string{"--verbose", "version"})

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	err := Execute()
	require.NoError(t, err)

	assert.True(t, IsVerbose())
}

func TestRootCmd_InvalidFlag(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	rootCmd.SetArgs([]string{"--invalid-flag"})
	err := Execute()
	assert.Error(t, err)
}

func TestGetConfig(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	// Initially nil
	assert.Nil(t, GetConfig())

	// Set a config
	testCfg := config.NewWithDefaults()
	SetConfig(testCfg)

	assert.NotNil(t, GetConfig())
	assert.Equal(t, testCfg, GetConfig())
}

func TestSetVersion(t *testing.T) {
	oldVersion := version
	defer func() { version = oldVersion }()

	SetVersion("1.2.3")
	assert.Equal(t, "1.2.3", GetVersion())
}

func TestGetOutput(t *testing.T) {
	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	assert.Equal(t, output, GetOutput())
}

func TestResetOutput(t *testing.T) {
	output := new(bytes.Buffer)
	SetOutput(output)
	ResetOutput()

	// Output should be reset to stdout
	assert.NotEqual(t, output, GetOutput())
}

func TestRootCmd(t *testing.T) {
	cmd := RootCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "codo", cmd.Use)
}

func TestAddStartCommand(t *testing.T) {
	// Test that the startCmd exists and is a subcommand of root
	assert.Equal(t, "start", startCmd.Use)
	assert.Equal(t, "Start individual API servers", startCmd.Short)

	// Test the AddStartCommand function by adding a test command
	testCmd := &cobra.Command{Use: "test-start-sub", Short: "Test subcommand"}
	AddStartCommand(testCmd)
	// Verify it was added
	found := false
	for _, c := range startCmd.Commands() {
		if c.Use == "test-start-sub" {
			found = true
			break
		}
	}
	assert.True(t, found, "test-start-sub should be added to start command")
}

func TestAddDBCommand(t *testing.T) {
	// Test that the dbCmd exists and is a subcommand of root
	assert.Equal(t, "db", dbCmd.Use)
	assert.Equal(t, "Database commands", dbCmd.Short)

	// Test the AddDBCommand function by adding a test command
	testCmd := &cobra.Command{Use: "test-db-sub", Short: "Test subcommand"}
	AddDBCommand(testCmd)
	// Verify it was added
	found := false
	for _, c := range dbCmd.Commands() {
		if c.Use == "test-db-sub" {
			found = true
			break
		}
	}
	assert.True(t, found, "test-db-sub should be added to db command")
}

func TestAddTaskCommand(t *testing.T) {
	// Test that the taskCmd exists and is a subcommand of root
	assert.Equal(t, "task", taskCmd.Use)
	assert.Equal(t, "Task commands", taskCmd.Short)

	// Test the AddTaskCommand function by adding a test command
	testCmd := &cobra.Command{Use: "test-task-sub", Short: "Test subcommand"}
	AddTaskCommand(testCmd)
	// Verify it was added
	found := false
	for _, c := range taskCmd.Commands() {
		if c.Use == "test-task-sub" {
			found = true
			break
		}
	}
	assert.True(t, found, "test-task-sub should be added to task command")
}

func TestAddInfoCommand(t *testing.T) {
	// Test that the infoCmd exists and is a subcommand of root
	assert.Equal(t, "info", infoCmd.Use)
	assert.Equal(t, "Information commands", infoCmd.Short)

	// Test the AddInfoCommand function by adding a test command
	testCmd := &cobra.Command{Use: "test-info-sub", Short: "Test subcommand"}
	AddInfoCommand(testCmd)
	// Verify it was added
	found := false
	for _, c := range infoCmd.Commands() {
		if c.Use == "test-info-sub" {
			found = true
			break
		}
	}
	assert.True(t, found, "test-info-sub should be added to info command")
}

func TestRootCmd_ConfigFlagInvalid(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	// Test that LoadFromFile fails for a non-existent file
	// This is what would happen if --config points to a non-existent file
	_, err := config.LoadFromFile("/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config")
}

func TestResetFlags(t *testing.T) {
	// Set some values
	cfgFile = "test.yaml"
	devMode = true
	verbose = true

	ResetFlags()

	assert.Equal(t, "", cfgFile)
	assert.False(t, devMode)
	assert.False(t, verbose)
	assert.Nil(t, cfg)
}
