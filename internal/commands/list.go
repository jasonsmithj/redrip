package commands

import (
	"encoding/json"
	"fmt"

	"github.com/jasonsmithj/redrip/internal/logger"
	"github.com/jasonsmithj/redrip/internal/redash"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Redash queries",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting list command")

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

		if outputFormat == "json" {
			// Output as JSON
			jsonOutput, err := json.MarshalIndent(queries, "", "  ")
			if err != nil {
				logger.Error("Failed to marshal queries to JSON", "error", err)
				return fmt.Errorf("failed to marshal queries to JSON: %v", err)
			}
			fmt.Println(string(jsonOutput))
		} else {
			// Output in plain text format
			for _, q := range queries {
				logger.Debug("Query", "id", q.ID, "name", q.Name)
				fmt.Printf("ID: %d\tName: %s\n", q.ID, q.Name)
			}
		}

		logger.Info("Finished listing queries", "count", len(queries))
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format: json or text")
}
