package start

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
)

func TestProtectedCmd_Help(t *testing.T) {
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.RootCmd().SetArgs([]string{"start", "protected", "--help"})
	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Start the protected API server")
}

func TestProtectedCmd_Properties(t *testing.T) {
	assert.Equal(t, "protected", protectedCmd.Use)
	assert.Equal(t, "Start the protected API server", protectedCmd.Short)
	assert.Contains(t, protectedCmd.Long, "Kratos session")
}
