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
	cmd := exec.Command(pgDumpPath, args...) //#nosec

	if database.Password != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PGPASSWORD="+database.Password)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			utils.LogErrorInterface("pg_dump error: %s", "POSTGRESQL", err, errMsg)
		}
		utils.LogErrorInterface("Failed to backup database '%s': %v", "POSTGRESQL", err, database.Name)
		return nil, fmt.Errorf("backup failed: %v: %s", err, errMsg)
	}

	if stdout.Len() == 0 {
		return nil, fmt.Errorf("backup produced no output, this may indicate a connection issue")
	}

	return &stdout, nil
}

func validateRequirements(database PostgresqlRequirements) error {
	if database == (PostgresqlRequirements{}) {
		return fmt.Errorf("database requirements cannot be empty")
	}

	if database.Name == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	if database.User == "" {
		return fmt.Errorf("database user cannot be empty")
	}

	// Default hostname to localhost if not specified
	if database.Hostname == "" {
		database.Hostname = "127.0.0.1"
	}

	// Default port to 5432 if not specified
	if database.Port == "" {
		database.Port = "5432"
	}

	return nil
}

func buildCommandArgsBackup(database PostgresqlRequirements) []string {
	var args []string

	// Always specify host and port for consistency
	if database.Hostname != "" {
		args = append(args, "-h", database.Hostname)
	} else {
		args = append(args, "-h", "127.0.0.1")
	}
	if database.Port != "" {
		args = append(args, "-p", database.Port)
	} else {
		args = append(args, "-p", "5432")
	}
	args = append(args, "-U", database.User)

	// Database name should be last
	args = append(args, database.Name)

	return args
}
