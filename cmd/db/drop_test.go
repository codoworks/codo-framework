package db

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
)

func TestDropCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	ResetDropFlags()
	defer func() {
		cmd.ResetFlags()
		ResetDropFlags()
	}()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"db", "drop", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Drop the database")
	assert.Contains(t, output.String(), "--force")
}

func TestDropCmd_ForceFlag(t *testing.T) {
	cmd.ResetFlags()
	ResetDropFlags()
	defer func() {
		cmd.ResetFlags()
		ResetDropFlags()
	}()

	flag := dropCmd.Flags().Lookup("force")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestDropCmd_Properties(t *testing.T) {
	assert.Equal(t, "drop", dropCmd.Use)
	assert.Equal(t, "Drop the database", dropCmd.Short)
}

func TestDropCmd_WithoutForce(t *testing.T) {
	ResetDropFlags()
	defer ResetDropFlags()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the dropCmd RunE function
	err := dropCmd.RunE(dropCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "WARNING")
	assert.Contains(t, output.String(), "--force")
}

func TestResetDropFlags(t *testing.T) {
	forceDropDB = true
	ResetDropFlags()
	assert.False(t, forceDropDB)
}
