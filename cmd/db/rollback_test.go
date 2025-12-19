package db

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
)

func TestRollbackCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	ResetRollbackFlags()
	defer func() {
		cmd.ResetFlags()
		ResetRollbackFlags()
	}()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"db", "rollback", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Rollback the last migration(s)")
	assert.Contains(t, output.String(), "--steps")
	assert.Contains(t, output.String(), "--all")
}

func TestRollbackCmd_StepsFlag(t *testing.T) {
	cmd.ResetFlags()
	ResetRollbackFlags()
	defer func() {
		cmd.ResetFlags()
		ResetRollbackFlags()
	}()

	flag := rollbackCmd.Flags().Lookup("steps")
	require.NotNil(t, flag)
	assert.Equal(t, "1", flag.DefValue)
}

func TestRollbackCmd_AllFlag(t *testing.T) {
	cmd.ResetFlags()
	ResetRollbackFlags()
	defer func() {
		cmd.ResetFlags()
		ResetRollbackFlags()
	}()

	flag := rollbackCmd.Flags().Lookup("all")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestRollbackCmd_Properties(t *testing.T) {
	assert.Equal(t, "rollback", rollbackCmd.Use)
	assert.Equal(t, "Rollback database migrations", rollbackCmd.Short)
}

func TestRollbackCmd_WithConfig(t *testing.T) {
	cmd.ResetFlags()
	ResetRollbackFlags()
	defer func() {
		cmd.ResetFlags()
		ResetRollbackFlags()
	}()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the RunE function
	err := rollbackCmd.RunE(rollbackCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Rolling back 1 migration(s)")
	assert.Contains(t, output.String(), "Rollback complete")
}

func TestRollbackCmd_AllWithConfig(t *testing.T) {
	cmd.ResetFlags()
	ResetRollbackFlags()
	defer func() {
		cmd.ResetFlags()
		ResetRollbackFlags()
	}()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Set the --all flag directly
	rollbackAll = true

	// Directly run the RunE function
	err := rollbackCmd.RunE(rollbackCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Rolling back all migrations")
	assert.Contains(t, output.String(), "Rollback complete")
}

func TestRollbackCmd_StepsWithConfig(t *testing.T) {
	cmd.ResetFlags()
	ResetRollbackFlags()
	defer func() {
		cmd.ResetFlags()
		ResetRollbackFlags()
	}()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Set the --steps flag directly
	rollbackSteps = 3

	// Directly run the RunE function
	err := rollbackCmd.RunE(rollbackCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Rolling back 3 migration(s)")
}

func TestResetRollbackFlags(t *testing.T) {
	rollbackSteps = 5
	rollbackAll = true
	ResetRollbackFlags()
	assert.Equal(t, 1, rollbackSteps)
	assert.False(t, rollbackAll)
}
