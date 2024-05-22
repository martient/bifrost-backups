package setup

import (
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/s3"
	"github.com/martient/bifrost-backup/pkg/setup/interactives"
	"github.com/martient/golang-utils/utils"
	"github.com/pkg/errors"
)

func InteractiveRegisterStorage() {
	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	fmt.Println("Error getting home directory:", err)
	// 	return
	// }
	// configFilePath := filepath.Join(homeDir, ".config", "bifrost_backups.json")

	// file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_TRUNC, 0644)
	// if err != nil {
	// 	fmt.Println("Error opening config file:", err)
	// 	return
	// }
	// defer file.Close()

	if _, err := tea.NewProgram(interactives.LocalStorageInitialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}

	// encoder := json.NewEncoder(file)
	// if err := encoder.Encode(config); err != nil {
	// 	fmt.Println("Error encoding JSON:", err)
	// 	return
	// }
}

func RegisterLocalStorage(path string) (*localstorage.LocalStorageRequirements, error) {
	requirements := &localstorage.LocalStorageRequirements{}
	if len(path) <= 0 {
		utils.LogError("Path can't be empty", "Register postgresql database", nil)
		return nil, fmt.Errorf("username can't be empty")
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

func RegisterStorage(storageType StorageType, name string, storage interface{}) error {
	if storage == nil {
		return fmt.Errorf("storage is null")
	}
	configMutex.Lock()
	defer configMutex.Unlock()

	newStorage := &Storage{
		Type: storageType,
		Name: name,
	}

	switch db := storage.(type) {
	case *localstorage.LocalStorageRequirements:
		newStorage.LocalStorage = *db
	case *s3.S3Requirements:
		newStorage.S3 = *db
	default:
		return fmt.Errorf("unsupported storage type")
	}

	alreadyExist := false
	for i, ite := range config.Storages {
		if ite.Name == newStorage.Name {
			config.Storages[i] = *newStorage
			alreadyExist = true
			utils.LogInfo("Storage %s already register, it as been updated", "REGISTER STORAGE", ite.Name)
			break
		}
	}
	if !alreadyExist {
		config.Storages = append(config.Storages, *newStorage)
	}

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "error opening config file")
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		return errors.Wrap(err, "error encoding config file")
	}
	return nil
}
