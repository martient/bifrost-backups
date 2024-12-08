package localfiles

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/martient/golang-utils/utils"
)

func RunRestore(config LocalFilesRequirements, backupData *bytes.Buffer) error {
	if err := validateRequirements(config); err != nil {
		return err
	}

	scanner := bufio.NewScanner(backupData)
	var currentFile *os.File
	var basePath string

	for scanner.Scan() {
		line := scanner.Text()

		// Handle directory marker
		if strings.HasPrefix(line, "DIR:") {
			dirPath := strings.TrimPrefix(line, "DIR:")
			if basePath == "" {
				basePath = dirPath
			}
			targetDir := config.Path
			if basePath != dirPath {
				rel, err := filepath.Rel(basePath, dirPath)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}
				targetDir = filepath.Join(config.Path, rel)
			}
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Handle file marker
		if strings.HasPrefix(line, "FILE:") {
			// Close previous file if any
			if currentFile != nil {
				currentFile.Close()
			}

			filePath := strings.TrimPrefix(line, "FILE:")
			var targetFile string
			if basePath == "" {
				// Single file backup
				targetFile = config.Path
			} else {
				// Part of directory backup
				rel, err := filepath.Rel(basePath, filePath)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}
				targetFile = filepath.Join(config.Path, rel)
			}

			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
				return fmt.Errorf("failed to create parent directories: %w", err)
			}

			// Create the file
			var err error
			currentFile, err = os.Create(targetFile)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			continue
		}

		// Handle end marker
		if line == "---END---" {
			if currentFile != nil {
				currentFile.Close()
				currentFile = nil
			}
			continue
		}

		// Write content to current file
		if currentFile != nil {
			if _, err := currentFile.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		utils.LogError("Failed to restore from backup", "LOCAL_FILES", err)
		return fmt.Errorf("restore failed: %w", err)
	}

	return nil
}
