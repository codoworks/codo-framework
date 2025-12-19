// Package cmd provides the CLI commands for the Codo Framework.
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/core/config"
)

var (
	cfgFile string
	devMode bool
	verbose bool
	cfg     *config.Config
	version string
	output  io.Writer = os.Stdout
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "codo",
	Short: "Codo Framework CLI",
	Long:  "Codo Framework - A production-ready Go backend framework",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for certain commands
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}

		// Skip if config is already set (e.g., for testing)
		if cfg != nil {
			// Apply dev mode from flag even if config is pre-set
			if devMode {
				cfg.SetDevMode(true)
			}
			return nil
		}

		var err error
		if cfgFile != "" {
			cfg, err = config.LoadFromFile(cfgFile)
		} else {
			cfg, err = config.Load()
		}

		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Apply dev mode from flag
		if devMode {
			cfg.SetDevMode(true)
		}

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Command group parents
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start individual API servers",
	Long:  "Start individual API servers (public, protected, or hidden)",
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database commands",
	Long:  "Database management commands (migrate, seed, rollback, create, drop)",
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task commands",
	Long:  "Task execution commands (exec, list)",
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information commands",
	Long:  "Display information about the application (env, routes)",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the version string
func SetVersion(v string) {
	version = v
}

// GetVersion returns the version string
func GetVersion() string {
	return version
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}

// SetConfig sets the configuration (for testing)
func SetConfig(c *config.Config) {
	cfg = c
}

// IsVerbose returns true if verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// GetOutput returns the output writer
func GetOutput() io.Writer {
	return output
}

// SetOutput sets the output writer (for testing)
func SetOutput(w io.Writer) {
	output = w
	rootCmd.SetOut(w)
	rootCmd.SetErr(w)
}

// ResetOutput resets output to stdout
func ResetOutput() {
	output = os.Stdout
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}

// AddStartCommand adds a command to the start group
func AddStartCommand(cmd *cobra.Command) {
	startCmd.AddCommand(cmd)
}

// AddDBCommand adds a command to the db group
func AddDBCommand(cmd *cobra.Command) {
	dbCmd.AddCommand(cmd)
}

// AddTaskCommand adds a command to the task group
func AddTaskCommand(cmd *cobra.Command) {
	taskCmd.AddCommand(cmd)
}

// AddInfoCommand adds a command to the info group
func AddInfoCommand(cmd *cobra.Command) {
	infoCmd.AddCommand(cmd)
}

// RootCmd returns the root command (for testing)
func RootCmd() *cobra.Command {
	return rootCmd
}

// ResetFlags resets all flags to their default values (for testing)
func ResetFlags() {
	cfgFile = ""
	devMode = false
	verbose = false
	cfg = nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&devMode, "dev", "d", false, "enable development mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(infoCmd)
}
