package sqlite3

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/martient/golang-utils/utils"
)

const sqlite3Command = "sqlite3"

func RunBackup(database Sqlite3Requirements) (*bytes.Buffer, error) {
	if err := validateRequirements(database); err != nil {
		return nil, err
	}

	sqlite3Path, err := exec.LookPath(sqlite3Command)
	if err != nil {
		return nil, fmt.Errorf("sqlite3 command not found: %w", err)
	}

	args := buildCommandArgsBackup(database)
	cmd := exec.Command(sqlite3Path, args...)

	var buffer bytes.Buffer
	cmd.Stdout = &buffer

	err = cmd.Run()
	if err != nil {
		utils.LogError("Failed to backup database '%s'", "SQLITE3", err)
		return nil, fmt.Errorf("backup failed: %w", err)
	}
	return &buffer, nil
}

func validateRequirements(database Sqlite3Requirements) error {
	if database == (Sqlite3Requirements{}) {
		return fmt.Errorf("database requirements cannot be empty")
	}

	// Add additional validation checks for port, hostname, etc.

	return nil
}

func buildCommandArgsBackup(database Sqlite3Requirements) []string {
	var args []string

	args = append(args, database.Path)
	args = append(args, ".backup")

	return args
}
