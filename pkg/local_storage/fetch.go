package localstorage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/martient/golang-utils/utils"
)

func getBackupPath(path string) (string, error) {
	if len(path) <= 0 {
		return "", fmt.Errorf("the backup folder path can't be empty")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no backups found in folder %s", path)
	}

	// Sort entries by modification time
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no backup files found in %s", path)
	}

	// Sort files by name (which includes timestamp)
	sort.Strings(files)
	utils.LogDebug("Latest file: %s", "LOCAL-STORAGE", files[len(files)-1])
	return files[len(files)-1], nil
}

func PullBackup(storage LocalStorageRequirements, backup_name string, useCompression bool) (*bytes.Buffer, error) {
	if storage == (LocalStorageRequirements{}) {
		return nil, fmt.Errorf("storage can't be empty")
	}
	var err error

	latestBackupKey := backup_name
	if latestBackupKey == "" {
		latestBackupKey, err = getBackupPath(storage.FolderPath)
		if err != nil {
			return nil, err
		}
	}

	filePath := strings.TrimSuffix(filepath.Join(storage.FolderPath, latestBackupKey), "\n")

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0600) //#nosec
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var reader io.Reader = file

	if useCompression {
		zReader, err := zstd.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create new reader for file %s from folder %s: %v", latestBackupKey, storage.FolderPath, err)
		}
		defer zReader.Close()
		reader = zReader
	}
	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s from folder %s: %v", latestBackupKey, storage.FolderPath, err)
	}

	return buf, nil
}
