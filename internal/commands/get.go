package commands

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/jasonsmithj/redrip/internal/file"
	"github.com/jasonsmithj/redrip/internal/logger"
	"github.com/jasonsmithj/redrip/internal/redash"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <query_id>",
	Args:  cobra.ExactArgs(1),
	Short: "Get SQL for a specific query and save it as a file",
	RunE: func(_ *cobra.Command, args []string) error {
		logger.Info("Starting get command", "queryID", args[0], "profile", profile)

		queryID, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Error("Invalid query ID", "input", args[0], "error", err)
			return err
		}
		logger.Debug("Parsed query ID", "id", queryID)

		client, err := redash.NewClientWithProfile(profile)
		if err != nil {
			logger.Error("Failed to initialize Redash client", "error", err)
			return fmt.Errorf("failed to initialize Redash client: %v", err)
		}

		logger.Debug("Fetching query from Redash", "id", queryID)
		query, err := client.GetQuery(queryID)
		if err != nil {
			logger.Error("Failed to get query", "id", queryID, "error", err)
			redash.PrintCommonErrorSuggestions(err)
			return err
		}
		logger.Info("Retrieved query from Redash", "id", query.ID, "name", query.Name)

		// Get configured SQL directory
		sqlDir, err := redash.GetProfileSQLDir(profile)
		if err != nil {
			logger.Error("Failed to get SQL directory", "error", err)
			return fmt.Errorf("failed to get SQL directory: %v", err)
		}
		logger.Debug("Using SQL directory", "dir", sqlDir)

		// Create filename with path
		filename := fmt.Sprintf("%d.sql", query.ID)
		filePath := filepath.Join(sqlDir, filename)
		logger.Debug("File path", "path", filePath)

		// Write query to file
		logger.Debug("Writing query to file", "file", filePath)
		if err := file.WriteFile(filePath, []byte(query.Query), 0644); err != nil {
			logger.Error("Failed to write file", "file", filePath, "error", err)
			return fmt.Errorf("failed to write file: %v", err)
		}

		logger.Info("Query saved to file", "file", filePath)
		fmt.Printf("Query %d (%s) saved to %s\n", query.ID, query.Name, filePath)
		return nil
	},
}
