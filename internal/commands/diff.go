package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jasonsmithj/redrip/internal/diff"
	"github.com/jasonsmithj/redrip/internal/file"
	"github.com/jasonsmithj/redrip/internal/logger"
	"github.com/jasonsmithj/redrip/internal/redash"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare local SQL files with Redash queries",
}

var diffAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Compare all local SQL files with Redash queries",
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Info("Starting diff all command", "profile", profile)

		// Get Redash client
		client, err := redash.NewClientWithProfile(profile)
		if err != nil {
			logger.Error("Failed to initialize Redash client", "error", err)
			return fmt.Errorf("failed to initialize Redash client: %v", err)
		}

		// Get configured SQL directory
		sqlDir, err := redash.GetProfileSQLDir(profile)
		if err != nil {
			logger.Error("Failed to get SQL directory", "error", err)
			return fmt.Errorf("failed to get SQL directory: %v", err)
		}
		logger.Debug("Using SQL directory", "dir", sqlDir)

		// Check if SQL directory exists
		if !file.Exists(sqlDir) || !file.IsDirectory(sqlDir) {
			logger.Error("SQL directory does not exist", "dir", sqlDir)
			return fmt.Errorf("SQL directory does not exist: %s", sqlDir)
		}

		// Fetch all queries from Redash
		logger.Debug("Fetching queries from Redash")
		queries, err := client.ListQueries()
		if err != nil {
			logger.Error("Failed to list queries", "error", err)
			diff.HandleCommonAPIErrors(err)
			return err
		}
		logger.Info("Retrieved queries from Redash", "count", len(queries))

		// Create map of queries by ID for easy lookup
		queryMap := make(map[int]redash.Query)
		for _, q := range queries {
			queryMap[q.ID] = q
		}

		// Create summary for results
		summary := diff.Summary{
			Profile:      redash.CurrentProfile,
			SQLDirectory: sqlDir,
			Results:      []diff.Result{},
		}

		// Check each SQL file in the directory
		entries, err := os.ReadDir(sqlDir)
		if err != nil {
			logger.Error("Failed to read directory", "dir", sqlDir, "error", err)
			return fmt.Errorf("failed to read directory %s: %v", sqlDir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				continue
			}

			// Extract query ID from filename (assuming format is <id>.sql)
			idStr := strings.TrimSuffix(entry.Name(), ".sql")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				logger.Debug("Skipping file with non-numeric ID", "file", entry.Name())
				continue
			}

			localPath := filepath.Join(sqlDir, entry.Name())

			// Get the query from map if it exists
			redashQuery, exists := queryMap[id]
			var queryPtr *redash.Query
			if exists {
				queryPtr = &redashQuery
			}

			// Compare local and Redash query
			result, err := diff.CompareQueryWithLocal(id, queryPtr, localPath)
			if err != nil {
				logger.Error("Error comparing query", "id", id, "error", err)
				result.Status = "ERROR"
				result.ErrorMessage = err.Error()
			}

			// Update counters based on result status
			switch result.Status {
			case "MATCH":
				logger.Debug("No differences found", "id", id, "name", result.QueryName)
				summary.Matches++
			case "DIFFERENT":
				logger.Info("Differences found", "id", id, "name", result.QueryName)
				summary.Differences++
			case "MISSING_IN_REDASH":
				logger.Warn("Local query does not exist in Redash", "id", id, "file", entry.Name())
				summary.MissingInRedash++
			}

			summary.Results = append(summary.Results, result)
		}

		// Output as JSON
		jsonOutput, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			logger.Error("Failed to marshal results to JSON", "error", err)
			return fmt.Errorf("failed to marshal results to JSON: %v", err)
		}
		fmt.Println(string(jsonOutput))

		return nil
	},
}

var diffQueryCmd = &cobra.Command{
	Use:   "query <query_id>",
	Args:  cobra.ExactArgs(1),
	Short: "Compare a specific local SQL file with the corresponding Redash query",
	RunE: func(_ *cobra.Command, args []string) error {
		logger.Info("Starting diff query command", "queryID", args[0], "profile", profile)

		queryID, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Error("Invalid query ID", "input", args[0], "error", err)
			return fmt.Errorf("invalid query ID: %s", args[0])
		}
		logger.Debug("Parsed query ID", "id", queryID)

		// Get Redash client
		client, err := redash.NewClientWithProfile(profile)
		if err != nil {
			logger.Error("Failed to initialize Redash client", "error", err)
			return fmt.Errorf("failed to initialize Redash client: %v", err)
		}

		// Get configured SQL directory
		sqlDir, err := redash.GetProfileSQLDir(profile)
		if err != nil {
			logger.Error("Failed to get SQL directory", "error", err)
			return fmt.Errorf("failed to get SQL directory: %v", err)
		}
		logger.Debug("Using SQL directory", "dir", sqlDir)

		// Prepare paths
		filename := fmt.Sprintf("%d.sql", queryID)
		localPath := filepath.Join(sqlDir, filename)

		// Create an empty result struct
		result := diff.Result{
			QueryID:   queryID,
			LocalPath: localPath,
		}

		// Check if local file exists
		if !file.Exists(localPath) {
			logger.Error("Local SQL file does not exist", "file", localPath)
			result.Status = "ERROR"
			result.ErrorMessage = fmt.Sprintf("local SQL file does not exist: %s", localPath)
			jsonOutput, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(jsonOutput))
			return nil
		}

		// Get query from Redash
		logger.Debug("Fetching query from Redash", "id", queryID)
		redashQuery, err := client.GetQuery(queryID)
		if err != nil {
			logger.Error("Failed to get query from Redash", "id", queryID, "error", err)
			diff.HandleCommonAPIErrors(err)
			result.Status = "ERROR"
			result.ErrorMessage = fmt.Sprintf("failed to get query from Redash: %v", err)
			jsonOutput, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(jsonOutput))
			return nil
		}

		logger.Info("Retrieved query from Redash", "id", queryID, "name", redashQuery.Name)

		// Compare the query
		result, err = diff.CompareQueryWithLocal(queryID, redashQuery, localPath)
		if err != nil {
			logger.Error("Error comparing query", "id", queryID, "error", err)
			result.Status = "ERROR"
			result.ErrorMessage = err.Error()
		}

		// Output as JSON
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))

		return nil
	},
}

func init() {
	diffCmd.AddCommand(diffAllCmd)
	diffCmd.AddCommand(diffQueryCmd)
}
