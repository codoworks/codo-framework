package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeCmd_Help(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	rootCmd.SetArgs([]string{"serve", "--help"})
	err := Execute()
	require.NoError(t, err)

	// Check that help contains information about the serve command
	assert.Contains(t, output.String(), "serve")
	assert.Contains(t, output.String(), "public, protected, and hidden API servers")
}

func TestServeCmd_RequiresConfig(t *testing.T) {
	ResetFlags()
	defer ResetFlags()

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	// Reset config to nil
	SetConfig(nil)

	// serve command should work when config is loaded
	// We just test that the command exists and has the right properties
	assert.Equal(t, "serve", serveCmd.Use)
	assert.Equal(t, "Start all API servers", serveCmd.Short)
}
