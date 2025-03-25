package setup

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	configFilePath string
	configMutex    = &sync.Mutex{}
	noEncryption   bool
	version        = "1.0"
)

func init() {
	// First check if BIFROST_CONFIG environment variable is set
	if envPath := os.Getenv("BIFROST_CONFIG"); envPath != "" {
		configFilePath = envPath
		return
	}

	// Fall back to default path in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}
	configFilePath = filepath.Join(homeDir, ".config", "bifrost_backups.yaml")
}

func readConfig() (Config, error) {
	file, err := os.OpenFile(configFilePath, os.O_RDONLY, 0600) //#nosec
	if err != nil {
		return Config{}, fmt.Errorf("error opening config file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close config file: %v", err)
		}
	}()

	config := Config{}
	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %v", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	// Initialize secure manager and decrypt sensitive fields if encryption is enabled
	secureManager, err := NewSecureManager(noEncryption)
	if err != nil {
		return Config{}, fmt.Errorf("failed to initialize secure manager: %w", err)
	}

	err = secureManager.DecryptConfig(&config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decrypt config: %w", err)
	}

	return config, nil
}

func writeConfig(config Config) error {
	// Initialize secure manager and encrypt sensitive fields if encryption is enabled
	secureManager, err := NewSecureManager(noEncryption)
	if err != nil {
		return fmt.Errorf("failed to initialize secure manager: %w", err)
	}

	// Create a copy of the config to avoid modifying the original
	configCopy := config
	err = secureManager.SecureConfig(&configCopy)
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the config to a temporary file first
	tempFile, err := os.CreateTemp(configDir, "bifrost_backups.*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	tempFilePath := tempFile.Name()

	// Ensure temp file has secure permissions
	if err := os.Chmod(tempFilePath, 0600); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to set temp file permissions and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to set temp file permissions: %w", err)
	}

	// Write config to temp file
	encoder := yaml.NewEncoder(tempFile)
	if err := encoder.Encode(configCopy); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to encode config and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to close temp file and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomically replace the old config file
	if err := os.Rename(tempFilePath, configFilePath); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to save config file and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// UpdateConfig updates the configuration file with the provided config
func UpdateConfig(config Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Initialize secure manager and encrypt sensitive fields if encryption is enabled
	secureManager, err := NewSecureManager(noEncryption)
	if err != nil {
		return fmt.Errorf("failed to initialize secure manager: %w", err)
	}

	// Create a copy of the config to avoid modifying the original
	configCopy := config
	err = secureManager.SecureConfig(&configCopy)
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the config to a temporary file first
	tempFile, err := os.CreateTemp(configDir, "bifrost_backups.*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	tempFilePath := tempFile.Name()

	// Ensure temp file has secure permissions
	if err := os.Chmod(tempFilePath, 0600); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to set temp file permissions and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to set temp file permissions: %w", err)
	}

	// Write config to temp file
	encoder := yaml.NewEncoder(tempFile)
	if err := encoder.Encode(configCopy); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to encode config and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to close temp file and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomically replace the old config file
	if err := os.Rename(tempFilePath, configFilePath); err != nil {
		if removeErr := os.Remove(tempFilePath); removeErr != nil {
			return fmt.Errorf("failed to save config file and cleanup failed: %v, %v", err, removeErr)
		}
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
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

func GetStorageConfigName() ([]string, error) {
	config, err := readConfig()
	var names []string

	if err != nil {
		return nil, err
	}
	for i := 0; i < len(config.Storages); i++ {
		if config.Storages[i].Name != "" {
			names = append(names, config.Storages[i].Name)
		}
	}
	return names, nil
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
