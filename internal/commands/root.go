package commands

import (
	"log/slog"

	"github.com/jasonsmithj/redrip/internal/logger"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	debug   bool
	quiet   bool
	profile string
)

var rootCmd = &cobra.Command{
	Use:   "redrip",
	Short: "CLI tool for Redash",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		// Set log level based on flags
		logLevel := slog.LevelWarn // Default to warn level (suppresses INFO)

		// Only one flag should be active
		if quiet {
			// In quiet mode, only show errors
			logLevel = slog.LevelError
		} else if verbose {
			// In verbose mode, show info, warnings and errors
			logLevel = slog.LevelInfo
		} else if debug {
			// In debug mode, show all logs
			logLevel = slog.LevelDebug
		}

		logger.Initialize(logLevel)
		logger.Debug("redrip CLI starting")
	},
}

// Execute starts the application and processes command line arguments
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Add flags for controlling log level
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable warning and error logs")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable all logs (debug, info, warning, error)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all logs except errors")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "Use specific configuration profile (default: uses REDRIP_PROFILE env var or 'default' profile)")

	// Make flags mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "debug", "quiet")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(dumpCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(diffCmd)
}
