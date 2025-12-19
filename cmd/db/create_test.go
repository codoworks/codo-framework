package db

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
	"github.com/codoworks/codo-framework/core/config"
)

func TestCreateCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"db", "create", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Create the database")
}

func TestCreateCmd_Properties(t *testing.T) {
	assert.Equal(t, "create", createCmd.Use)
	assert.Equal(t, "Create the database", createCmd.Short)
}

func TestCreateCmd_SQLite(t *testing.T) {
	cfg := config.NewWithDefaults()
	cfg.Database.Driver = "sqlite"
	cfg.Database.Name = "test.db"
	cmd.SetConfig(cfg)
	defer cmd.SetConfig(nil)

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Directly run the createCmd RunE function
	err := createCmd.RunE(createCmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, output.String(), "SQLite database will be created automatically")
}
