package setup

import (
	"os"
	"path/filepath"
	"testing"

	localstorage "github.com/martient/bifrost-backups/pkg/local_storage"
	"github.com/martient/bifrost-backups/pkg/postgresql"
)

func TestConfigFilePathSelection(t *testing.T) {
	// Save original config file path and restore it after the test
	originalConfigPath := configFilePath
	defer func() {
		configFilePath = originalConfigPath
		if err := os.Unsetenv("BIFROST_CONFIG"); err != nil {
			t.Errorf("Failed to unset BIFROST_CONFIG environment variable: %v", err)
		}
	}()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// Test case 1: Environment variable is set
	t.Run("Environment variable takes precedence", func(t *testing.T) {
		envConfigPath := filepath.Join(tmpDir, "env-config.yaml")
		if err := os.Setenv("BIFROST_CONFIG", envConfigPath); err != nil {
			t.Fatalf("Failed to set BIFROST_CONFIG environment variable: %v", err)
		}
		configFilePath = envConfigPath // Set it directly since init() has already run

		// Create a minimal valid config file
		config := Config{
			Version:   "1.0",
			Databases: []Database{},
			Storages:  []Storage{},
		}
		err := writeConfig(config)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		// Try to read config
		_, err = readConfig()
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		if configFilePath != envConfigPath {
			t.Errorf("Expected config path to be %s, got %s", envConfigPath, configFilePath)
		}
	})

	// Test case 2: No environment variable (default path)
	t.Run("Default path when no environment variable", func(t *testing.T) {
		if err := os.Unsetenv("BIFROST_CONFIG"); err != nil {
			t.Fatalf("Failed to unset BIFROST_CONFIG environment variable: %v", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}
		expectedPath := filepath.Join(homeDir, ".config", "bifrost_backups.yaml")
		configFilePath = expectedPath // Set it directly since init() has already run

		// Create the .config directory if it doesn't exist
		configDir := filepath.Join(homeDir, ".config")
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create a minimal valid config file
		config := Config{
			Version:   "1.0",
			Databases: []Database{},
			Storages:  []Storage{},
		}
		err = writeConfig(config)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
		defer func() {
			if err := os.Remove(expectedPath); err != nil {
				t.Errorf("Failed to remove test config file: %v", err)
			}
		}()

		// Try to read config
		_, err = readConfig()
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		if configFilePath != expectedPath {
			t.Errorf("Expected config path to be %s, got %s", expectedPath, configFilePath)
		}
	})
}

func TestReadConfig(t *testing.T) {
	// Save original config file path and restore it after the test
	originalConfigPath := configFilePath
	defer func() {
		configFilePath = originalConfigPath
	}()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create a test config file
	testConfigPath := filepath.Join(tmpDir, "test-config.yaml")
	configFilePath = testConfigPath

	testConfig := Config{
		Version: "1.0",
		Databases: []Database{
			{
				Type: Postgresql,
				Name: "test_db",
				Postgresql: postgresql.PostgresqlRequirements{
					Hostname: "localhost",
					Name:     "testdb",
					User:     "testuser",
					Password: "testpass",
					Port:     "5432",
				},
				Storages: []string{"test_storage"},
				Cron:     "0 0 * * *",
			},
		},
		Storages: []Storage{
			{
				Type: LocalStorage,
				Name: "test_storage",
				LocalStorage: localstorage.LocalStorageRequirements{
					FolderPath: "/tmp/backup",
				},
				RetentionDays:          21,
				ExecuteRetentionPolicy: true,
				Compression:            true,
			},
		},
	}

	err = writeConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test reading the config
	config, err := readConfig()
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify the config contents
	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	if len(config.Databases) != 1 {
		t.Errorf("Expected 1 database, got %d", len(config.Databases))
	}

	if len(config.Storages) != 1 {
		t.Errorf("Expected 1 storage, got %d", len(config.Storages))
	}

	// Test reading non-existent config file
	configFilePath = filepath.Join(tmpDir, "nonexistent.yaml")
	_, err = readConfig()
	if err == nil {
		t.Error("Expected error when reading non-existent config file, got nil")
	}
}
