package db

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
)

func TestSeedCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"db", "seed", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Run all database seeds")
}

func TestSeedCmd_Properties(t *testing.T) {
	assert.Equal(t, "seed", seedCmd.Use)
	assert.Equal(t, "Run database seeds", seedCmd.Short)
}

func TestSeedCmd_WithConfig(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	cfg := config.NewWithDefaults()
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the RunE function
	err := seedCmd.RunE(seedCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Running seeds")
	assert.Contains(t, output.String(), "Seeds complete")
}
