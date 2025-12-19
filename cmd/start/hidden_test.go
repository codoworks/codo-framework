package start

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
)

func TestHiddenCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"start", "hidden", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Start the hidden API server")
}

func TestHiddenCmd_Properties(t *testing.T) {
	assert.Equal(t, "hidden", hiddenCmd.Use)
	assert.Equal(t, "Start the hidden API server", hiddenCmd.Short)
	assert.Contains(t, hiddenCmd.Long, "admin/internal")
}
