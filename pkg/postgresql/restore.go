package postgresql

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	internalutils "github.com/martient/bifrost-backup/pkg/utils"
	"github.com/martient/golang-utils/utils"
)

const pgRestoreCommand = "pg_restore"

var allowedPgCommands = map[string][]string{
	"pg_restore": {
		"-h",
		"-U",
		"-d",
		"-p",
		"--no-owner",
		"--no-privileges",
		"--clean",
		"--if-exists",
	},
}

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

	tempFile, err := os.CreateTemp("", "pg_restore_*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temporary file

	// Set secure permissions on the temporary file
	if err := os.Chmod(tempFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set temporary file permissions: %w", err)
	}

	// Write the backup buffer to the temporary file
	_, err = backup.WriteTo(tempFile)
	if err != nil {
		return fmt.Errorf("failed to write backup to temporary file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	args := buildCommandArgsRestore(database, tempFile.Name())
	if err := internalutils.ValidateCommand(pgRestorePath, args, allowedPgCommands); err != nil {
		return fmt.Errorf("invalid command arguments: %w", err)
	}

	cmd := exec.Command(pgRestorePath, args...) //#nosec
	if database.Password != "" {
		// Use a clean environment with only necessary variables
		cmd.Env = []string{
			"PGPASSWORD=" + database.Password,
			"PATH=" + os.Getenv("PATH"), // Needed for pg_restore
			"HOME=" + os.Getenv("HOME"), // Needed for .pgpass
		}
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils.LogErrorInterface("Failed to restore database '%s', %s", "POSTGRESQL", stderr.String(), err)
		return fmt.Errorf("backup failed: %w", err)
	}
	utils.LogInfo("Database '%s' restored", "POSTGRESQL", database.Name)
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
