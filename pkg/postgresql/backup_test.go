package postgresql

import (
	"os/exec"
	"testing"
)

func TestValidateRequirements(t *testing.T) {
	tests := []struct {
		name     string
		input    PostgresqlRequirements
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Empty requirements",
			input:    PostgresqlRequirements{},
			wantErr:  true,
			errMsg:   "database requirements cannot be empty",
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
		name     string
		input    PostgresqlRequirements
		expected []string
	}{
		{
			name: "Basic command args",
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
				"-F", "t",
				"-v",
				"testdb",
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
				"-F", "t",
				"-v",
				"testdb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCommandArgsBackup(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("buildCommandArgsBackup() got %v args, want %v args", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("buildCommandArgsBackup() arg[%d] = %v, want %v", i, result[i], tt.expected[i])
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
