// Package redash provides functionality for interacting with the Redash API.
// It handles configuration, authentication, and API calls for retrieving and managing Redash queries.
package redash

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonsmithj/redrip/internal/file"
	"github.com/jasonsmithj/redrip/internal/logger"
)

// DefaultConfigContent is the default content for the config file
const DefaultConfigContent = `# Default profile (used when no profile is specified)
[default]
# Redash API URL (required)
redash_url = 
# Redash API Key (required)
api_key = 
# Directory to save SQL files (optional, defaults to current directory)
sql_dir = 

# Example staging profile
# [profile stg]
# redash_url = https://redash-staging.example.com/api
# api_key = your_staging_api_key
# sql_dir = /path/to/staging/sql/dir

# Example production profile
# [profile prd]
# redash_url = https://redash-production.example.com/api
# api_key = your_production_api_key
# sql_dir = /path/to/production/sql/dir
`

// ProfileConfig holds configuration for a single profile
type ProfileConfig struct {
	RedashURL string
	APIKey    string
	SQLDir    string
}

// Config holds configuration for the Redash client including multiple profiles
type Config struct {
	Profiles map[string]ProfileConfig
}

// CurrentProfile is the active profile name being used
var CurrentProfile string

// EnsureConfigFile ensures that the config file exists, creating it if necessary
func EnsureConfigFile(configPath string) error {
	logger.Debug("Ensuring config file exists", "path", configPath)

	// Check if file exists
	if !file.Exists(configPath) {
		logger.Info("Config file does not exist, creating it", "path", configPath)

		// Write config file with default content
		if err := file.WriteFile(configPath, []byte(DefaultConfigContent), 0644); err != nil {
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

	config := &Config{
		Profiles: make(map[string]ProfileConfig),
	}

	var currentProfile string = "default"

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Check for profile section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			profileName := strings.TrimSpace(line[1 : len(line)-1])

			// Handle profile prefix
			if strings.HasPrefix(profileName, "profile ") {
				profileName = strings.TrimSpace(profileName[8:])
			} else if profileName != "default" {
				// If it doesn't have "profile " prefix and it's not "default",
				// it's not a valid profile section
				logger.Warn("Invalid profile section", "section", line,
					"hint", "Profile sections should be [default] or [profile name]")
				continue
			}

			currentProfile = profileName

			// Initialize profile if it doesn't exist
			if _, exists := config.Profiles[currentProfile]; !exists {
				config.Profiles[currentProfile] = ProfileConfig{}
			}

			continue
		}

		// Parse key-value settings
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Get the current profile config
		profileConfig := config.Profiles[currentProfile]

		// Update the appropriate field
		switch key {
		case "redash_url":
			profileConfig.RedashURL = value
			logger.Debug("Config loaded", "profile", currentProfile, "key", "redash_url", "value", value)
		case "api_key":
			profileConfig.APIKey = value
			logger.Debug("Config loaded", "profile", currentProfile, "key", "api_key", "value", "[REDACTED]")
		case "sql_dir":
			profileConfig.SQLDir = value
			logger.Debug("Config loaded", "profile", currentProfile, "key", "sql_dir", "value", value)
		}

		// Update the profile in the map
		config.Profiles[currentProfile] = profileConfig
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error scanning config file", "error", err)
		return nil, err
	}

	// Ensure default profile exists
	if _, exists := config.Profiles["default"]; !exists {
		config.Profiles["default"] = ProfileConfig{}
	}

	logger.Info("Configuration loaded successfully")
	return config, nil
}

// GetProfileConfig returns the config for the specified profile
func GetProfileConfig(config *Config, profileName string) *ProfileConfig {
	// If profile name is empty, check environment variable
	if profileName == "" {
		profileName = os.Getenv("REDRIP_PROFILE")
		if profileName != "" {
			logger.Debug("Using profile from environment variable", "profile", profileName)
		}
	}

	// If still empty, use default
	if profileName == "" {
		profileName = "default"
	}

	// Check if profile exists
	profileConfig, exists := config.Profiles[profileName]
	if !exists {
		logger.Warn("Profile does not exist, using default", "requested_profile", profileName)
		profileConfig = config.Profiles["default"]
		profileName = "default"
	}

	// Store the current profile name for later use
	CurrentProfile = profileName

	logger.Debug("Using profile", "profile", profileName)
	return &profileConfig
}

