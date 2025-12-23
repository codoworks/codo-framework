// Package cmd provides the CLI commands for the Codo Framework.
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/errors"
)

var (
	cfgFile string
	devMode bool
	verbose bool
	cfg     *config.Config
	version string
	appName  string = "codo"
	appShort string = "Codo Framework CLI"
	appLong  string = "Codo Framework - A production-ready Go backend framework"
	output  io.Writer = os.Stdout
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: appShort,
	Long:  appLong,
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

// handleCommandError handles command errors with pretty CLI rendering
func handleCommandError(err error) {
	if err == nil {
		return
	}

	// Use CLI renderer for beautiful colored output
	errors.RenderCLI(err)

	os.Exit(1)
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		handleCommandError(err)
	}
	return nil
}

// SetVersion sets the version string
func SetVersion(v string) {
	version = v
}

// GetVersion returns the version string
func GetVersion() string {
	return version
}

// SetAppName sets the application name (used in CLI usage text)
func SetAppName(name string) {
	appName = name
	rootCmd.Use = name
}

// SetAppShort sets the short description shown in help
func SetAppShort(short string) {
	appShort = short
	rootCmd.Short = short
}

// SetAppLong sets the long description shown in help
func SetAppLong(long string) {
	appLong = long
	rootCmd.Long = long
}

// SetAppInfo is a convenience function to set all app metadata at once
func SetAppInfo(name, short, long string) {
	SetAppName(name)
	SetAppShort(short)
	SetAppLong(long)
}

// GetAppName returns the application name
func GetAppName() string {
	return appName
}

// GetAppShort returns the short description
func GetAppShort() string {
	return appShort
}

// GetAppLong returns the long description
func GetAppLong() string {
	return appLong
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
	appName = "codo"
	appShort = "Codo Framework CLI"
	appLong = "Codo Framework - A production-ready Go backend framework"
	rootCmd.Use = appName
	rootCmd.Short = appShort
	rootCmd.Long = appLong
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
