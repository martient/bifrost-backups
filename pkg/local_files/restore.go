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
			if err := os.MkdirAll(targetDir, 0750); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Handle file marker
		if strings.HasPrefix(line, "FILE:") {
			// Close previous file if any
			if currentFile != nil {
				if err := currentFile.Close(); err != nil {
					return fmt.Errorf("failed to close file: %w", err)
				}
				currentFile = nil
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

			// Validate the target path
			cleanPath := filepath.Clean(targetFile)
			if !filepath.IsAbs(cleanPath) {
				return fmt.Errorf("target path must be absolute: %s", targetFile)
			}

			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(cleanPath), 0750); err != nil {
				return fmt.Errorf("failed to create parent directories: %w", err)
			}

			// Create the file
			var err error
			currentFile, err = os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			continue
		}

		// Handle end marker
		if line == "---END---" {
			if currentFile != nil {
				if err := currentFile.Close(); err != nil {
					return fmt.Errorf("failed to close file: %w", err)
				}
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
