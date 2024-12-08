package setup

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/martient/bifrost-backups/pkg/local_files"
	"github.com/martient/bifrost-backups/pkg/postgresql"
	"github.com/martient/bifrost-backups/pkg/setup/interactives"
	"github.com/martient/bifrost-backups/pkg/sqlite3"
	"github.com/martient/golang-utils/utils"
	"github.com/robfig/cron/v3"
)

func InteractiveRegisterDatabase() {
	if _, err := tea.NewProgram(interactives.PostgresqlInitialModel()).Run(); err != nil {
		utils.LogError("Could not start program: %s\n", "Register datbase", err)
		os.Exit(1)
	}
}

func RegisterPostgresqlDatabase(host string, user string, name string, password string) (*postgresql.PostgresqlRequirements, error) {
	requirements := &postgresql.PostgresqlRequirements{}
	if len(user) <= 0 {
		return nil, fmt.Errorf("username can't be empty")
	} else if len(name) <= 0 {
		return nil, fmt.Errorf("database name can't be empty")
	}
	requirements.Hostname = host
	requirements.User = user
	requirements.Name = name
	requirements.Password = password
	return requirements, nil
}

func RegisterSqlite3Database(path string) (*sqlite3.Sqlite3Requirements, error) {
	requirements := &sqlite3.Sqlite3Requirements{}
	if len(path) <= 0 {
		return nil, fmt.Errorf("path can't be empty")
	}
	requirements.Path = path
	return requirements, nil
}

func RegisterLocalFilesDatabase(path string) (*localfiles.LocalFilesRequirements, error) {
	requirements := &localfiles.LocalFilesRequirements{}
	if len(path) <= 0 {
		return nil, fmt.Errorf("path can't be empty")
	}
	requirements.Path = path
	return requirements, nil
}

func RegisterDatabase(databaseType DatabaseType, name string, cronExpr string, storages []string, requirements interface{}) error {
	// Validate inputs
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	if requirements == nil {
		return fmt.Errorf("database requirements cannot be nil")
	}

	// Validate cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Lock during the entire operation to prevent race conditions
	configMutex.Lock()
	defer configMutex.Unlock()

	// Read the current config
	currentConfig, err := readConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	newDatabase := &Database{
		Type:     databaseType,
		Name:     name,
		Cron:     cronExpr,
		Storages: storages,
	}

	switch req := requirements.(type) {
	case *postgresql.PostgresqlRequirements:
		if req.User == "" || req.Name == "" {
			return fmt.Errorf("PostgreSQL database user and name cannot be empty")
		}
		newDatabase.Postgresql = *req
	case *sqlite3.Sqlite3Requirements:
		if req.Path == "" {
			return fmt.Errorf("SQLite3 database path cannot be empty")
		}
		newDatabase.Sqlite3 = *req
	case *localfiles.LocalFilesRequirements:
		if req.Path == "" {
			return fmt.Errorf("Local files database path cannot be empty")
		}
		newDatabase.LocalFiles = *req
	default:
		return fmt.Errorf("unsupported database type: %T", requirements)
	}

	// Find and update existing database, or append new one
	found := false
	for i := range currentConfig.Databases {
		if currentConfig.Databases[i].Name == name {
			currentConfig.Databases[i] = *newDatabase
			found = true
			utils.LogInfo("Database %s configuration updated", "REGISTER DATABASE", name)
			break
		}
	}

	if !found {
		currentConfig.Databases = append(currentConfig.Databases, *newDatabase)
		utils.LogInfo("Database %s registered", "REGISTER DATABASE", name)
	}

	// Write the updated config back
	if err := writeConfig(currentConfig); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