// ValidateProfileConfig checks if required values are missing
func ValidateProfileConfig(profileConfig *ProfileConfig) error {
	// Check if required values are missing and provide helpful messages
	if profileConfig.RedashURL == "" || profileConfig.APIKey == "" {
		logger.Error("Required configuration values missing",
			"redash_url_set", profileConfig.RedashURL != "",
			"api_key_set", profileConfig.APIKey != "")

		missingFields := []string{}
		if profileConfig.RedashURL == "" {
			missingFields = append(missingFields, "redash_url")
		}
		if profileConfig.APIKey == "" {
			missingFields = append(missingFields, "api_key")
		}

		missingMsg := fmt.Sprintf("Missing required configuration: %s", strings.Join(missingFields, ", "))
		logger.Warn(missingMsg)
		logger.Warn("Please edit your config file to set these values for the current profile", "profile", CurrentProfile)

		return fmt.Errorf("required configuration values not found: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// GetProfileSQLDir returns the configured SQL directory for the specified profile or current directory if not set
func GetProfileSQLDir(profileName string) (string, error) {
	logger.Debug("Getting SQL directory from config", "profile", profileName)

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

	// Get profile config
	profileConfig := GetProfileConfig(config, profileName)

	// If SQLDir is not set or doesn't exist, use current directory
	if profileConfig.SQLDir == "" {
		logger.Info("SQL directory not set, using current directory", "profile", CurrentProfile)
		return ".", nil
	}

	// Check if the directory exists
	if !file.Exists(profileConfig.SQLDir) || !file.IsDirectory(profileConfig.SQLDir) {
		logger.Warn("SQL directory does not exist, using current directory",
			"profile", CurrentProfile, "configured_dir", profileConfig.SQLDir)
		return ".", nil
	}

	logger.Debug("Using configured SQL directory", "profile", CurrentProfile, "path", profileConfig.SQLDir)
	return profileConfig.SQLDir, nil
}

// GetSQLDir returns the configured SQL directory or current directory if not set
// This is maintained for backward compatibility
func GetSQLDir() (string, error) {
	return GetProfileSQLDir("")
}

// Client represents a connection to a Redash instance.
// It handles API requests and authentication using the configured API key.
type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
	profile string
}

// NewClientWithProfile creates a new Redash client instance for the specified profile
func NewClientWithProfile(profileName string) (*Client, error) {
	logger.Debug("Creating new Redash client", "profile", profileName)

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

	// Get profile config
	profileConfig := GetProfileConfig(config, profileName)

	// Validate profile config
	if err := ValidateProfileConfig(profileConfig); err != nil {
		return nil, fmt.Errorf("cannot create Redash client for profile '%s': %v", CurrentProfile, err)
	}

	logger.Info("Redash client created", "profile", CurrentProfile, "url", profileConfig.RedashURL)
	return &Client{
		client:  &http.Client{},
		baseURL: profileConfig.RedashURL,
		apiKey:  profileConfig.APIKey,
		profile: CurrentProfile,
	}, nil
}

// NewClient creates a new Redash client instance using the default or environment-specified profile
// This is maintained for backward compatibility
func NewClient() (*Client, error) {
	return NewClientWithProfile("")
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
			return nil, fmt.Errorf("failed to execute request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			// 非200レスポンスの場合、レスポンスボディの内容を診断用にログに出力
			body, _ := io.ReadAll(resp.Body)
			contentPreview := string(body)
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:200] + "..."
			}
			logger.Error("Received non-200 response", "status", resp.StatusCode, "response_preview", contentPreview)
			resp.Body.Close()
			return nil, fmt.Errorf("received non-200 response: %d (content: %s)", resp.StatusCode, contentPreview)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Error("Failed to read response body", "error", err)
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		var response queryListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			// HTMLレスポンスの場合、より具体的なエラーメッセージを提供
			if bytes.HasPrefix(body, []byte("<")) {
				logger.Error("Received HTML instead of JSON", "response_preview", string(body[:100]))
				return nil, fmt.Errorf("received HTML instead of JSON. This may indicate authentication issues or an incorrect URL. Please check your API key and Redash URL")
			}
			logger.Error("Failed to unmarshal response", "error", err)
			return nil, fmt.Errorf("failed to unmarshal response: %v", err)
		}

		logger.Debug("Fetched queries", "count", len(response.Results), "total", response.Count)
		allQueries = append(allQueries, response.Results...)

		// Check if we've fetched all pages
		if len(allQueries) >= response.Count || len(response.Results) == 0 {
			break
		}

		page++
	}

	logger.Info("Retrieved all queries", "count", len(allQueries))
	return allQueries, nil
}

// GetQuery retrieves a single query by ID.
func (c *Client) GetQuery(id int) (*Query, error) {
	logger.Debug("Getting query", "id", id)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/queries/%d", c.baseURL, id), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Key %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error("Failed to execute request", "error", err)
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// 非200レスポンスの場合、レスポンスボディの内容を診断用にログに出力
		body, _ := io.ReadAll(resp.Body)
		contentPreview := string(body)
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		logger.Error("Received non-200 response", "status", resp.StatusCode, "response_preview", contentPreview)
		resp.Body.Close()
		return nil, fmt.Errorf("received non-200 response: %d (content: %s)", resp.StatusCode, contentPreview)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var query Query
	if err := json.Unmarshal(body, &query); err != nil {
		// HTMLレスポンスの場合、より具体的なエラーメッセージを提供
		if bytes.HasPrefix(body, []byte("<")) {
			logger.Error("Received HTML instead of JSON", "response_preview", string(body[:100]))
			return nil, fmt.Errorf("received HTML instead of JSON. This may indicate authentication issues or an incorrect URL. Please check your API key and Redash URL")
		}
		logger.Error("Failed to unmarshal response", "error", err)
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	logger.Info("Retrieved query", "id", query.ID, "name", query.Name)
	return &query, nil
}
