package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/cmd"
)

func TestVersion(t *testing.T) {
	// Test that the Version variable is set to the default value
	assert.Equal(t, "0.0.0-dev", Version)
}

func TestVersionCanBeSet(t *testing.T) {
	// Test that Version can be modified (as done by ldflags at build time)
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "1.2.3-test"
	assert.Equal(t, "1.2.3-test", Version)
}

func TestMainSetsVersion(t *testing.T) {
	// Verify that cmd.SetVersion is called with the Version variable
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "test-version-123"
	cmd.SetVersion(Version)

	assert.Equal(t, "test-version-123", cmd.GetVersion())
}

func TestMainExecuteSuccess(t *testing.T) {
	// Test that Execute succeeds for help command
	// This tests the flow without os.Exit

	// Reset flags before test
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Set args to help (which doesn't require config)
	cmd.RootCmd().SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "Codo Framework")
}

func TestMainExecuteVersionCommand(t *testing.T) {
	// Test that version command works
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	cmd.SetVersion("1.0.0-test")
	cmd.RootCmd().SetArgs([]string{"version"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "1.0.0-test")
}

func TestMainExecuteError(t *testing.T) {
	// Test that Execute returns error for invalid flag
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	cmd.RootCmd().SetArgs([]string{"--invalid-flag-xyz"})

	err := cmd.Execute()
	assert.Error(t, err)
}

// TestMainFunctionErrorExit tests that main exits with code 1 on error.
// This uses subprocess testing because main() calls os.Exit().
func TestMainFunctionErrorExit(t *testing.T) {
	if os.Getenv("TEST_MAIN_EXIT") == "1" {
		// Override os.Args to simulate invalid flag
		os.Args = []string{"codo", "--nonexistent-flag"}
		main()
		return
	}

	// Run this test as a subprocess
	execCmd := exec.Command(os.Args[0], "-test.run=TestMainFunctionErrorExit")
	execCmd.Env = append(os.Environ(), "TEST_MAIN_EXIT=1")
	err := execCmd.Run()

	// Should fail with exit code 1
	if e, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, e.ExitCode())
	} else {
		t.Errorf("expected ExitError, got %T: %v", err, err)
	}
}

// TestMainFunctionSuccess tests that main succeeds with valid args.
// This uses subprocess testing because main() calls os.Exit().
func TestMainFunctionSuccess(t *testing.T) {
	if os.Getenv("TEST_MAIN_SUCCESS") == "1" {
		// Override os.Args to use version command
		os.Args = []string{"codo", "version"}
		main()
		return
	}

	// Run this test as a subprocess
	execCmd := exec.Command(os.Args[0], "-test.run=TestMainFunctionSuccess")
	execCmd.Env = append(os.Environ(), "TEST_MAIN_SUCCESS=1")
	output, err := execCmd.CombinedOutput()

	// Should succeed (exit code 0)
	assert.NoError(t, err, "main should exit with code 0 for version command: %s", string(output))
}

// TestMainFunctionHelp tests that main succeeds with --help.
func TestMainFunctionHelp(t *testing.T) {
	if os.Getenv("TEST_MAIN_HELP") == "1" {
		os.Args = []string{"codo", "--help"}
		main()
		return
	}

	execCmd := exec.Command(os.Args[0], "-test.run=TestMainFunctionHelp")
	execCmd.Env = append(os.Environ(), "TEST_MAIN_HELP=1")
	output, err := execCmd.CombinedOutput()

	assert.NoError(t, err, "main should exit with code 0 for --help: %s", string(output))
	assert.Contains(t, string(output), "Codo Framework")
}

func TestCmdSetVersionIntegration(t *testing.T) {
	// Integration test: verify SetVersion is properly exposed
	originalVersion := cmd.GetVersion()
	defer cmd.SetVersion(originalVersion)

	cmd.SetVersion("integration-test-1.0.0")
	assert.Equal(t, "integration-test-1.0.0", cmd.GetVersion())
}

func TestCmdExecuteIntegration(t *testing.T) {
	// Integration test: verify Execute is properly exposed and works
	cmd.ResetFlags()
	defer cmd.ResetFlags()

	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	defer cmd.ResetOutput()

	// Use version command since it doesn't need config
	cmd.RootCmd().SetArgs([]string{"version"})
	err := cmd.Execute()
	assert.NoError(t, err)
}
