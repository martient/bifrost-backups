package localstorage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
	internalutils "github.com/martient/bifrost-backups/pkg/utils"
	"github.com/martient/golang-utils/utils"
)

func StoreBackup(storage LocalStorageRequirements, buffer *bytes.Buffer, useCompression bool) error {
	if buffer == nil {
		return fmt.Errorf("buffer can't be empty")
	} else if storage == (LocalStorageRequirements{}) {
		return fmt.Errorf("storage can't be empty")
	}

	if _, err := os.Stat(storage.FolderPath); os.IsNotExist(err) {
		err = os.MkdirAll(storage.FolderPath, 0750)
		if err != nil {
			utils.LogError("Folder creation went wrong", "Local storage", err)
			return err
		}
	}
	currentTime := time.Now().UTC()

	backupPath := filepath.Join(storage.FolderPath, internalutils.FormatBackupTimestamp(currentTime)) // Use the imported function

	// Validate backup path
	allowedPaths := []string{storage.FolderPath}
	if err := internalutils.ValidatePath(backupPath, allowedPaths); err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	file, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY, 0600) //#nosec
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			utils.LogError("Failed to close file", "Local storage", err)
		}
	}()

	var dataToWrite []byte

	if useCompression {
		encoder, err := zstd.NewWriter(nil)
		if err != nil {
			utils.LogError("Compression failed", "Local storage", err)
			return err
		}
		defer func() {
			if err := encoder.Close(); err != nil {
				utils.LogError("Failed to close encoder", "Local storage", err)
			}
		}()

		// Compress the input string
		compressed := encoder.EncodeAll([]byte(buffer.Bytes()), nil)
		dataToWrite = compressed
	} else {
		dataToWrite = buffer.Bytes()
	}

	_, err = file.Write(dataToWrite)
	if err != nil {
		return err
	}

	// _, err = buffer.WriteTo(file)
	// if err != nil {
	// 	return err
	// }

	return nil
}
