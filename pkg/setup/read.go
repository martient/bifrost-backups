package setup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func readConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	configFilePath := filepath.Join(homeDir, ".config", "bifrost_backups.yaml")

	file, err := os.OpenFile(configFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return Config{}, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	config := Config{}
	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %v", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func GetDatabaseConfigName() ([]string, error) {
	config, err := readConfig()
	var names []string

	if err != nil {
		return nil, err
	}
	for i := 0; i < len(config.Databases); i++ {
		if config.Databases[i].Name != "" {
			names = append(names, config.Databases[i].Name)
		}
	}
	return names, nil
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
			if config.Storages[i].RetentionDays == 0 {
				config.Storages[i].RetentionDays = 21 // Default retention period is 21 days
			}
			return config.Storages[i], nil
		}
	}
	return Storage{}, fmt.Errorf("storage with name %q not found", name)
}
