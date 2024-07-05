package localstorage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/martient/golang-utils/utils"
)

func getBackupFiles(folderPath string) ([]string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var backupFiles []string
	for _, file := range files {
		if !file.IsDir() {
			backupFiles = append(backupFiles, file.Name())
		}
	}

	return backupFiles, nil
}

func deleteOldBackups(folderPath string, retentionDays int) error {
	backupFiles, err := getBackupFiles(folderPath)
	if err != nil {
		return err
	}

	sort.Strings(backupFiles)

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	for _, fileName := range backupFiles {
		backupTime, err := time.Parse("2006-01-02T15:04:05Z", fileName)
		if err != nil {
			// Skip files that don't match the expected date format
			utils.LogError("Failed to parse backup file %s: %v", fileName, err)
			continue
		}

		if backupTime.Before(cutoffTime) {
			filePath := filepath.Join(folderPath, fileName)
			err = os.Remove(filePath)
			if err != nil {
				return fmt.Errorf("failed to delete backup file %s: %v", filePath, err)
			}
		}
	}

	return nil
}

func ExecuteRetentionPolicy(storage LocalStorageRequirements, retention_days int) error {
	if storage == (LocalStorageRequirements{}) {
		return fmt.Errorf("storage can't be empty")
	}

	return deleteOldBackups(storage.FolderPath, retention_days)
}
