package localstorage

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
)

func TestStoreBackup(t *testing.T) {
	t.Run("should create a backup file with the correct name and content", func(t *testing.T) {
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
		buffer := bytes.NewBufferString("test data")
		expectedFilename := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%03dZ",
			time.Now().UTC().Year(),
			time.Now().UTC().Month(),
			time.Now().UTC().Day(),
			time.Now().UTC().Hour(),
			time.Now().UTC().Minute(),
			time.Now().UTC().Second(),
		)
		expectedFilePath := filepath.Join(tempDir, expectedFilename)

		err = StoreBackup(storage, buffer, false)
		if err != nil {
			t.Fatal(err)
		}

		_, err = os.Stat(expectedFilePath)
		if err != nil {
			t.Errorf("Expected backup file to exist, got error: %v", err)
		}

		data, err := os.ReadFile(expectedFilePath)
		if err != nil {
			t.Errorf("Error reading backup file: %v", err)
		}
		if string(data) != "test data" {
			t.Errorf("Expected backup file content to be 'test data', got '%s'", string(data))
		}
	})

	t.Run("should create a compressed backup file with the correct name and content", func(t *testing.T) {
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
		buffer := bytes.NewBufferString("test data")
		expectedFilename := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%03dZ",
			time.Now().UTC().Year(),
			time.Now().UTC().Month(),
			time.Now().UTC().Day(),
			time.Now().UTC().Hour(),
			time.Now().UTC().Minute(),
			time.Now().UTC().Second(),
		)
		expectedFilePath := filepath.Join(tempDir, expectedFilename)

		err = StoreBackup(storage, buffer, true)
		if err != nil {
			t.Fatal(err)
		}

		_, err = os.Stat(expectedFilePath)
		if err != nil {
			t.Errorf("Expected backup file to exist, got error: %v", err)
		}

		data, err := os.ReadFile(expectedFilePath)
		if err != nil {
			t.Errorf("Error reading backup file: %v", err)
		}
		// Decompress the data
		decoder, err := zstd.NewReader(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer decoder.Close()

		// Decompress the data
		decompressed, err := decoder.DecodeAll(data, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err != nil {
			t.Errorf("Error decompressing backup file: %v", err)
		}
		if string(decompressed) != "test data" {
			t.Errorf("Expected backup file content to be 'test data', got '%s'", string(decompressed))
		}
	})

	t.Run("should return an error if the buffer is empty", func(t *testing.T) {
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
		var buffer *bytes.Buffer

		err = StoreBackup(storage, buffer, false)

		if err == nil {
			t.Error("Expected error for empty buffer, got nil")
		}
		if err != nil && err.Error() != "buffer can't be empty" {
			t.Errorf("Expected error message 'buffer can't be empty', got '%s'", err.Error())
		}
	})

	t.Run("should return an error if the storage is empty", func(t *testing.T) {
		buffer := bytes.NewBufferString("test data")
		var storage LocalStorageRequirements

		err := StoreBackup(storage, buffer, false)

		if err == nil {
			t.Error("Expected error for empty storage, got nil")
		}
		if err != nil && err.Error() != "storage can't be empty" {
			t.Errorf("Expected error message 'storage can't be empty', got '%s'", err.Error())
		}
	})

	t.Run("should not return an error if the folder path does not exist", func(t *testing.T) {
		storage := LocalStorageRequirements{FolderPath: "/tmp/non-existent-folder"}
		buffer := bytes.NewBufferString("test data")

		err := StoreBackup(storage, buffer, false)

		if err != nil && !os.IsNotExist(err) {
			t.Errorf("Expected error to be 'os.IsNotExist', got '%s'", err.Error())
		}
	})
}
