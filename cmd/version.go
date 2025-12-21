package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var jsonOutput bool

// VersionInfo holds version information
type VersionInfo struct {
	AppName   string `json:"app_name"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print version information including Go version, OS, and architecture",
	Run: func(cmd *cobra.Command, args []string) {
		info := VersionInfo{
			AppName:   appName,
			Version:   version,
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(info, "", "  ")
			fmt.Fprintln(GetOutput(), string(data))
		} else {
			fmt.Fprintf(GetOutput(), "%s %s\n", info.AppName, info.Version)
			fmt.Fprintf(GetOutput(), "Go version: %s\n", info.GoVersion)
			fmt.Fprintf(GetOutput(), "OS/Arch: %s/%s\n", info.OS, info.Arch)
		}
	},
}

// ResetVersionFlags resets version command flags (for testing)
func ResetVersionFlags() {
	jsonOutput = false
}

func init() {
	versionCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	rootCmd.AddCommand(versionCmd)
}
