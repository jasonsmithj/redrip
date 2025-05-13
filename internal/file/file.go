// Package file provides utilities for file operations
package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectory ensures that a directory exists, creating it if necessary
func EnsureDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// WriteFile writes data to a file, creating parent directories if necessary
func WriteFile(path string, data []byte, perm os.FileMode) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := EnsureDirectory(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write file
	return os.WriteFile(path, data, perm)
}

// Exists checks if a file or directory exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if a path is a regular file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
