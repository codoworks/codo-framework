package config

import (
	"os"

	"github.com/joho/godotenv"
)

// DotEnvResult contains the result of .env file loading
type DotEnvResult struct {
	// Loaded is true if file was found and loaded successfully
	Loaded bool
	// FilePath is the path to the file that was attempted
	FilePath string
	// Error is non-nil if file exists but was malformed
	Error error
}

// LoadDotEnv attempts to load environment variables from a .env file.
//
// Behavior:
//   - Only loads .env from the current working directory
//   - Does NOT overwrite existing environment variables (OS env > .env)
//   - Missing file is OK (returns Loaded=false, Error=nil)
//   - Malformed file returns an error
func LoadDotEnv() DotEnvResult {
	const envFile = ".env"

	// Check if file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return DotEnvResult{
			Loaded:   false,
			FilePath: envFile,
			Error:    nil, // Missing file is OK
		}
	}

	// File exists, try to load it
	// godotenv.Load() does NOT overwrite existing env vars by default
	if err := godotenv.Load(envFile); err != nil {
		return DotEnvResult{
			Loaded:   false,
			FilePath: envFile,
			Error:    err, // Malformed file
		}
	}

	return DotEnvResult{
		Loaded:   true,
		FilePath: envFile,
		Error:    nil,
	}
}
