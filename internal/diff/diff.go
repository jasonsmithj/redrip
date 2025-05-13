// Package diff provides utilities for comparing SQL content between local files and Redash
package diff

import (
	"fmt"
	"os"
	"strings"

	"github.com/jasonsmithj/redrip/internal/file"
	"github.com/jasonsmithj/redrip/internal/redash"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Result represents the result of a diff operation
type Result struct {
	QueryID      int    `json:"query_id"`
	QueryName    string `json:"query_name"`
	Status       string `json:"status"` // "MATCH", "DIFFERENT", "MISSING_IN_REDASH", "ERROR"
	ErrorMessage string `json:"error_message,omitempty"`
	LocalPath    string `json:"local_path,omitempty"`
	Differences  string `json:"differences,omitempty"`
}

// Summary represents a summary of diff operations
type Summary struct {
	Profile         string   `json:"profile"`
	SQLDirectory    string   `json:"sql_directory"`
	Matches         int      `json:"matches"`
	Differences     int      `json:"differences"`
	MissingInRedash int      `json:"missing_in_redash"`
	Results         []Result `json:"results"`
}

// CompareQueryWithLocal compares a local SQL file with a Redash query and returns a Result
func CompareQueryWithLocal(queryID int, redashQuery *redash.Query, localPath string) (Result, error) {
	result := Result{
		QueryID:   queryID,
		LocalPath: localPath,
	}

	// Check if local file exists
	if !file.Exists(localPath) || !file.IsFile(localPath) {
		return result, fmt.Errorf("local SQL file does not exist: %s", localPath)
	}

	// If we have a query from Redash, set the name
	if redashQuery != nil {
		result.QueryName = redashQuery.Name
	}

	// Read local file content
	localContent, err := os.ReadFile(localPath)
	if err != nil {
		return result, fmt.Errorf("failed to read local file: %v", err)
	}

	// If no Redash query, it's missing in Redash
	if redashQuery == nil {
		result.Status = "MISSING_IN_REDASH"
		return result, nil
	}

	// Compare contents
	localSQL := strings.TrimSpace(string(localContent))
	redashSQL := strings.TrimSpace(redashQuery.Query)

	if localSQL == redashSQL {
		result.Status = "MATCH"
		return result, nil
	}

	// Generate diff details
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(localSQL, redashSQL, false)
	result.Status = "DIFFERENT"
	result.Differences = dmp.DiffPrettyText(diffs)

	return result, nil
}

// HandleCommonAPIErrors checks for common API errors and provides user-friendly messages
func HandleCommonAPIErrors(err error) {
	redash.PrintCommonErrorSuggestions(err)
}
