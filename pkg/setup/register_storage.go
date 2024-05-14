package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/setup/interactives"
	"github.com/martient/golang-utils/utils"
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
		utils.LogError("Could not start program: %s\n", "Register datbase", err)
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
		utils.LogError("Path can't be empty", "Register local storage", nil)
		return nil, fmt.Errorf("path can't be empty")
	}
	requirements.FolderPath = path
	return requirements, nil
}

func checkIfAlreadyExist(name string, Storages []Storage) bool {
	if len(Storages) >= 1 {
		for i := 0; i < len(Storages); i++ {
			if name == Storages[i].Name {
				return true
			}
		}
	}
	return false
}

func RegisterStorage(storageType StorageType, name string, storage interface{}) error {
	if storage == nil {
		return fmt.Errorf("storage is null")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory")
	}
	configFilePath := filepath.Join(homeDir, ".config", "bifrost_backups.json")

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file")
	}
	defer file.Close()

	config := Config{}

	json.NewDecoder(file).Decode(&config)

	exist := checkIfAlreadyExist(name, config.Storages)
	if exist {
		return fmt.Errorf("storage name already used")
	}

	newStorage := &Storage{
		Type: storageType,
		Name: name,
	}

	switch db := storage.(type) {
	case *localstorage.LocalStorageRequirements:
		newStorage.LocalStorage = *db
	default:
		return fmt.Errorf("unsupported storage type")
	}

	config.Storages = append(config.Storages, *newStorage)

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("error truncate config file")
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error seek config file")
	}
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding JSON")
	}
	return nil
}
