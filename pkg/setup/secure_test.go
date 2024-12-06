package setup

import (
	"strings"
	"testing"

	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/s3"
)

func TestNewSecureManager(t *testing.T) {
	tests := []struct {
		name         string
		noEncryption bool
		wantErr      bool
	}{
		{
			name:         "Create manager with encryption",
			noEncryption: false,
			wantErr:      false,
		},
		{
			name:         "Create manager without encryption",
			noEncryption: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewSecureManager(tt.noEncryption)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSecureManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if manager == nil {
					t.Error("NewSecureManager() returned nil manager")
				} else if manager.noEncryption != tt.noEncryption {
					t.Errorf("NewSecureManager() noEncryption = %v, want %v", manager.noEncryption, tt.noEncryption)
				}
			}
		})
	}
}

func TestSecureConfig(t *testing.T) {
	// Create a test configuration
	config := &Config{
		Version: "1.0",
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
			},
		},
		Storages: []Storage{
			{
				Type: S3,
				Name: "tests3",
				S3: s3.S3Requirements{
					BucketName:      "test-bucket",
					AccessKeyId:     "test-key-id",
					AccessKeySecret: "test-secret",
					Region:          "us-west-1",
				},
				CipherKey: "test-cipher-key",
			},
		},
	}

	tests := []struct {
		name         string
		config       *Config
		noEncryption bool
		wantErr      bool
	}{
		{
			name:         "Secure config with encryption",
			config:       config,
			noEncryption: false,
			wantErr:      false,
		},
		{
			name:         "Secure config without encryption",
			config:       config,
			noEncryption: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewSecureManager(tt.noEncryption)
			if err != nil {
				t.Fatalf("Failed to create secure manager: %v", err)
			}

			err = manager.SecureConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.noEncryption {
				// Check if sensitive fields are encrypted
				if !strings.HasPrefix(tt.config.Databases[0].Postgresql.Password, "ENC[AES256,") {
					t.Error("PostgreSQL password was not encrypted")
				}
				if !strings.HasPrefix(tt.config.Storages[0].S3.AccessKeySecret, "ENC[AES256,") {
					t.Error("S3 access key secret was not encrypted")
				}
				if !strings.HasPrefix(tt.config.Storages[0].CipherKey, "ENC[AES256,") {
					t.Error("Storage cipher key was not encrypted")
				}
			}
		})
	}
}

func TestDecryptConfig(t *testing.T) {
	// Create a secure manager for encryption
	manager, err := NewSecureManager(false)
	if err != nil {
		t.Fatalf("Failed to create secure manager: %v", err)
	}

	// Create a test configuration with unencrypted values
	originalConfig := &Config{
		Version: "1.0",
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
			},
		},
		Storages: []Storage{
			{
				Type: S3,
				Name: "tests3",
				S3: s3.S3Requirements{
					BucketName:      "test-bucket",
					AccessKeyId:     "test-key-id",
					AccessKeySecret: "test-secret",
					Region:          "us-west-1",
				},
				CipherKey: "test-cipher-key",
			},
		},
	}

	// First encrypt the configuration
	err = manager.SecureConfig(originalConfig)
	if err != nil {
		t.Fatalf("Failed to encrypt config: %v", err)
	}

	tests := []struct {
		name         string
		config       *Config
		noEncryption bool
		wantErr      bool
	}{
		{
			name:         "Decrypt config with encryption",
			config:       originalConfig,
			noEncryption: false,
			wantErr:      false,
		},
		{
			name:         "Decrypt config without encryption",
			config:       originalConfig,
			noEncryption: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewSecureManager(tt.noEncryption)
			if err != nil {
				t.Fatalf("Failed to create secure manager: %v", err)
			}

			err = manager.DecryptConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.noEncryption {
				// Check if sensitive fields are decrypted correctly
				if tt.config.Databases[0].Postgresql.Password != "testpass" {
					t.Errorf("PostgreSQL password not decrypted correctly, got %v, want %v",
						tt.config.Databases[0].Postgresql.Password, "testpass")
				}
				if tt.config.Storages[0].S3.AccessKeySecret != "test-secret" {
					t.Errorf("S3 access key secret not decrypted correctly, got %v, want %v",
						tt.config.Storages[0].S3.AccessKeySecret, "test-secret")
				}
				if tt.config.Storages[0].CipherKey != "test-cipher-key" {
					t.Errorf("Storage cipher key not decrypted correctly, got %v, want %v",
						tt.config.Storages[0].CipherKey, "test-cipher-key")
				}
			}
		})
	}
}
