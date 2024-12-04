package setup

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/martient/bifrost-backup/pkg/crypto"
	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/s3"
	"github.com/martient/bifrost-backup/pkg/setup/interactives"
	"github.com/martient/golang-utils/utils"
	"github.com/pkg/errors"
)

func InteractiveRegisterStorage() {
	if _, err := tea.NewProgram(interactives.LocalStorageInitialModel()).Run(); err != nil {
		utils.LogError("Could not start program: %s\n", "Register datbase", err)
		os.Exit(1)
	}
}

func RegisterLocalStorage(path string) (*localstorage.LocalStorageRequirements, error) {
	requirements := &localstorage.LocalStorageRequirements{}
	if len(path) <= 0 {
		utils.LogError("Path can't be empty", "Register local storage", nil)
		return nil, fmt.Errorf("path can't be empty")
	}
	requirements.FolderPath = path
	return requirements, nil
}

func RegisterS3Storage(bucket_name string, access_key_id string, access_key_secret string, endpoint string, region string) (*s3.S3Requirements, error) {
	requirements := &s3.S3Requirements{}
	if len(bucket_name) <= 0 || len(access_key_id) <= 0 || len(access_key_secret) <= 0 || len(region) <= 0 {
		utils.LogError("bucket_name, access_key_id, access_key_secret, region can't be empty", "Register s3 database", nil)
		return nil, fmt.Errorf("bucket_name, account_id, access_key_id, access_key_secret, endpoint, region can't be empty")
	}
	requirements.BucketName = bucket_name
	requirements.AccessKeyId = access_key_id
	requirements.AccessKeySecret = access_key_secret
	requirements.Endpoint = endpoint
	requirements.Region = region
	return requirements, nil
}

func RegisterStorage(storageType StorageType, name string, retention int, cipher_key string, compression bool, storage interface{}) error {
	if storage == nil {
		return fmt.Errorf("storage is null")
	}
	configMutex.Lock()
	defer configMutex.Unlock()

	// Read current config
	currentConfig, err := readConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Generate cipher key if not provided
	if cipher_key == "" {
		cipher_key, err = crypto.GenerateCipherKey(32)
		if err != nil {
			return errors.Wrap(err, "could not generate cipher key")
		}
	}

	newStorage := &Storage{
		Type:          storageType,
		Name:          name,
		RetentionDays: retention,
		CipherKey:     cipher_key,
		Compression:   compression,
	}

	switch req := storage.(type) {
	case *localstorage.LocalStorageRequirements:
		newStorage.LocalStorage = *req
	case *s3.S3Requirements:
		newStorage.S3 = *req
	default:
		return fmt.Errorf("unsupported storage type")
	}

	// Find and update existing storage, or append new one
	found := false
	for i := range currentConfig.Storages {
		if currentConfig.Storages[i].Name == name {
			currentConfig.Storages[i] = *newStorage
			found = true
			utils.LogInfo("Storage %s configuration updated", "REGISTER STORAGE", name)
			break
		}
	}

	if !found {
		currentConfig.Storages = append(currentConfig.Storages, *newStorage)
		utils.LogInfo("Storage %s registered", "REGISTER STORAGE", name)
	}

	// Write the updated config back using the secure system
	if err := writeConfig(currentConfig); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
