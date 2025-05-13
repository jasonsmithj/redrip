package redash

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// テスト用の設定ファイルを作成
	configPath := filepath.Join(tempDir, "config.conf")
	configData := `redash_url = https://test-redash.com/api
api_key = test-api-key
sql_dir = /tmp/sql`

	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// テスト実行
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	// 結果の検証
	// リファクタリング後は直接アクセスではなくプロファイル経由でアクセスする
	defaultProfile := config.Profiles["default"]
	if defaultProfile.RedashURL != "https://test-redash.com/api" {
		t.Errorf("Expected RedashURL = %s, got %s", "https://test-redash.com/api", defaultProfile.RedashURL)
	}
	if defaultProfile.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey = %s, got %s", "test-api-key", defaultProfile.APIKey)
	}
	if defaultProfile.SQLDir != "/tmp/sql" {
		t.Errorf("Expected SQLDir = %s, got %s", "/tmp/sql", defaultProfile.SQLDir)
	}
}

func TestLoadConfigMissingRequired(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// 必須項目が欠けたテスト用の設定ファイルを作成
	configPath := filepath.Join(tempDir, "config.conf")
	configData := `sql_dir = /tmp/sql` // RedashURLとAPIKeyがない

	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// テスト実行
	config, err := LoadConfig(configPath)
	// LoadConfigはエラーを返さないが、その後のValidateProfileConfigでエラーになるはず
	if err != nil {
		t.Fatalf("LoadConfig returned unexpected error: %v", err)
	}

	// デフォルトプロファイルを取得して検証
	profileConfig := GetProfileConfig(config, "default")
	err = ValidateProfileConfig(profileConfig)
	if err == nil {
		t.Error("ValidateProfileConfig should return error when required values are missing")
	}
}

// GetSQLDirとNewClientのテストはユーザーのホームディレクトリに依存するため、
// モックを利用したり、テスト用の関数を作成するとよいです。
// ここでは簡単な例として、テスト用の関数を用意します。

// テスト用のクライアント作成関数
func newTestClient(baseURL, apiKey string) *Client {
	return &Client{
		client:  &http.Client{},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func TestListQueries(t *testing.T) {
	// モックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization ヘッダーの検証
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Key test-api-key" {
			t.Errorf("Expected Authorization header = %s, got %s", "Key test-api-key", authHeader)
		}

		// パスの検証
		if r.URL.Path != "/queries" {
			t.Errorf("Expected path = %s, got %s", "/queries", r.URL.Path)
		}

		// レスポンスの作成
		response := queryListResponse{
			Results: []Query{
				{ID: 1, Name: "Test Query 1", Query: "SELECT * FROM test1"},
				{ID: 2, Name: "Test Query 2", Query: "SELECT * FROM test2"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// テスト用のクライアントを作成
	client := newTestClient(server.URL, "test-api-key")

	// テスト実行
	queries, err := client.ListQueries()
	if err != nil {
		t.Fatalf("ListQueries returned error: %v", err)
	}

	// 結果の検証
	if len(queries) != 2 {
		t.Errorf("Expected 2 queries, got %d", len(queries))
	}
	if queries[0].ID != 1 || queries[0].Name != "Test Query 1" || queries[0].Query != "SELECT * FROM test1" {
		t.Errorf("Query 1 data does not match expected values")
	}
	if queries[1].ID != 2 || queries[1].Name != "Test Query 2" || queries[1].Query != "SELECT * FROM test2" {
		t.Errorf("Query 2 data does not match expected values")
	}
}

func TestGetQuery(t *testing.T) {
	// モックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization ヘッダーの検証
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Key test-api-key" {
			t.Errorf("Expected Authorization header = %s, got %s", "Key test-api-key", authHeader)
		}

		// パスの検証
		if r.URL.Path != "/queries/1" {
			t.Errorf("Expected path = %s, got %s", "/queries/1", r.URL.Path)
		}

		// レスポンスの作成
		query := Query{ID: 1, Name: "Test Query", Query: "SELECT * FROM test"}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(query); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// テスト用のクライアントを作成
	client := newTestClient(server.URL, "test-api-key")

	// テスト実行
	query, err := client.GetQuery(1)
	if err != nil {
		t.Fatalf("GetQuery returned error: %v", err)
	}

	// 結果の検証
	if query.ID != 1 || query.Name != "Test Query" || query.Query != "SELECT * FROM test" {
		t.Errorf("Query data does not match expected values")
	}
}

func TestGetQueryError(t *testing.T) {
	// エラーを返すモックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// 500エラーを返す
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, "Internal Server Error")
		if err != nil {
			t.Logf("Error writing error response: %v", err)
		}
	}))
	defer server.Close()

	// テスト用のクライアントを作成
	client := newTestClient(server.URL, "test-api-key")

	// テスト実行
	_, err := client.GetQuery(1)
	if err == nil {
		t.Fatal("GetQuery should return error on server failure")
	}
}

func TestEnsureConfigFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.conf")

	// 設定ファイルが作成されることを確認
	err := EnsureConfigFile(configPath)
	if err != nil {
		t.Fatalf("EnsureConfigFile returned error: %v", err)
	}

	// ファイルが存在することを確認
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// ファイルの内容を確認
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "redash_url =") ||
		!strings.Contains(content, "api_key =") ||
		!strings.Contains(content, "sql_dir =") {
		t.Error("Config file does not contain expected content")
	}

	// 既存のファイルを変更しないことを確認
	customContent := "redash_url = https://test-redash.com"
	if err := os.WriteFile(configPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to write to config file: %v", err)
	}

	// 再度実行
	err = EnsureConfigFile(configPath)
	if err != nil {
		t.Fatalf("EnsureConfigFile returned error on existing file: %v", err)
	}

	// 内容が変更されていないことを確認
	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(data) != customContent {
		t.Error("Existing config file was modified")
	}
}

func TestLoadConfigWithEmptyValues(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.conf")

	// 空の値を持つ設定ファイルを作成
	configData := `
redash_url = 
api_key = 
sql_dir = /tmp/sql
`
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// テスト実行
	config, err := LoadConfig(configPath)
	// LoadConfigはエラーを返さないはず
	if err != nil {
		t.Fatalf("LoadConfig returned unexpected error: %v", err)
	}

	// デフォルトプロファイルを取得して検証
	profileConfig := GetProfileConfig(config, "default")

	// フィールドが空になっていることを確認
	if profileConfig.RedashURL != "" {
		t.Errorf("Expected empty RedashURL, got: %s", profileConfig.RedashURL)
	}
	if profileConfig.APIKey != "" {
		t.Errorf("Expected empty APIKey, got: %s", profileConfig.APIKey)
	}

	// 検証してエラーが返されることを確認
	err = ValidateProfileConfig(profileConfig)
	if err == nil {
		t.Error("ValidateProfileConfig should return error when values are empty")
	}

	// エラーメッセージに期待する情報が含まれていることを確認
	if !strings.Contains(err.Error(), "redash_url") || !strings.Contains(err.Error(), "api_key") {
		t.Errorf("Error message should mention missing fields, got: %v", err)
	}
}
