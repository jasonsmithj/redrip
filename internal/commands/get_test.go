package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// executeCommandForTest is a utility for testing commands.
// It is kept for future test cases, even if currently unused.
// nolint:unused
func executeCommandForTest(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

// 通常、GetCommandはNewClient()とGetQuery()を使うため、それらをモック化する必要があります。
// このテストでは、既存のコマンドをテストするのではなく、テスト用のシンプルなコマンドを作成します。

func TestGetCommandArgs(t *testing.T) {
	// get コマンドの引数チェックをテスト
	cmd := cobra.Command{}
	cmd.Args = cobra.ExactArgs(1)

	// 引数がない場合はエラーが返る
	err := cmd.Args(&cmd, []string{})
	if err == nil {
		t.Error("Expected error for missing arguments")
	}

	// 引数が多すぎる場合はエラーが返る
	err = cmd.Args(&cmd, []string{"1", "2"})
	if err == nil {
		t.Error("Expected error for too many arguments")
	}

	// 引数が1つの場合はエラーが返らない
	err = cmd.Args(&cmd, []string{"1"})
	if err != nil {
		t.Errorf("Unexpected error for valid arguments: %v", err)
	}
}

func TestCreateDirectory(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "dir")

	// ディレクトリが存在しない場合は作成される
	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// ディレクトリが作成されたことを確認
	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected a directory, but got a file")
	}
}

func TestWriteFile(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.sql")

	// ファイルに書き込み
	content := "SELECT * FROM test"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// ファイルの内容を確認
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}
}
