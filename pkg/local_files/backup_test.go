package localfiles

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRequirements(t *testing.T) {
	tests := []struct {
		name    string
		input   LocalFilesRequirements
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty requirements",
			input:   LocalFilesRequirements{},
			wantErr: true,
			errMsg:  "source path cannot be empty",
		},
		{
			name: "Empty path",
			input: LocalFilesRequirements{
				Path: "",
			},
			wantErr: true,
			errMsg:  "source path cannot be empty",
		},
		{
			name: "Relative path",
			input: LocalFilesRequirements{
				Path: "relative/path",
			},
			wantErr: true,
			errMsg:  "source path must be absolute",
		},
		{
			name: "Valid requirements",
			input: LocalFilesRequirements{
				Path: "/absolute/path",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequirements(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validateRequirements() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestBackupAndRestore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "localfiles_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name        string
		setupFunc   func(string) (string, error)
		verifyFunc  func(string, string) error
		excludePats []string
		wantErr     bool
	}{
		{
			name: "Single file backup and restore",
			setupFunc: func(dir string) (string, error) {
				path := filepath.Join(dir, "test.txt")
				return path, os.WriteFile(path, []byte("test content"), 0644)
			},
			verifyFunc: func(original, restored string) error {
				origContent, err := os.ReadFile(original)
				if err != nil {
					return err
				}
				restContent, err := os.ReadFile(restored)
				if err != nil {
					return err
				}
				if !bytes.Equal(origContent, restContent) {
					return err
				}
				return nil
			},
		},
		{
			name: "Directory backup with exclusions",
			setupFunc: func(dir string) (string, error) {
				// Create test directory structure
				testDir := filepath.Join(dir, "testdir")
				if err := os.MkdirAll(testDir, 0755); err != nil {
					return "", err
				}
				// Create main file
				if err := os.WriteFile(filepath.Join(testDir, "main.txt"), []byte("main content"), 0644); err != nil {
					return "", err
				}
				// Create excluded file
				if err := os.WriteFile(filepath.Join(testDir, "temp.txt"), []byte("temp content"), 0644); err != nil {
					return "", err
				}
				return testDir, nil
			},
			excludePats: []string{"temp.txt"},
			verifyFunc: func(original, restored string) error {
				// Verify main file exists and content matches
				origContent, err := os.ReadFile(filepath.Join(original, "main.txt"))
				if err != nil {
					return err
				}
				restContent, err := os.ReadFile(filepath.Join(restored, "main.txt"))
				if err != nil {
					return err
				}
				if !bytes.Equal(origContent, restContent) {
					return err
				}
				// Verify excluded file doesn't exist in backup
				if _, err := os.Stat(filepath.Join(restored, "temp.txt")); !os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test files/directories
			sourcePath, err := tt.setupFunc(tempDir)
			if err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Create backup
			config := LocalFilesRequirements{
				Path:            sourcePath,
				ExcludePatterns: tt.excludePats,
			}
			backupBuffer, err := RunBackup(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Restore to a new location
			restorePath := filepath.Join(tempDir, "restored_"+filepath.Base(sourcePath))
			restoreConfig := LocalFilesRequirements{
				Path: restorePath,
			}
			err = RunRestore(restoreConfig, backupBuffer)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunRestore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify the restored content
			if err := tt.verifyFunc(sourcePath, restorePath); err != nil {
				t.Errorf("Verification failed: %v", err)
			}
		})
	}
}
