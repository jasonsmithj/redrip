// Package redash provides functionality for interacting with the Redash API.
// It handles configuration, authentication, and API calls for retrieving and managing Redash queries.
package redash

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonsmithj/redrip/internal/logger"
)

// DefaultConfigContent is the default content for the config file
const DefaultConfigContent = `# Redash API URL (required)
redash_url = 
# Redash API Key (required)
api_key = 
# Directory to save SQL files (optional, defaults to current directory)
sql_dir = 
`

// Config holds configuration for the Redash client
type Config struct {
	RedashURL string
	APIKey    string
	SQLDir    string
}

// EnsureConfigFile ensures that the config file exists, creating it if necessary
func EnsureConfigFile(configPath string) error {
	logger.Debug("Ensuring config file exists", "path", configPath)

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("Config file does not exist, creating it", "path", configPath)

		// Create directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			logger.Error("Failed to create config directory", "dir", configDir, "error", err)
			return fmt.Errorf("failed to create config directory: %v", err)
		}

		// Create config file with default content
		if err := os.WriteFile(configPath, []byte(DefaultConfigContent), 0644); err != nil {
			logger.Error("Failed to create config file", "path", configPath, "error", err)
			return fmt.Errorf("failed to create config file: %v", err)
		}

		logger.Info("Created default config file", "path", configPath)
		logger.Warn("Please edit the config file to set your Redash URL and API Key", "path", configPath)
		fmt.Printf("Created default config file at %s\nPlease edit it to set your Redash URL and API Key\n", configPath)
	}

	return nil
}

// LoadConfig loads configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	logger.Debug("Loading configuration", "path", configPath)

	// Ensure config file exists
	if err := EnsureConfigFile(configPath); err != nil {
		return nil, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		logger.Error("Failed to open config file", "path", configPath, "error", err)
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error("Failed to close config file", "path", configPath, "error", err)
		}
	}()

	config := &Config{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "redash_url":
			config.RedashURL = value
			logger.Debug("Config loaded", "key", "redash_url", "value", value)
		case "api_key":
			config.APIKey = value
			logger.Debug("Config loaded", "key", "api_key", "value", "[REDACTED]")
		case "sql_dir":
			config.SQLDir = value
			logger.Debug("Config loaded", "key", "sql_dir", "value", value)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error scanning config file", "error", err)
		return nil, err
	}

	// Check if required values are missing and provide helpful messages
	if config.RedashURL == "" || config.APIKey == "" {
		logger.Error("Required configuration values missing",
			"redash_url_set", config.RedashURL != "",
			"api_key_set", config.APIKey != "")

		missingFields := []string{}
		if config.RedashURL == "" {
			missingFields = append(missingFields, "redash_url")
		}
		if config.APIKey == "" {
			missingFields = append(missingFields, "api_key")
		}

		missingMsg := fmt.Sprintf("Missing required configuration: %s", strings.Join(missingFields, ", "))
		logger.Warn(missingMsg)
		logger.Warn("Please edit your config file", "path", configPath)
		fmt.Printf("Error: %s\nPlease edit %s to set these values\n", missingMsg, configPath)

		return nil, fmt.Errorf("required configuration values not found: %s", strings.Join(missingFields, ", "))
	}

	logger.Info("Configuration loaded successfully")
	return config, nil
}

// GetSQLDir returns the configured SQL directory or current directory if not set
func GetSQLDir() (string, error) {
	logger.Debug("Getting SQL directory from config")

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get home directory", "error", err)
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	// Load configuration from ~/.redrip/config.conf
	configPath := filepath.Join(homeDir, ".redrip", "config.conf")
	config, err := LoadConfig(configPath)
	if err != nil {
		// If error is about missing required fields, we still want to return a valid directory
		if strings.Contains(err.Error(), "required configuration values not found") {
			logger.Info("Using current directory due to missing config values")
			return ".", nil
		}

		logger.Error("Failed to load configuration", "error", err)
		return "", fmt.Errorf("failed to load configuration: %v", err)
	}

	// If SQLDir is not set or doesn't exist, use current directory
	if config.SQLDir == "" {
		logger.Info("SQL directory not set, using current directory")
		return ".", nil
	}

	// Check if the directory exists
	if _, err := os.Stat(config.SQLDir); os.IsNotExist(err) {
		logger.Warn("SQL directory does not exist, using current directory",
			"configured_dir", config.SQLDir)
		return ".", nil
	}

	logger.Debug("Using configured SQL directory", "path", config.SQLDir)
	return config.SQLDir, nil
}

// Client represents a connection to a Redash instance.
// It handles API requests and authentication using the configured API key.
type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// NewClient creates a new Redash client instance
func NewClient() (*Client, error) {
	logger.Debug("Creating new Redash client")

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get home directory", "error", err)
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Load configuration from ~/.redrip/config.conf
	configPath := filepath.Join(homeDir, ".redrip", "config.conf")
	config, err := LoadConfig(configPath)
	if err != nil {
		// If the error is about missing required fields, we want to provide a clear error message
		if strings.Contains(err.Error(), "required configuration values not found") {
			return nil, fmt.Errorf("cannot create Redash client: %v", err)
		}
		logger.Error("Failed to load configuration", "error", err)
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	logger.Info("Redash client created", "url", config.RedashURL)
	return &Client{
		client:  &http.Client{},
		baseURL: config.RedashURL,
		apiKey:  config.APIKey,
	}, nil
}

// Query represents a Redash query with its metadata and SQL content.
type Query struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Query string `json:"query"`
}

type queryListResponse struct {
	Results  []Query `json:"results"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	Count    int     `json:"count"`
}

// ListQueries retrieves all queries from the Redash instance.
// It handles pagination automatically to fetch all available queries.
func (c *Client) ListQueries() ([]Query, error) {
	logger.Debug("Listing queries")

	var allQueries []Query
	page := 1
	pageSize := 100

	for {
		logger.Debug("Fetching page of queries", "page", page, "page_size", pageSize)

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/queries?page=%d&page_size=%d", c.baseURL, page, pageSize), nil)
		req.Header.Add("Authorization", fmt.Sprintf("Key %s", c.apiKey))

		resp, err := c.client.Do(req)
		if err != nil {
			logger.Error("Failed to execute request", "error", err)
			return nil, err
		}

		body, _ := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			logger.Error("Failed to close response body", "error", err)
		}

		var result queryListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			logger.Error("Failed to unmarshal response", "error", err)
			return nil, err
		}

		logger.Debug("Page received",
			"page", page,
			"results_count", len(result.Results),
			"count", result.Count)

		allQueries = append(allQueries, result.Results...)

		// Check if we've fetched all pages
		if len(result.Results) == 0 || len(allQueries) >= result.Count {
			break
		}

		page++
	}

	logger.Info("All queries retrieved", "count", len(allQueries))
	return allQueries, nil
}

// GetQuery retrieves a specific query by its ID from the Redash instance.
// It returns the query details including the SQL content.
func (c *Client) GetQuery(id int) (*Query, error) {
	logger.Debug("Getting query", "id", id, "url", fmt.Sprintf("%s/queries/%d", c.baseURL, id))

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/queries/%d", c.baseURL, id), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Key %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error("Failed to execute request", "id", id, "error", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}()

	logger.Debug("Response received", "status", resp.Status)

	body, _ := io.ReadAll(resp.Body)
	var result Query
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Failed to unmarshal response", "error", err)
		return nil, err
	}

	logger.Info("Query retrieved", "id", result.ID, "name", result.Name)
	return &result, nil
}
