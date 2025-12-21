package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmd(t *testing.T) {
	ResetFlags()
	ResetVersionFlags()
	defer func() {
		ResetFlags()
		ResetVersionFlags()
	}()

	SetVersion("1.0.0-test")

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	rootCmd.SetArgs([]string{"version"})
	err := Execute()
	require.NoError(t, err)

	assert.Contains(t, output.String(), "codo 1.0.0-test")
	assert.Contains(t, output.String(), "Go version:")
	assert.Contains(t, output.String(), "OS/Arch:")
}

func TestVersionCmd_JSON(t *testing.T) {
	ResetFlags()
	ResetVersionFlags()
	defer func() {
		ResetFlags()
		ResetVersionFlags()
	}()

	SetVersion("2.0.0-json")

	output := new(bytes.Buffer)
	SetOutput(output)
	defer ResetOutput()

	rootCmd.SetArgs([]string{"version", "--json"})
	err := Execute()
	require.NoError(t, err)

	var info VersionInfo
	err = json.Unmarshal(output.Bytes(), &info)
	require.NoError(t, err)

	assert.Equal(t, "codo", info.AppName)
	assert.Equal(t, "2.0.0-json", info.Version)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
}

func TestVersionInfo(t *testing.T) {
	info := VersionInfo{
		AppName:   "test-app",
		Version:   "1.0.0",
		GoVersion: "go1.21.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	assert.Equal(t, "test-app", info.AppName)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "go1.21.0", info.GoVersion)
	assert.Equal(t, "linux", info.OS)
	assert.Equal(t, "amd64", info.Arch)
}

func TestResetVersionFlags(t *testing.T) {
	jsonOutput = true
	ResetVersionFlags()
	assert.False(t, jsonOutput)
}
