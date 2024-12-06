package postgresql

import (
	"bytes"
	"os/exec"
	"reflect"
	"testing"
)

func TestValidateRequirements(t *testing.T) {
	tests := []struct {
		name    string
		input   PostgresqlRequirements
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty requirements",
			input:   PostgresqlRequirements{},
			wantErr: true,
			errMsg:  "database requirements cannot be empty",
		},
		{
			name: "Missing database name",
			input: PostgresqlRequirements{
				User:     "testuser",
				Password: "testpass",
				Hostname: "localhost",
				Port:     "5432",
			},
			wantErr: true,
			errMsg:  "database name cannot be empty",
		},
		{
			name: "Missing user",
			input: PostgresqlRequirements{
				Name:     "testdb",
				Password: "testpass",
				Hostname: "localhost",
				Port:     "5432",
			},
			wantErr: true,
			errMsg:  "database user cannot be empty",
		},
		{
			name: "Valid requirements with defaults",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "Valid requirements fully specified",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "localhost",
				Port:     "5432",
			},
			wantErr: false,
		},
		{
			name: "Empty hostname should default to 127.0.0.1",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Port:     "5432",
			},
			wantErr: false,
		},
		{
			name: "Empty port should default to 5432",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "localhost",
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

func TestBuildCommandArgsBackup(t *testing.T) {
	tests := []struct {
		name         string
		requirements PostgresqlRequirements
		wantArgs     []string
		wantErr      bool
	}{
		{
			name: "Basic command args",
			requirements: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "127.0.0.1",
				Port:     "5432",
			},
			wantArgs: []string{
				"-h", "127.0.0.1",
				"-p", "5432",
				"-U", "testuser",
				"testdb",
			},
			wantErr: false,
		},
		{
			name: "Custom host and port",
			requirements: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "localhost",
				Port:     "5433",
			},
			wantArgs: []string{
				"-h", "localhost",
				"-p", "5433",
				"-U", "testuser",
				"testdb",
			},
			wantErr: false,
		},
		{
			name: "Default host and port",
			requirements: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "",
				Port:     "",
			},
			wantArgs: []string{
				"-h", "127.0.0.1",
				"-p", "5432",
				"-U", "testuser",
				"testdb",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequirements(tt.requirements)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				args := buildCommandArgsBackup(tt.requirements)
				if !reflect.DeepEqual(args, tt.wantArgs) {
					t.Errorf("buildCommandArgsBackup() = %v, want %v", args, tt.wantArgs)
					for i := range args {
						if i < len(tt.wantArgs) && args[i] != tt.wantArgs[i] {
							t.Errorf("buildCommandArgsBackup() arg[%d] = %v, want %v", i, args[i], tt.wantArgs[i])
						}
					}
				}
			}
		})
	}
}

func TestRunBackup(t *testing.T) {
	// Skip if pg_dump is not installed
	if _, err := exec.LookPath(pgDumpCommand); err != nil {
		t.Skip("pg_dump not installed, skipping integration test")
	}

	tests := []struct {
		name    string
		input   PostgresqlRequirements
		wantErr bool
		errType string
	}{
		{
			name: "Invalid credentials",
			input: PostgresqlRequirements{
				Name:     "nonexistentdb",
				User:     "invaliduser",
				Password: "wrongpass",
				Hostname: "localhost",
				Port:     "5432",
			},
			wantErr: true,
			errType: "connection error",
		},
		{
			name: "Invalid port",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "localhost",
				Port:     "54321",
			},
			wantErr: true,
			errType: "connection error",
		},
		{
			name: "Invalid hostname",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "nonexistent.host",
				Port:     "5432",
			},
			wantErr: true,
			errType: "connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RunBackup(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunRestoration(t *testing.T) {
	// Skip if pg_restore is not installed
	if _, err := exec.LookPath(pgRestoreCommand); err != nil {
		t.Skip("pg_restore not installed, skipping integration test")
	}

	tests := []struct {
		name    string
		input   PostgresqlRequirements
		backup  *bytes.Buffer
		wantErr bool
	}{
		{
			name: "Empty backup",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "localhost",
				Port:     "5432",
			},
			backup:  &bytes.Buffer{},
			wantErr: true,
		},
		{
			name: "Invalid backup format",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				Hostname: "localhost",
				Port:     "5432",
			},
			backup:  bytes.NewBufferString("invalid backup content"),
			wantErr: true,
		},
		{
			name: "Invalid credentials",
			input: PostgresqlRequirements{
				Name:     "nonexistentdb",
				User:     "invaliduser",
				Password: "wrongpass",
				Hostname: "localhost",
				Port:     "5432",
			},
			backup:  bytes.NewBufferString("some backup content"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunRestoration(tt.input, tt.backup)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunRestoration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildCommandArgsRestore(t *testing.T) {
	tempFile := "/tmp/test.sql"
	tests := []struct {
		name     string
		input    PostgresqlRequirements
		expected []string
	}{
		{
			name: "Basic restore args",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Hostname: "localhost",
				Port:     "5432",
			},
			expected: []string{
				"-h", "localhost",
				"-p", "5432",
				"-U", "testuser",
				"-d", "testdb",
				"--clean", "-v",
				tempFile,
			},
		},
		{
			name: "Custom host and port",
			input: PostgresqlRequirements{
				Name:     "testdb",
				User:     "testuser",
				Hostname: "custom.host",
				Port:     "5433",
			},
			expected: []string{
				"-h", "custom.host",
				"-p", "5433",
				"-U", "testuser",
				"-d", "testdb",
				"--clean", "-v",
				tempFile,
			},
		},
		{
			name: "Minimal args",
			input: PostgresqlRequirements{
				Name: "testdb",
				User: "testuser",
			},
			expected: []string{
				"-U", "testuser",
				"-d", "testdb",
				"--clean", "-v",
				tempFile,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCommandArgsRestore(tt.input, tempFile)
			if len(result) != len(tt.expected) {
				t.Errorf("buildCommandArgsRestore() got %v args, want %v args", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("buildCommandArgsRestore() arg[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
