package postgresql

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/martient/golang-utils/utils"
)

const pgDumpCommand = "pg_dump"

func RunBackup(database PostgresqlRequirements) (*bytes.Buffer, error) {
	if err := validateRequirements(database); err != nil {
		return nil, err
	}

	pgDumpPath, err := exec.LookPath(pgDumpCommand)
	if err != nil {
		return nil, fmt.Errorf("pg_dump command not found: %w", err)
	}

	args := buildCommandArgsBackup(database)
	cmd := exec.Command(pgDumpPath, args...)

	if database.Password != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PGPASSWORD="+database.Password)
	}

	var buffer bytes.Buffer
	cmd.Stdout = &buffer

	err = cmd.Run()
	if err != nil {
		utils.LogError("Failed to backup database '%s'", "POSTGRESQL", err)
		return nil, fmt.Errorf("backup failed: %w", err)
	}
	return &buffer, nil
}

func validateRequirements(database PostgresqlRequirements) error {
	if database == (PostgresqlRequirements{}) {
		return fmt.Errorf("database requirements cannot be empty")
	}

	// Add additional validation checks for port, hostname, etc.

	return nil
}

func buildCommandArgsBackup(database PostgresqlRequirements) []string {
	var args []string

	if database.Hostname != "" {
		args = append(args, "-h", database.Hostname)
	}

	if database.Port != "" {
		args = append(args, "-p", database.Port)
	}

	if database.User != "" {
		args = append(args, "-U", database.User)
	}

	args = append(args, "-F", "t")
	args = append(args, database.Name)

	return args
}
