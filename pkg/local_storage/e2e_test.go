package localstorage

import (
	"bytes"
	"os"
	"testing"
)

func TestStoreAndFetchBackup(t *testing.T) {
	t.Run("should store and fetch a backup file with compression", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "bifrost-backups")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Errorf("Failed to remove temp directory: %v", err)
			}
		}()

		storage := LocalStorageRequirements{FolderPath: tempDir}
		originalData := "test backup data"
		buffer := bytes.NewBufferString(originalData)

		// Store the backup
		err = StoreBackup(storage, buffer, true)
		if err != nil {
			t.Fatalf("Failed to store backup: %v", err)
		}

		// Fetch the backup
		fetchedBuffer, err := PullBackup(storage, "", true)
		if err != nil {
			t.Fatalf("Failed to fetch backup: %v", err)
		}

		// Compare the original and fetched data
		if fetchedBuffer.String() != originalData {
			t.Errorf("Expected fetched data to be '%s', got '%s'", originalData, fetchedBuffer.String())
		}
	})

	t.Run("should store and fetch a backup file without compression", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "bifrost-backups")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Errorf("Failed to remove temp directory: %v", err)
			}
		}()

		storage := LocalStorageRequirements{FolderPath: tempDir}
		originalData := "test backup data without compression"
		buffer := bytes.NewBufferString(originalData)

		// Store the backup
		err = StoreBackup(storage, buffer, false)
		if err != nil {
			t.Fatalf("Failed to store backup: %v", err)
		}

		// Fetch the backup
		fetchedBuffer, err := PullBackup(storage, "", false)
		if err != nil {
			t.Fatalf("Failed to fetch backup: %v", err)
		}

		// Compare the original and fetched data
		if fetchedBuffer.String() != originalData {
			t.Errorf("Expected fetched data to be '%s', got '%s'", originalData, fetchedBuffer.String())
		}
	})
}
