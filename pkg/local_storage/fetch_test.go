package localstorage

import (
	"fmt"

	"github.com/DataDog/zstd"

	"os"
	"path/filepath"
	"testing"
)

func TestGetBackupPath(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		_, err := getBackupPath("")
		if err == nil {
			t.Error("Expected error for empty path, got nil")
		}
		if err != nil && err.Error() != "the backup folder path can't be empty" {
			t.Errorf("Expected error message 'the backup folder path can't be empty', got '%s'", err.Error())
		}
	})
	t.Run("no backups found", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "bifrost-backups")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		_, err = getBackupPath(tempDir)
		if err == nil {
			t.Error("Expected error for no backups found, got nil")
		}
		if err != nil && err.Error() != fmt.Sprintf("no backups found in folder %s", tempDir) {
			t.Errorf("Expected error message '%s', got '%s'", fmt.Sprintf("no backups found in folder %s", tempDir), err.Error())
		}
	})

	t.Run("valid path", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "bifrost-backups")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		// Create a test backup file
		backupFile := filepath.Join(tempDir, "test_backup.json")
		err = os.WriteFile(backupFile, []byte("test backup data"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		backupPath, err := getBackupPath(tempDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if backupPath != "test_backup.json" {
			t.Errorf("Expected backup path 'test_backup.json', got '%s'", backupPath)
		}
	})
}

func TestPullBackup(t *testing.T) {
	t.Run("empty storage", func(t *testing.T) {
		_, err := PullBackup(LocalStorageRequirements{}, "", false)
		if err == nil {
			t.Error("Expected error for empty storage, got nil")
		}
		if err != nil && err.Error() != "storage can't be empty" {
			t.Errorf("Expected error message 'storage can't be empty', got '%s'", err.Error())
		}
	})

	t.Run("valid storage, no backup name", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "bifrost-backups")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		// Create a test backup file
		backupFile := filepath.Join(tempDir, "test_backup.json")
		err = os.WriteFile(backupFile, []byte("test backup data"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		storage := LocalStorageRequirements{FolderPath: tempDir}
		buf, err := PullBackup(storage, "", false)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if buf.String() != "test backup data" {
			t.Errorf("Expected backup data 'test backup data', got '%s'", buf.String())
		}
	})

	t.Run("valid storage, with backup name", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "bifrost-backups")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		// Create a test backup file
		backupFile := filepath.Join(tempDir, "test_backup.json")
		err = os.WriteFile(backupFile, []byte("test backup data"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		storage := LocalStorageRequirements{FolderPath: tempDir}
		buf, err := PullBackup(storage, "test_backup.json", false)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if buf.String() != "test backup data" {
			t.Errorf("Expected backup data 'test backup data', got '%s'", buf.String())
		}
	})

	t.Run("valid storage, with backup name, using compression", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "bifrost-backups")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		// Create a test backup file
		backupFile := filepath.Join(tempDir, "test_backup.json.zst")
		data := []byte("test backup data")
		compressedData, err := zstd.CompressLevel(nil, data, 1)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(backupFile, compressedData, 0644)
		if err != nil {
			t.Fatal(err)
		}

		storage := LocalStorageRequirements{FolderPath: tempDir}
		buf, err := PullBackup(storage, "test_backup.json.zst", true)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if buf.String() != "test backup data" {
			t.Errorf("Expected backup data 'test backup data', got '%s'", buf.String())
		}
	})
}
