package localstorage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DataDog/zstd"
	"github.com/martient/golang-utils/utils"
)

func StoreBackup(storage LocalStorageRequirements, buffer *bytes.Buffer, useCompression bool) error {
	if buffer == nil {
		return fmt.Errorf("buffer can't be empty")
	} else if storage == (LocalStorageRequirements{}) {
		return fmt.Errorf("storage can't be empty")
	}

	if _, err := os.Stat(storage.FolderPath); os.IsNotExist(err) {
		err = os.MkdirAll(storage.FolderPath, 0755)
		if err != nil {
			utils.LogError("Folder creation went wrong", "Local storage", err)
			return err
		}
	}
	currentTime := time.Now().UTC()

	backupPath := filepath.Join(storage.FolderPath, fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%03dZ",
		currentTime.Year(),
		currentTime.Month(),
		currentTime.Day(),
		currentTime.Hour(),
		currentTime.Minute(),
		currentTime.Second()))
	file, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var dataToWrite []byte

	if useCompression {
		compressed, err := zstd.Compress(nil, buffer.Bytes())
		if err != nil {
			utils.LogError("Compression failed", "Local storage", err)
			return err
		}
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
