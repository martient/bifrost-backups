package postgresql

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/martient/golang-utils/utils"
)

const pgRestoreCommand = "pg_restore"

func RunRestoration(database PostgresqlRequirements, backup *bytes.Buffer) error {
	if backup.Len() <= 0 {
		return fmt.Errorf("backup can't be empty for the restoration process")
	} else if err := validateRequirements(database); err != nil {
		return err
	}

	pgRestorePath, err := exec.LookPath(pgRestoreCommand)
	if err != nil {
		return fmt.Errorf("pg_dump command not found: %w", err)
	}

	tempFile, err := os.CreateTemp("", "bf_*.sql")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temporary file

	// Write the backup buffer to the temporary file
	_, err = backup.WriteTo(tempFile)
	if err != nil {
		return fmt.Errorf("failed to write backup to temporary file: %w", err)
	}

	// Close the temporary file
	err = tempFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	args := buildCommandArgsRestore(database, tempFile.Name())
	cmd := exec.Command(pgRestorePath, args...)

	if database.Password != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PGPASSWORD="+database.Password)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils.LogErrorInterface("Failed to restore database '%s', %s", "POSTGRESQL", stderr.String(), err)
		return fmt.Errorf("backup failed: %w", err)
	}
	return nil
}

func buildCommandArgsRestore(database PostgresqlRequirements, tempFile string) []string {
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

	if database.Name != "" {
		args = append(args, "-d", database.Name)
	}

	args = append(args, "--clean", "-v", tempFile)

	return args
}
