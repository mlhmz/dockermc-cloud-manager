package commands

import (
	"log/slog"
	"os"

	"github.com/mlhmz/dockermc-cloud-manager/internal/config"
	"github.com/spf13/cobra"
)

var (
	logger *slog.Logger
	cfg    *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dockermc-cloud-manager",
	Short: "Docker Minecraft Cloud Manager",
	Long: `A cloud management platform for running and orchestrating Minecraft servers using Docker containers.

This tool provides both a REST API server and CLI commands to manage Minecraft servers
running in isolated Docker containers with persistent storage.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger for all commands
		logger = config.SetupLogger()

		// Load configuration
		var err error
		cfg, err = config.Load()
		if err != nil {
			logger.Error("Failed to load configuration", "error", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags here if needed
	rootCmd.PersistentFlags().StringP("log-level", "l", "", "Log level (DEBUG, INFO, WARN, ERROR)")
	rootCmd.PersistentFlags().StringP("log-format", "f", "", "Log format (json, text)")
}
