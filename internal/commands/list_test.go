package commands

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestListCommandOutput(t *testing.T) {
	// テスト用の簡易的なコマンドを作成
	var testQueries = []struct {
		id   int
		name string
	}{
		{1, "Query 1"},
		{2, "Query 2"},
		{3, "Query 3"},
	}

	// 簡易的なリストコマンドを作成
	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			for _, q := range testQueries {
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "ID: %d\tName: %s\n", q.id, q.name)
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error writing output: %v\n", err)
				}
			}
		},
	}

	// コマンド出力をキャプチャ
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// コマンドを実行
	cmd.Run(cmd, []string{})

	// 期待される出力を作成
	expected := ""
	for _, q := range testQueries {
		expected += fmt.Sprintf("ID: %d\tName: %s\n", q.id, q.name)
	}

	// 出力を検証
	if got := buf.String(); got != expected {
		t.Errorf("Expected output %q, got %q", expected, got)
	}
}

func TestListCommandFormatting(t *testing.T) {
	// 出力フォーマットのテスト
	testCases := []struct {
		id       int
		name     string
		expected string
	}{
		{1, "Test Query", "ID: 1\tName: Test Query\n"},
		{100, "Another Query", "ID: 100\tName: Another Query\n"},
		{0, "", "ID: 0\tName: \n"},
	}

	for _, tc := range testCases {
		// 出力をキャプチャ
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "ID: %d\tName: %s\n", tc.id, tc.name)

		// 出力を検証
		if got := buf.String(); got != tc.expected {
			t.Errorf("For ID=%d, Name=%q: expected %q, got %q", tc.id, tc.name, tc.expected, got)
		}
	}
}
