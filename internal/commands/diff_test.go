package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jasonsmithj/redrip/internal/diff"
)

func TestDiffPathHandling(t *testing.T) {
	// Create temporary directory and file for testing
	tempDir := t.TempDir()
	filename := "123.sql"
	filePath := filepath.Join(tempDir, filename)

	// Create a test SQL file
	testSQL := "SELECT * FROM users WHERE id = 1"
	err := os.WriteFile(filePath, []byte(testSQL), 0644)
	if err != nil {
		t.Fatalf("Failed to create test SQL file: %v", err)
	}

	// Verify file was created
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Fatalf("Test SQL file was not created: %v", err)
	}

	// Test file path construction
	testFilePath := filepath.Join(tempDir, "123.sql")
	if testFilePath != filePath {
		t.Errorf("Path mismatch: expected %s, got %s", filePath, testFilePath)
	}

	// Test file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	if string(content) != testSQL {
		t.Errorf("Content mismatch: expected %q, got %q", testSQL, string(content))
	}
}

func TestDiffResultJSONFormat(t *testing.T) {
	// Create a test DiffResult
	result := diff.Result{
		QueryID:   123,
		QueryName: "Test Query",
		Status:    "MATCH",
		LocalPath: "/path/to/123.sql",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal DiffResult to JSON: %v", err)
	}

	// Unmarshal back to verify structure
	var unmarshaledResult diff.Result
	err = json.Unmarshal(jsonData, &unmarshaledResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to DiffResult: %v", err)
	}

	// Verify fields
	if unmarshaledResult.QueryID != result.QueryID {
		t.Errorf("QueryID mismatch: expected %d, got %d", result.QueryID, unmarshaledResult.QueryID)
	}
	if unmarshaledResult.QueryName != result.QueryName {
		t.Errorf("QueryName mismatch: expected %s, got %s", result.QueryName, unmarshaledResult.QueryName)
	}
	if unmarshaledResult.Status != result.Status {
		t.Errorf("Status mismatch: expected %s, got %s", result.Status, unmarshaledResult.Status)
	}
	if unmarshaledResult.LocalPath != result.LocalPath {
		t.Errorf("LocalPath mismatch: expected %s, got %s", result.LocalPath, unmarshaledResult.LocalPath)
	}
}

func TestDiffSummaryJSONFormat(t *testing.T) {
	// Create a test DiffSummary
	summary := diff.Summary{
		Profile:         "test",
		SQLDirectory:    "/path/to/sql",
		Matches:         10,
		Differences:     5,
		MissingInRedash: 2,
		Results: []diff.Result{
			{
				QueryID:   123,
				QueryName: "Test Query 1",
				Status:    "MATCH",
				LocalPath: "/path/to/123.sql",
			},
			{
				QueryID:     456,
				QueryName:   "Test Query 2",
				Status:      "DIFFERENT",
				LocalPath:   "/path/to/456.sql",
				Differences: "Sample differences",
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Failed to marshal DiffSummary to JSON: %v", err)
	}

	// Unmarshal back to verify structure
	var unmarshaledSummary diff.Summary
	err = json.Unmarshal(jsonData, &unmarshaledSummary)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to DiffSummary: %v", err)
	}

	// Verify fields
	if unmarshaledSummary.Profile != summary.Profile {
		t.Errorf("Profile mismatch: expected %s, got %s", summary.Profile, unmarshaledSummary.Profile)
	}
	if unmarshaledSummary.SQLDirectory != summary.SQLDirectory {
		t.Errorf("SQLDirectory mismatch: expected %s, got %s", summary.SQLDirectory, unmarshaledSummary.SQLDirectory)
	}
	if unmarshaledSummary.Matches != summary.Matches {
		t.Errorf("Matches mismatch: expected %d, got %d", summary.Matches, unmarshaledSummary.Matches)
	}
	if unmarshaledSummary.Differences != summary.Differences {
		t.Errorf("Differences mismatch: expected %d, got %d", summary.Differences, unmarshaledSummary.Differences)
	}
	if unmarshaledSummary.MissingInRedash != summary.MissingInRedash {
		t.Errorf("MissingInRedash mismatch: expected %d, got %d", summary.MissingInRedash, unmarshaledSummary.MissingInRedash)
	}
	if len(unmarshaledSummary.Results) != len(summary.Results) {
		t.Errorf("Results length mismatch: expected %d, got %d", len(summary.Results), len(unmarshaledSummary.Results))
	}
}
