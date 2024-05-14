package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/golang-utils/utils"
)

func GenerateDefaultConfig(current_version string) {
	config := &Config{}

	homeDir, err := os.UserHomeDir()
	config.Version = current_version
	config.Storages = append(config.Storages, Storage{
		Type: LocalStorage,
		Name: "default",
		LocalStorage: localstorage.LocalStorageRequirements{
			FolderPath: filepath.Join(homeDir, ".bifrost-backups"),
		},
	})

	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}
	configFilePath := filepath.Join(homeDir, ".config", "bifrost_backups.json")

	// Create the config file if it doesn't exist
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
			utils.LogError("Error creating config directory:", "SETUP", err)
			return
		}
		file, err := os.Create(configFilePath)
		if err != nil {
			utils.LogError("Error creating config file:", "SETUP", err)
			return
		}
		defer file.Close()
		utils.LogInfo("Config file created at:", "SETUP", configFilePath)
	} else {
		utils.LogInfo("config file already exist", "SETUP")
		return
	}

	file, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		utils.LogError("Error opening config file", "SETUP", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		utils.LogError("Error encoding JSON", "SETUP", err)
		return
	}
}
