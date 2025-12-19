package db

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
)

func TestMigrateCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	ResetMigrateFlags()
	defer func() {
		cmd.ResetFlags()
		ResetMigrateFlags()
	}()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"db", "migrate", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	// Check for the long description
	assert.Contains(t, output.String(), "pending database migrations")
	assert.Contains(t, output.String(), "--fresh")
}

func TestMigrateCmd_FreshFlag(t *testing.T) {
	cmd.ResetFlags()
	ResetMigrateFlags()
	defer func() {
		cmd.ResetFlags()
		ResetMigrateFlags()
	}()

	// Verify --fresh flag is available
	flag := migrateCmd.Flags().Lookup("fresh")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestMigrateCmd_Properties(t *testing.T) {
	assert.Equal(t, "migrate", migrateCmd.Use)
	assert.Equal(t, "Run database migrations", migrateCmd.Short)
}

func TestMigrateCmd_WithConfig(t *testing.T) {
	cmd.ResetFlags()
	ResetMigrateFlags()
	defer func() {
		cmd.ResetFlags()
		ResetMigrateFlags()
	}()

	// Set up a test config
	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the RunE function
	err := migrateCmd.RunE(migrateCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Running migrations")
	assert.Contains(t, output.String(), "Migrations complete")
}

func TestMigrateCmd_FreshWithConfig(t *testing.T) {
	cmd.ResetFlags()
	ResetMigrateFlags()
	defer func() {
		cmd.ResetFlags()
		ResetMigrateFlags()
	}()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Set the fresh flag directly
	freshMigrate = true

	// Directly run the RunE function
	err := migrateCmd.RunE(migrateCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Rolling back all migrations")
	assert.Contains(t, output.String(), "Running migrations")
}

func TestResetMigrateFlags(t *testing.T) {
	freshMigrate = true
	ResetMigrateFlags()
	assert.False(t, freshMigrate)
}
