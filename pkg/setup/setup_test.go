package setup

import (
	"os"
	"testing"

	"github.com/martient/bifrost-backups/pkg/postgresql"
)

func TestSecureManager(t *testing.T) {
	manager, err := NewSecureManager(true)
	if err != nil {
		t.Fatalf("Failed to create secure manager: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Encrypt and decrypt normal string",
			input:   "test-password",
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "Special characters",
			input:   "test@password#123!",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := manager.encrypt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Test decryption
			decrypted, err := manager.decrypt(encrypted)
			if err != nil {
				t.Errorf("decrypt() error = %v", err)
				return
			}

			if decrypted != tt.input {
				t.Errorf("decrypt() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestConfigOperations(t *testing.T) {
	// Create a temporary directory for test config files
	tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	t.Run("WriteAndReadConfig", func(t *testing.T) {
		testConfig := Config{
			Version: version,
			Databases: []Database{
				{
					Type: Postgresql,
					Name: "testdb",
					Postgresql: postgresql.PostgresqlRequirements{
						Name:     "testdb",
						User:     "testuser",
						Password: "testpass",
						Hostname: "localhost",
						Port:     "5432",
					},
					Storages: []string{"local"},
					Cron:     "0 0 * * *",
				},
			},
		}

		// Test writing config
		err := writeConfig(testConfig)
		if err != nil {
			t.Fatalf("writeConfig() error = %v", err)
		}

		// Test reading config
		readConfig, err := readConfig()
		if err != nil {
			t.Fatalf("readConfig() error = %v", err)
		}

		// Verify the read config matches what we wrote
		if len(readConfig.Databases) != len(testConfig.Databases) {
			t.Errorf("readConfig() got %v databases, want %v", len(readConfig.Databases), len(testConfig.Databases))
		}

		// Check specific fields
		if readConfig.Databases[0].Name != testConfig.Databases[0].Name {
			t.Errorf("readConfig() database name = %v, want %v", readConfig.Databases[0].Name, testConfig.Databases[0].Name)
		}

		if readConfig.Databases[0].Type != Postgresql {
			t.Errorf("readConfig() database type = %v, want %v", readConfig.Databases[0].Type, Postgresql)
		}
	})
}

func TestDatabaseOperations(t *testing.T) {
	testConfig := Config{
		Version: version,
		Databases: []Database{
			{
				Type: Postgresql,
				Name: "testdb",
				Postgresql: postgresql.PostgresqlRequirements{
					Name:     "testdb",
					User:     "testuser",
					Password: "testpass",
					Hostname: "localhost",
					Port:     "5432",
				},
				Storages: []string{"local"},
				Cron:     "0 0 * * *",
			},
		},
	}

	// Write test config
	err := writeConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	t.Run("GetDatabaseConfigName", func(t *testing.T) {
		names, err := GetDatabaseConfigName()
		if err != nil {
			t.Fatalf("GetDatabaseConfigName() error = %v", err)
		}

		if len(names) != 1 || names[0] != "testdb" {
			t.Errorf("GetDatabaseConfigName() = %v, want [testdb]", names)
		}
	})

	t.Run("ReadDatabaseConfig", func(t *testing.T) {
		db, err := ReadDatabaseConfig("testdb")
		if err != nil {
			t.Fatalf("ReadDatabaseConfig() error = %v", err)
		}

		if db.Name != "testdb" || db.Type != Postgresql {
			t.Errorf("ReadDatabaseConfig() got unexpected database config: %+v", db)
		}
	})
}
