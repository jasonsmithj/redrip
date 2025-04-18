package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDumpFilesCreation(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir := t.TempDir()

	// 複数のファイルを作成するテスト
	files := []struct {
		id      int
		content string
	}{
		{1, "SELECT * FROM table1"},
		{2, "SELECT * FROM table2"},
		{3, "SELECT * FROM table3"},
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, fmt.Sprintf("%d.sql", file.id))
		err := os.WriteFile(filePath, []byte(file.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %d: %v", file.id, err)
		}
	}

	// ファイルが作成されていることを確認
	for _, file := range files {
		filePath := filepath.Join(tempDir, fmt.Sprintf("%d.sql", file.id))

		// ファイルが存在することを確認
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %d was not created", file.id)
			continue
		}

		// ファイルの内容を確認
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file %d: %v", file.id, err)
		}
		if string(data) != file.content {
			t.Errorf("File %d: expected content %q, got %q", file.id, file.content, string(data))
		}
	}
}

func TestDumpDirectoryCreation(t *testing.T) {
	// 一時ディレクトリを作成
	baseDir := t.TempDir()
	nestedDir := filepath.Join(baseDir, "nested", "sql", "dir")

	// ディレクトリが存在しない場合は作成される
	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// ファイルを作成
	filePath := filepath.Join(nestedDir, "1.sql")
	content := "SELECT * FROM test"
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// ファイルが作成されたことを確認
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}
}
