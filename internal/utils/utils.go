package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func ValidateFilePath(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	dir := filepath.Dir(absPath)

	fileInfo, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("directory check failed for %s: %w", dir, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("path %s is not a directory", dir)
	}

	tmp, err := os.CreateTemp(dir, "perm_check_")
	if err != nil {
		return fmt.Errorf("directory %s is not writable: %w", dir, err)
	}
	tmp.Close()
	os.Remove(tmp.Name())

	return nil
}
