package localstorage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/martient/golang-utils/utils"
)

func getBackupPath(path string) (string, error) {
	if len(path) <= 0 {
		return "", fmt.Errorf("the backup folder path can't be empty")
	}
	var out bytes.Buffer

	cmd := exec.Command("sh", "-c", fmt.Sprintf("ls -Art %s | tail -n 1", path))

	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	utils.LogDebug("Latest file: %s", "LOCAL-STORAGE", out.String())
	if out.String() == "" {
		return "", fmt.Errorf("no backups found in folder %s", path)
	}

	return strings.TrimSuffix(out.String(), "\n"), nil
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

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
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
