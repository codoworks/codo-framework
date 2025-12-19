// Package main is the entry point for the Codo Framework CLI.
package main

import (
	"os"

	"github.com/codoworks/codo-framework/cmd"

	// Register CLI subcommands via their init() functions
	_ "github.com/codoworks/codo-framework/cmd/db"
	_ "github.com/codoworks/codo-framework/cmd/info"
	_ "github.com/codoworks/codo-framework/cmd/start"
	_ "github.com/codoworks/codo-framework/cmd/task"
)

// Version is set at build time via -ldflags.
// Example: go build -ldflags "-X main.Version=1.0.0"
var Version = "0.0.0-dev"

func main() {
	cmd.SetVersion(Version)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
