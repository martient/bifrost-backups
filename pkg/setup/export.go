package setup

import (
	"fmt"
	"os"
	"path/filepath"

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
	// Validate and clean the path
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	// Create the directory with secure permissions
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file with secure permissions
	file, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error closing config file: %w", cerr)
		}
	}()

	// Write config to file
	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return err
}
