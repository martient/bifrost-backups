package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martient/bifrost-backups/pkg/local_files"
	localstorage "github.com/martient/bifrost-backups/pkg/local_storage"
	"github.com/martient/bifrost-backups/pkg/postgresql"
	"github.com/martient/bifrost-backups/pkg/s3"
	"github.com/martient/bifrost-backups/pkg/sqlite3"
)

func TestRegisterDatabase(t *testing.T) {
	tests := []struct {
		name         string
		dbType       DatabaseType
		dbName       string
		requirements interface{}
		cronExpr     string
		wantErr      bool
	}{
		{
			name:   "Register PostgreSQL database",
			dbType: Postgresql,
			dbName: "test_postgres",
			requirements: &postgresql.PostgresqlRequirements{
				Hostname: "localhost",
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Port:     "5432",
			},
			cronExpr: "0 0 * * *",
			wantErr:  false,
		},
		{
			name:   "Register SQLite3 database",
			dbType: Sqlite3,
			dbName: "test_sqlite",
			requirements: &sqlite3.Sqlite3Requirements{
				Path: "/tmp/test.db",
			},
			cronExpr: "0 0 * * *",
			wantErr:  false,
		},
		{
			name:   "Register Local Files database",
			dbType: LocalFiles,
			dbName: "test_local",
			requirements: &localfiles.LocalFilesRequirements{
				Path: "/tmp/test.db",
			},
			cronExpr: "0 0 * * *",
			wantErr:  false,
		},
		{
			name:         "Invalid database type",
			dbType:       DatabaseType(999),
			dbName:       "test_invalid",
			requirements: &struct{}{}, // Invalid type
			cronExpr:     "0 0 * * *",
			wantErr:      true,
		},
		{
			name:   "Empty database name",
			dbType: Postgresql,
			dbName: "",
			requirements: &postgresql.PostgresqlRequirements{
				Hostname: "localhost",
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
			},
			cronExpr: "0 0 * * *",
			wantErr:  true,
		},
		{
			name:   "Invalid cron expression",
			dbType: Postgresql,
			dbName: "test_invalid_cron",
			requirements: &postgresql.PostgresqlRequirements{
				Hostname: "localhost",
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
			},
			cronExpr: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original config file path and restore it after the test
			originalConfigPath := configFilePath
			defer func() {
				configFilePath = originalConfigPath
				os.Unsetenv("BIFROST_CONFIG")
			}()

			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create initial config file
			configPath := filepath.Join(tmpDir, "config.yaml")
			config := Config{
				Version:   "1.0",
				Databases: []Database{},
				Storages:  []Storage{},
			}

			// Set config file path and write initial config
			configFilePath = configPath
			err = writeConfig(config)
			if err != nil {
				t.Fatalf("Failed to write initial config: %v", err)
			}

			err = RegisterDatabase(tt.dbType, tt.dbName, tt.cronExpr, []string{}, tt.requirements)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterStorage(t *testing.T) {
	tests := []struct {
		name          string
		storageType   StorageType
		storageName   string
		retentionDays int
		cipherKey     string
		compression   bool
		storageReq    interface{}
		wantErr       bool
	}{
		{
			name:          "Register local storage",
			storageType:   LocalStorage,
			storageName:   "test_local",
			retentionDays: 7,
			cipherKey:     "test-key",
			compression:   true,
			storageReq:    &localstorage.LocalStorageRequirements{FolderPath: "/tmp/backup"},
			wantErr:       false,
		},
		{
			name:          "Register S3 storage",
			storageType:   S3,
			storageName:   "test_s3",
			retentionDays: 7,
			cipherKey:     "test-key",
			compression:   true,
			storageReq:    &s3.S3Requirements{BucketName: "test-bucket", Region: "us-east-1"},
			wantErr:       false,
		},
		{
			name:          "Invalid storage type",
			storageType:   StorageType(999),
			storageName:   "test_invalid",
			retentionDays: 7,
			cipherKey:     "test-key",
			compression:   true,
			storageReq:    &struct{}{}, // Invalid type
			wantErr:       true,
		},
		{
			name:          "Empty storage name",
			storageType:   LocalStorage,
			storageName:   "",
			retentionDays: 7,
			cipherKey:     "test-key",
			compression:   true,
			storageReq:    &localstorage.LocalStorageRequirements{FolderPath: "/tmp/backup"},
			wantErr:       true,
		},
		{
			name:          "Invalid retention period",
			storageType:   LocalStorage,
			storageName:   "test_invalid_retention",
			retentionDays: -1,
			cipherKey:     "test-key",
			compression:   true,
			storageReq:    &localstorage.LocalStorageRequirements{FolderPath: "/tmp/backup"},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original config file path and restore it after the test
			originalConfigPath := configFilePath
			defer func() {
				configFilePath = originalConfigPath
				os.Unsetenv("BIFROST_CONFIG")
			}()

			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create initial config file
			configPath := filepath.Join(tmpDir, "config.yaml")
			config := Config{
				Version:   "1.0",
				Databases: []Database{},
				Storages:  []Storage{},
			}

			// Set config file path and write initial config
			configFilePath = configPath
			err = writeConfig(config)
			if err != nil {
				t.Fatalf("Failed to write initial config: %v", err)
			}

			err = RegisterStorage(tt.storageType, tt.storageName, tt.retentionDays, tt.cipherKey, tt.compression, tt.storageReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
