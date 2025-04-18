// Package commands implements the command-line interface for the redrip application.
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jasonsmithj/redrip/internal/logger"
	"github.com/jasonsmithj/redrip/internal/redash"

	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all queries as .sql files",
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Info("Starting dump command")

		client, err := redash.NewClient()
		if err != nil {
			logger.Error("Failed to initialize Redash client", "error", err)
			return fmt.Errorf("failed to initialize Redash client: %v", err)
		}

		logger.Debug("Fetching queries from Redash")
		queries, err := client.ListQueries()
		if err != nil {
			logger.Error("Failed to list queries", "error", err)
			return err
		}
		logger.Info("Retrieved queries from Redash", "count", len(queries))

		// Get configured SQL directory
		sqlDir, err := redash.GetSQLDir()
		if err != nil {
			logger.Error("Failed to get SQL directory", "error", err)
			return fmt.Errorf("failed to get SQL directory: %v", err)
		}
		logger.Debug("Using SQL directory", "dir", sqlDir)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(sqlDir, 0755); err != nil {
			logger.Error("Failed to create directory", "dir", sqlDir, "error", err)
			return fmt.Errorf("failed to create directory %s: %v", sqlDir, err)
		}

		// Generate timestamp for the JSON file
		timestamp := time.Now().Format("20060102150405") // YYYYmmDDHHMMSS format
		jsonFilename := fmt.Sprintf("%s.json", timestamp)
		jsonFilePath := filepath.Join(sqlDir, jsonFilename)

		// Save the full query list as JSON
		logger.Debug("Creating JSON file with all queries", "file", jsonFilePath)
		jsonOutput, err := json.MarshalIndent(queries, "", "  ")
		if err != nil {
			logger.Error("Failed to marshal queries to JSON", "error", err)
			return fmt.Errorf("failed to marshal queries to JSON: %v", err)
		}

		if err := os.WriteFile(jsonFilePath, jsonOutput, 0644); err != nil {
			logger.Error("Failed to write JSON file", "file", jsonFilePath, "error", err)
			return fmt.Errorf("failed to write JSON file %s: %v", jsonFilePath, err)
		}
		logger.Info("Queries saved to JSON file", "file", jsonFilePath)

		// Dump individual SQL files
		logger.Info("Dumping queries to SQL files", "count", len(queries), "dir", sqlDir)
		for _, q := range queries {
			filename := fmt.Sprintf("%d.sql", q.ID)
			filePath := filepath.Join(sqlDir, filename)
			logger.Debug("Writing query to file", "id", q.ID, "name", q.Name, "file", filePath)

			if err := os.WriteFile(filePath, []byte(q.Query), 0644); err != nil {
				logger.Error("Failed to write query to file", "id", q.ID, "file", filePath, "error", err)
				return err
			}
		}

		logger.Info("All queries dumped successfully", "dir", sqlDir)
		fmt.Printf("All queries dumped to %s\nJSON list saved as %s\n", sqlDir, jsonFilename)
		return nil
	},
}
