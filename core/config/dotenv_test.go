package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDotEnv_MissingFileIsOK(t *testing.T) {
	// Change to temp dir without .env
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldDir)

	result := LoadDotEnv()

	assert.False(t, result.Loaded)
	assert.Nil(t, result.Error)
	assert.Equal(t, ".env", result.FilePath)
}

func TestLoadDotEnv_LoadsValidFile(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldDir)

	// Create .env file
	envContent := "TEST_DOTENV_VAR=hello\nTEST_DOTENV_VAR2=world"
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644))

	// Clear any existing values
	os.Unsetenv("TEST_DOTENV_VAR")
	os.Unsetenv("TEST_DOTENV_VAR2")
	defer func() {
		os.Unsetenv("TEST_DOTENV_VAR")
		os.Unsetenv("TEST_DOTENV_VAR2")
	}()

	result := LoadDotEnv()

	assert.True(t, result.Loaded)
	assert.Nil(t, result.Error)
	assert.Equal(t, "hello", os.Getenv("TEST_DOTENV_VAR"))
	assert.Equal(t, "world", os.Getenv("TEST_DOTENV_VAR2"))
}

func TestLoadDotEnv_DoesNotOverwriteExistingVars(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldDir)

	// Create .env file with a value
	envContent := "TEST_EXISTING_VAR=from_dotenv"
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644))

	// Set OS env var BEFORE loading
	os.Setenv("TEST_EXISTING_VAR", "from_shell")
	defer os.Unsetenv("TEST_EXISTING_VAR")

	result := LoadDotEnv()

	assert.True(t, result.Loaded)
	// OS value should be preserved (not overwritten)
	assert.Equal(t, "from_shell", os.Getenv("TEST_EXISTING_VAR"))
}

func TestLoadDotEnv_HandlesComments(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldDir)

	// Create .env file with comments
	envContent := `# This is a comment
TEST_COMMENT_VAR=value
# Another comment
TEST_COMMENT_VAR2=value2`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644))

	os.Unsetenv("TEST_COMMENT_VAR")
	os.Unsetenv("TEST_COMMENT_VAR2")
	defer func() {
		os.Unsetenv("TEST_COMMENT_VAR")
		os.Unsetenv("TEST_COMMENT_VAR2")
	}()

	result := LoadDotEnv()

	assert.True(t, result.Loaded)
	assert.Nil(t, result.Error)
	assert.Equal(t, "value", os.Getenv("TEST_COMMENT_VAR"))
	assert.Equal(t, "value2", os.Getenv("TEST_COMMENT_VAR2"))
}

func TestLoadDotEnv_HandlesQuotedValues(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldDir)

	// Create .env file with quoted values
	envContent := `TEST_QUOTED_DOUBLE="hello world"
TEST_QUOTED_SINGLE='hello world'`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644))

	os.Unsetenv("TEST_QUOTED_DOUBLE")
	os.Unsetenv("TEST_QUOTED_SINGLE")
	defer func() {
		os.Unsetenv("TEST_QUOTED_DOUBLE")
		os.Unsetenv("TEST_QUOTED_SINGLE")
	}()

	result := LoadDotEnv()

	assert.True(t, result.Loaded)
	assert.Nil(t, result.Error)
	assert.Equal(t, "hello world", os.Getenv("TEST_QUOTED_DOUBLE"))
	assert.Equal(t, "hello world", os.Getenv("TEST_QUOTED_SINGLE"))
}

func TestLoadDotEnv_HandlesEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(oldDir)

	// Create .env file with empty lines
	envContent := `TEST_EMPTY_LINE_VAR1=value1

TEST_EMPTY_LINE_VAR2=value2

`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644))

	os.Unsetenv("TEST_EMPTY_LINE_VAR1")
	os.Unsetenv("TEST_EMPTY_LINE_VAR2")
	defer func() {
		os.Unsetenv("TEST_EMPTY_LINE_VAR1")
		os.Unsetenv("TEST_EMPTY_LINE_VAR2")
	}()

	result := LoadDotEnv()

	assert.True(t, result.Loaded)
	assert.Nil(t, result.Error)
	assert.Equal(t, "value1", os.Getenv("TEST_EMPTY_LINE_VAR1"))
	assert.Equal(t, "value2", os.Getenv("TEST_EMPTY_LINE_VAR2"))
}
