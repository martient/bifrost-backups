package setup

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ReadConfigUnciphered reads the config file and returns an unciphered copy
func ReadConfigUnciphered() (Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Read current config
	config, err := readConfig()
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	// Create a copy of the config to avoid modifying the original
	configCopy := config

	// Initialize secure manager with encryption disabled
	secureManager, err := NewSecureManager(true)
	if err != nil {
		return Config{}, fmt.Errorf("failed to initialize secure manager: %w", err)
	}

	// Decrypt all sensitive fields
	err = secureManager.DecryptConfig(&configCopy)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decrypt config: %w", err)
	}

	return configCopy, nil
}

// WriteConfigUnciphered writes the config to the specified path without encryption
func WriteConfigUnciphered(path string, config Config) error {
	// Create the file with secure permissions
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error creating config file: %v", err)
	}
	defer file.Close()

	// Write config to file
	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
