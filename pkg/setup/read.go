package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func readConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	configFilePath := filepath.Join(homeDir, ".config", "bifrost_backups.json")

	file, err := os.OpenFile(configFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return Config{}, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	config := Config{}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
func ReadDatabaseConfig(name string) (Database, error) {
	config, err := readConfig()

	if err != nil {
		return Database{}, err
	}
	for i := 0; i < len(config.Databases); i++ {
		if config.Databases[i].Name == name {
			return config.Databases[i], nil
		}
	}
	return Database{}, fmt.Errorf("database with name %q not found", name)
}

func ReadStorageConfig(name string) (Storage, error) {
	config, err := readConfig()

	if err != nil {
		return Storage{}, err
	}
	for i := 0; i < len(config.Storages); i++ {
		if config.Storages[i].Name == name {
			return config.Storages[i], nil
		}
	}
	return Storage{}, fmt.Errorf("storage with name %q not found", name)
}
