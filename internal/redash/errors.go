package redash

import (
	"fmt"
	"strings"
)

// IsHTMLResponseError checks if the error indicates an HTML response instead of JSON
func IsHTMLResponseError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "HTML instead of JSON")
}

// PrintCommonErrorSuggestions prints helpful suggestions for common Redash API errors
func PrintCommonErrorSuggestions(err error) {
	if IsHTMLResponseError(err) {
		fmt.Printf("エラー: %v\n\n", err)
		fmt.Println("考えられる解決策:")
		fmt.Println("1. APIキーが正しいか確認してください。")
		fmt.Println("2. RedashのURLが正しいか確認してください（末尾にスラッシュがあるか、APIパスが含まれていないか）。")
		fmt.Println("3. RedashのURLにアクセス可能か確認してください。")
		fmt.Println("4. 設定ファイル（~/.redrip/config.conf）の内容を確認してください。")
	}
}
