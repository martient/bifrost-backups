package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martient/bifrost-backups/pkg/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfigExportAndMasterPassword(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// Save the original config file path and restore it after the test
	originalConfigPath := configFilePath
	defer func() {
		configFilePath = originalConfigPath
	}()

	// Set up a test config file path
	configFilePath = filepath.Join(tmpDir, "config.yaml")

	// Test cases
	tests := []struct {
		name     string
		setup    func(t *testing.T) Config
		password string
		wantErr  bool
	}{
		{
			name: "Set and validate master password",
			setup: func(t *testing.T) Config {
				config := Config{
					Version: "1.0",
					Databases: []Database{
						{
							Name: "test-db",
							Type: Postgresql,
							Postgresql: postgresql.PostgresqlRequirements{
								Hostname: "localhost",
								Port:     "5432",
								User:     "test",
								Password: "secret",
								Name:     "test-db",
							},
							Storages: []string{"test-storage"},
							Cron:     "@daily",
						},
					},
				}
				return config
			},
			password: "testpassword123",
			wantErr:  false,
		},
		{
			name: "Empty password should fail",
			setup: func(t *testing.T) Config {
				return Config{Version: "1.0"}
			},
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test config
			config := tt.setup(t)

			// Test setting master password
			err := SetMasterPassword(&config, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, config.MasterHash)

			// Update config with master password
			err = UpdateConfig(config)
			require.NoError(t, err)

			// Read back and verify
			readConfig, err := ReadConfigUnciphered()
			require.NoError(t, err)
			assert.Equal(t, config.MasterHash, readConfig.MasterHash)

			// Validate correct password
			err = ValidateMasterPassword(&readConfig, tt.password)
			assert.NoError(t, err)

			// Validate incorrect password
			err = ValidateMasterPassword(&readConfig, "wrongpassword")
			assert.Error(t, err)

			// Test config export
			exportPath := filepath.Join(tmpDir, "export.yaml")
			err = WriteConfigUnciphered(exportPath, readConfig)
			require.NoError(t, err)

			// Read and verify exported config
			exportedData, err := os.ReadFile(exportPath)
			require.NoError(t, err)

			var exportedConfig Config
			err = yaml.Unmarshal(exportedData, &exportedConfig)
			require.NoError(t, err)

			// Verify sensitive data is not encrypted in exported config
			if len(exportedConfig.Databases) > 0 && exportedConfig.Databases[0].Type == Postgresql {
				assert.Equal(t, "secret", exportedConfig.Databases[0].Postgresql.Password)
				assert.Equal(t, "test-db", exportedConfig.Databases[0].Postgresql.Name)
				assert.Equal(t, []string{"test-storage"}, exportedConfig.Databases[0].Storages)
			}
		})
	}
}

func TestExportConfigWithoutMasterPassword(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// Save the original config file path and restore it after the test
	originalConfigPath := configFilePath
	defer func() {
		configFilePath = originalConfigPath
	}()

	// Set up a test config file path
	configFilePath = filepath.Join(tmpDir, "config.yaml")

	// Create test config without master password
	config := Config{
		Version: "1.0",
		Databases: []Database{
			{
				Name: "test-db",
				Type: Postgresql,
				Postgresql: postgresql.PostgresqlRequirements{
					Hostname: "localhost",
					Port:     "5432",
					User:     "test",
					Password: "secret",
					Name:     "test-db",
				},
				Storages: []string{"test-storage"},
				Cron:     "@daily",
			},
		},
	}

	// Update config
	err = UpdateConfig(config)
	require.NoError(t, err)

	// Test export without master password
	exportPath := filepath.Join(tmpDir, "export.yaml")
	readConfig, err := ReadConfigUnciphered()
	require.NoError(t, err)

	err = WriteConfigUnciphered(exportPath, readConfig)
	require.NoError(t, err)

	// Read and verify exported config
	exportedData, err := os.ReadFile(exportPath)
	require.NoError(t, err)

	var exportedConfig Config
	err = yaml.Unmarshal(exportedData, &exportedConfig)
	require.NoError(t, err)

	// Verify sensitive data is not encrypted in exported config
	assert.Equal(t, "secret", exportedConfig.Databases[0].Postgresql.Password)
	assert.Equal(t, "test-db", exportedConfig.Databases[0].Postgresql.Name)
	assert.Equal(t, []string{"test-storage"}, exportedConfig.Databases[0].Storages)
}

func TestMasterPasswordEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		password string
		wantErr  bool
	}{
		{
			name:     "Nil config",
			config:   nil,
			password: "test",
			wantErr:  true,
		},
		{
			name:     "Very long password",
			config:   &Config{Version: "1.0"},
			password: string(make([]byte, 1024*1024)), // 1MB password
			wantErr:  true,
		},
		{
			name:     "Invalid master hash format",
			config:   &Config{Version: "1.0", MasterHash: "!@#$%^&*()"},
			password: "test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				if tt.config.MasterHash == "" {
					err := SetMasterPassword(tt.config, tt.password)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}

				err := ValidateMasterPassword(tt.config, tt.password)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			} else {
				err := SetMasterPassword(tt.config, tt.password)
				assert.Error(t, err)
			}
		})
	}
}
