package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonsmithj/redrip/internal/file"
	"github.com/jasonsmithj/redrip/internal/logger"
	"github.com/jasonsmithj/redrip/internal/redash"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage redrip configuration",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current configuration settings",
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Info("Starting config list command", "profile", profile)

		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error("Failed to get home directory", "error", err)
			return fmt.Errorf("failed to get home directory: %v", err)
		}

		// Get config path
		configPath := filepath.Join(homeDir, ".redrip", "config.conf")
		logger.Debug("Using config path", "path", configPath)

		// Check if config file exists
		if !file.Exists(configPath) {
			logger.Warn("Config file does not exist", "path", configPath)
			fmt.Println("Config file does not exist. Creating default config file...")

			// Create default config file
			if err := redash.EnsureConfigFile(configPath); err != nil {
				logger.Error("Failed to create config file", "path", configPath, "error", err)
				return fmt.Errorf("failed to create config file: %v", err)
			}

			fmt.Printf("Default config file created at %s\n", configPath)
			fmt.Println("Please edit it to set your Redash URL and API Key")
			return nil
		}

		// Load config
		config, err := redash.LoadConfig(configPath)
		if err != nil {
			// If error is not about missing required fields, return error
			logger.Error("Failed to load config", "error", err)
			return fmt.Errorf("failed to load config: %v", err)
		}

		// Get active profile (from flag or env var)
		activeProfile := profile
		if activeProfile == "" {
			activeProfile = os.Getenv("REDRIP_PROFILE")
			if activeProfile == "" {
				activeProfile = "default"
			}
		}

		// Check if specified profile exists
		if _, exists := config.Profiles[activeProfile]; !exists {
			activeProfile = "default"
		}

		// Display config info
		fmt.Printf("Configuration file: %s\n\n", configPath)

		// If profile is specified, only show that profile
		if profile != "" {
			showProfileConfig(config, profile)
		} else {
			// Get environment profile
			envProfile := os.Getenv("REDRIP_PROFILE")

			// Show active profile first
			if activeProfile != "" {
				fmt.Printf("Active profile: %s", activeProfile)
				if envProfile != "" && envProfile == activeProfile {
					fmt.Printf(" (from REDRIP_PROFILE environment variable)")
				}
				fmt.Println()
				showProfileConfig(config, activeProfile)
				fmt.Println()
			}

			// Show all profiles
			fmt.Println("Available profiles:")
			fmt.Println("------------------")
			for profileName := range config.Profiles {
				// Skip active profile as we already showed it
				if profileName == activeProfile {
					continue
				}

				fmt.Printf("[%s]\n", profileName)
				showProfileConfig(config, profileName)
				fmt.Println()
			}
		}

		return nil
	},
}

// showProfileConfig displays the configuration for a specific profile
func showProfileConfig(config *redash.Config, profileName string) {
	if profileConfig, exists := config.Profiles[profileName]; exists {
		var redashURLStatus, apiKeyStatus string

		if profileConfig.RedashURL != "" {
			redashURLStatus = profileConfig.RedashURL
		} else {
			redashURLStatus = "[NOT SET]"
		}

		if profileConfig.APIKey != "" {
			apiKeyStatus = "[REDACTED]"
		} else {
			apiKeyStatus = "[NOT SET]"
		}

		// Display SQL directory with fallback to current directory
		sqlDir := profileConfig.SQLDir
		if sqlDir == "" {
			sqlDir = "[NOT SET - using current directory]"
		} else {
			// Check if directory exists
			if !file.Exists(sqlDir) {
				sqlDir = fmt.Sprintf("%s [DIRECTORY DOES NOT EXIST - will use current directory]", sqlDir)
			}
		}

		fmt.Printf("  redash_url = %s\n", redashURLStatus)
		fmt.Printf("  api_key = %s\n", apiKeyStatus)
		fmt.Printf("  sql_dir = %s\n", sqlDir)
	} else {
		fmt.Printf("Profile '%s' does not exist\n", profileName)
	}
}

// Helper function to check if the error is about missing required fields
func isMissingRequiredFields(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(fmt.Sprintf("%v", err), "required configuration values not found")
}

func init() {
	configCmd.AddCommand(configListCmd)
}
