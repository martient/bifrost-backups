package sqlite3

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	internalutils "github.com/martient/bifrost-backups/pkg/utils"
	"github.com/martient/golang-utils/utils"
)

var allowedSqliteCommands = map[string][]string{
	"sqlite3": {
		"-c",
		"DROP TABLE IF EXISTS",
		".dump",
		".exit",
		"-d",
		"--database",
	},
}

func RunRestoration(database Sqlite3Requirements, backup *bytes.Buffer) error {
	if backup.Len() <= 0 {
		return fmt.Errorf("backup can't be empty for the restoration process")
	} else if err := validateRequirements(database); err != nil {
		return err
	}

	sqlite3Path, err := exec.LookPath("sqlite3")
	if err != nil {
		return fmt.Errorf("sqlite3 command not found: %w", err)
	}

	// Validate database path
	dbDir := filepath.Dir(database.Path)
	if err := internalutils.ValidatePath(database.Path, []string{dbDir}); err != nil {
		return fmt.Errorf("invalid database path: %w", err)
	}

	args := buildCommandArgsClear(database)
	if err := internalutils.ValidateCommand(sqlite3Path, args, allowedSqliteCommands); err != nil {
		return fmt.Errorf("invalid command arguments: %w", err)
	}

	cmdClear := exec.Command(sqlite3Path, args...) //#nosec
	cmdClear.Env = []string{
		"PATH=" + os.Getenv("PATH"),
	}

	var stderrClear bytes.Buffer
	cmdClear.Stderr = &stderrClear

	_, err = os.Stat(database.Path)
	if !os.IsNotExist(err) {
		err = os.Remove(database.Path)
		if err != nil {
			utils.LogErrorInterface("Failed to remove database file '%s'", "SQLITE3", err, database.Path)
			return fmt.Errorf("database file '%s' could not be removed: %w", database.Path, err)
		}
		utils.LogInfo("Database '%s' removed", "SQLITE3", database.Path)
	}

	err = cmdClear.Run()
	if err != nil {
		utils.LogErrorInterface("Failed to recreate the database '%s', %s", "SQLITE3", stderrClear.String(), err)
		return fmt.Errorf("database creation before restoration failed: %w", err)
	}
	utils.LogInfo("Database '%s' recreated", "SQLITE3", database.Path)

	args = buildCommandArgsRestore(database)
	if err := internalutils.ValidateCommand(sqlite3Path, args, allowedSqliteCommands); err != nil {
		return fmt.Errorf("invalid restore command arguments: %w", err)
	}

	cmdRestore := exec.Command(sqlite3Path, args...) //#nosec
	cmdRestore.Env = []string{
		"PATH=" + os.Getenv("PATH"),
	}
	cmdRestore.Stdin = bytes.NewBuffer(backup.Bytes())

	var stderrRestore bytes.Buffer
	cmdRestore.Stderr = &stderrRestore

	err = cmdRestore.Run()
	if err != nil {
		utils.LogErrorInterface("Failed to restore database '%s', %s", "SQLITE3", stderrRestore.String(), err)
		return fmt.Errorf("backup restoration failed: %w", err)
	}
	utils.LogInfo("Database '%s' restored", "SQLITE3", database.Path)
	return nil
}

func buildCommandArgsClear(database Sqlite3Requirements) []string {
	var args []string

	args = append(args, database.Path)
	args = append(args, ".database")
	return args
}

func buildCommandArgsRestore(database Sqlite3Requirements) []string {
	var args []string

	args = append(args, database.Path)
	return args
}
