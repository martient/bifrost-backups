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
	args := buildCommandArgsBackup(database)
	cmd := exec.Command(pgRestorePath, args...)

	if database.Password != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PGPASSWORD="+database.Password)
	}
	cmd.Stdin = backup

	err = cmd.Run()
	if err != nil {
		utils.LogError("Failed to restore database '%s'", "POSTGRESQL", err)
		return fmt.Errorf("backup failed: %w", err)
	}
	return nil
}

func buildCommandArgsRestore(database PostgresqlRequirements) []string {
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

	args = append(args, database.Name)

	return args
}
