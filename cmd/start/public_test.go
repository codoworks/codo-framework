package start

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
)

func TestPublicCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"start", "public", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Start the public API server")
}

func TestPublicCmd_Properties(t *testing.T) {
	assert.Equal(t, "public", publicCmd.Use)
	assert.Equal(t, "Start the public API server", publicCmd.Short)
	assert.Contains(t, publicCmd.Long, "no authentication")
}
