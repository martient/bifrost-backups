package localstorage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalStorageOperations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bifrost-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("Store operations", func(t *testing.T) {
		tests := []struct {
			name           string
			storage        LocalStorageRequirements
			buffer         *bytes.Buffer
			useCompression bool
			wantErr        bool
		}{
			{
				name: "Store uncompressed backup",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				buffer:         bytes.NewBufferString("test backup content"),
				useCompression: false,
				wantErr:        false,
			},
			{
				name: "Store compressed backup",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				buffer:         bytes.NewBufferString("test backup content"),
				useCompression: true,
				wantErr:        false,
			},
			{
				name:           "Empty storage requirements",
				storage:        LocalStorageRequirements{},
				buffer:         bytes.NewBufferString("test backup content"),
				useCompression: false,
				wantErr:        true,
			},
			{
				name: "Empty buffer",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				buffer:         nil,
				useCompression: false,
				wantErr:        true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := StoreBackup(tt.storage, tt.buffer, tt.useCompression)
				if (err != nil) != tt.wantErr {
					t.Errorf("StoreBackup() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr && tt.buffer != nil {
					// Verify the backup was stored
					files, err := os.ReadDir(tmpDir)
					if err != nil {
						t.Errorf("Failed to read temp directory: %v", err)
						return
					}

					if len(files) == 0 {
						t.Errorf("No backup file was created")
					}
				}
			})
		}
	})

	t.Run("Pull operations", func(t *testing.T) {
		// Create a test backup file
		testContent := "test backup content"
		backupName := time.Now().UTC().Format("2006-01-02T15:04:005Z")
		backupPath := filepath.Join(tmpDir, backupName)
		err = os.WriteFile(backupPath, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test backup file: %v", err)
		}

		tests := []struct {
			name           string
			storage        LocalStorageRequirements
			backupName     string
			useCompression bool
			wantErr        bool
			wantContent    string
		}{
			{
				name: "Pull existing backup",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				backupName:     backupName,
				useCompression: false,
				wantErr:        false,
				wantContent:    testContent,
			},
			{
				name: "Pull non-existent backup",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				backupName:     "nonexistent.bak",
				useCompression: false,
				wantErr:        true,
			},
			{
				name:           "Empty storage requirements",
				storage:        LocalStorageRequirements{},
				backupName:     backupName,
				useCompression: false,
				wantErr:        true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				buffer, err := PullBackup(tt.storage, tt.backupName, tt.useCompression)
				if (err != nil) != tt.wantErr {
					t.Errorf("PullBackup() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr {
					if buffer == nil {
						t.Error("PullBackup() returned nil buffer")
						return
					}

					content := buffer.String()
					if content != tt.wantContent {
						t.Errorf("PullBackup() content = %v, want %v", content, tt.wantContent)
					}
				}
			})
		}
	})

	t.Run("Retention operations", func(t *testing.T) {
		tests := []struct {
			name                   string
			storage                LocalStorageRequirements
			executeRetentionPolicy bool
			retentionDays          int
			wantErr                bool
			wantFiles              int
		}{
			{
				name: "Retain 23 days",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				retentionDays:          23,
				executeRetentionPolicy: true,
				wantErr:                false,
				wantFiles:              3, // Should keep the 23-day-old and current backups
			},
			{
				name: "Retain 15 days",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				retentionDays:          15,
				executeRetentionPolicy: true,
				wantErr:                false,
				wantFiles:              2, // Should keep the 10-day-old and current backups
			},
			{
				name: "Retain all backups",
				storage: LocalStorageRequirements{
					FolderPath: tmpDir,
				},
				executeRetentionPolicy: false,
				wantErr:                false,
				wantFiles:              4, // Should keep all backups
			},
			{
				name:          "Empty storage requirements",
				storage:       LocalStorageRequirements{},
				retentionDays: 15,
				wantErr:       true,
				wantFiles:     0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				//setup: create test backup files with different dates
				now := time.Now().UTC()
				backupDates := []time.Time{
					now.AddDate(0, 0, -30), // 30 days old
					now.AddDate(0, 0, -20), // 20 days old
					now.AddDate(0, 0, -10), // 10 days old
					now,                    // current
				}

				for _, date := range backupDates {
					backupName := date.Format("2006-01-02T15:04:005Z")
					backupPath := filepath.Join(tmpDir, backupName)
					err = os.WriteFile(backupPath, []byte("test backup"), 0644)
					if err != nil {
						t.Fatalf("Failed to create test backup file: %v", err)
					}
				}

				if tt.executeRetentionPolicy {
					err := ExecuteRetentionPolicy(tt.storage, tt.retentionDays)
					if err != nil {
						t.Errorf("ExecuteRetentionPolicy() error = %v", err)
						return
					}
				}

				if !tt.wantErr {
					files, err := os.ReadDir(tmpDir)
					if err != nil {
						t.Errorf("Failed to read temp directory: %v", err)
						return
					}

					if len(files) != tt.wantFiles {
						t.Errorf("ExecuteRetentionPolicy() files = %v, want %v", len(files), tt.wantFiles)
					}
				}
			})
		}
	})
}
