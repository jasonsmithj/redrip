package redash

import (
	"os"
	"path/filepath"
	"testing"
)

// GetSQLDirWithConfigPath は GetSQLDir の代替関数で、設定ファイルのパスを指定可能
func GetSQLDirWithConfigPath(configPath string) (string, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return "", err
	}

	// SQLDirが設定されていない場合はカレントディレクトリを使用
	if config.SQLDir == "" {
		return ".", nil
	}

	// ディレクトリが存在するか確認
	if _, err := os.Stat(config.SQLDir); os.IsNotExist(err) {
		return ".", nil
	}

	return config.SQLDir, nil
}

func TestGetSQLDirWithConfig(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	sqlDir := filepath.Join(tempDir, "sql")

	// SQLディレクトリを作成
	err := os.MkdirAll(sqlDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create SQL directory: %v", err)
	}

	// テスト用の設定ファイルを作成
	configPath := filepath.Join(tempDir, "config.conf")
	configContent := `redash_url = https://test-redash.com/api
api_key = test-api-key
sql_dir = ` + sqlDir

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// テスト実行
	dir, err := GetSQLDirWithConfigPath(configPath)
	if err != nil {
		t.Fatalf("GetSQLDirWithConfigPath returned error: %v", err)
	}

	// 結果の検証
	if dir != sqlDir {
		t.Errorf("Expected SQL dir = %s, got %s", sqlDir, dir)
	}
}

func TestGetSQLDirWithoutSQLDirConfig(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// SQLディレクトリが設定されていない設定ファイルを作成
	configPath := filepath.Join(tempDir, "config.conf")
	configContent := `redash_url = https://test-redash.com/api
api_key = test-api-key`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// テスト実行
	dir, err := GetSQLDirWithConfigPath(configPath)
	if err != nil {
		t.Fatalf("GetSQLDirWithConfigPath returned error: %v", err)
	}

	// 結果の検証（カレントディレクトリになるはず）
	if dir != "." {
		t.Errorf("Expected SQL dir = %s, got %s", ".", dir)
	}
}

func TestGetSQLDirWithNonExistentDir(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// 存在しないディレクトリを指定した設定ファイルを作成
	nonExistentDir := filepath.Join(tempDir, "non-existent-dir")
	configPath := filepath.Join(tempDir, "config.conf")
	configContent := `redash_url = https://test-redash.com/api
api_key = test-api-key
sql_dir = ` + nonExistentDir

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// テスト実行
	dir, err := GetSQLDirWithConfigPath(configPath)
	if err != nil {
		t.Fatalf("GetSQLDirWithConfigPath returned error: %v", err)
	}

	// 結果の検証（ディレクトリが存在しないのでカレントディレクトリになるはず）
	if dir != "." {
		t.Errorf("Expected SQL dir = %s, got %s", ".", dir)
	}
}
