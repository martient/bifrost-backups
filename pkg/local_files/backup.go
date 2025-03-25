package localfiles

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/martient/golang-utils/utils"
)

func RunBackup(config LocalFilesRequirements) (*bytes.Buffer, error) {
	if err := validateRequirements(config); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	sourcePath := config.Path

	// Get source info
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get source info: %w", err)
	}

	if sourceInfo.IsDir() {
		err = backupDirectory(sourcePath, &buffer, config)
	} else {
		// Write file marker for single file
		if _, err := buffer.WriteString(fmt.Sprintf("FILE:%s\n", sourcePath)); err != nil {
			return nil, err
		}
		err = backupFile(sourcePath, &buffer)
	}

	if err != nil {
		utils.LogError("Failed to backup '%s'", "LOCAL_FILES", err)
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	return &buffer, nil
}

func backupDirectory(sourcePath string, buffer *bytes.Buffer, config LocalFilesRequirements) error {
	// Write directory marker
	if _, err := fmt.Fprintf(buffer, "DIR:%s\n", sourcePath); err != nil {
		return err
	}

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == sourcePath {
			return nil
		}

		// Skip if path matches any exclude pattern
		for _, pattern := range config.ExcludePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return nil
			}
		}

		if info.IsDir() {
			// Write directory marker
			_, err := fmt.Fprintf(buffer, "DIR:%s\n", path)
			return err
		}

		// Write file marker and content
		if _, err := fmt.Fprintf(buffer, "FILE:%s\n", path); err != nil {
			return err
		}

		return backupFile(path, buffer)
	})
}

func backupFile(sourcePath string, buffer *bytes.Buffer) error {
	// Validate the path
	cleanPath := filepath.Clean(sourcePath)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("source path must be absolute: %s", sourcePath)
	}

	sourceFile, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			log.Printf("failed to close source file: %v", err)
		}
	}()

	_, err = io.Copy(buffer, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Write file separator
	_, err = buffer.WriteString("\n---END---\n")
	return err
}

func validateRequirements(config LocalFilesRequirements) error {
	if config.Path == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if !filepath.IsAbs(config.Path) {
		return fmt.Errorf("source path must be absolute")
	}
	return nil
}
