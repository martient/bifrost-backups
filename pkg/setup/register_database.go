package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/setup/interactives"
	"github.com/martient/bifrost-backup/pkg/sqlite3"
	"github.com/martient/golang-utils/utils"
)

func InteractiveRegisterDatabase() {
	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	fmt.Println("Error getting home directory:", err)
	// 	return
	// }
	// configFilePath := filepath.Join(homeDir, ".config", "bifrost_backups.json")

	// file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_TRUNC, 0644)
	// if err != nil {
	// 	fmt.Println("Error opening config file:", err)
	// 	return
	// }
	// defer file.Close()

	if _, err := tea.NewProgram(interactives.PostgresqlInitialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}

	// encoder := json.NewEncoder(file)
	// if err := encoder.Encode(config); err != nil {
	// 	fmt.Println("Error encoding JSON:", err)
	// 	return
	// }
}

func RegisterPostgresqlDatabase(host string, user string, name string, password string) (*postgresql.PostgresqlRequirements, error) {
	requirements := &postgresql.PostgresqlRequirements{}
	if len(user) <= 0 {
		utils.LogError("Username can't be empty", "Register postgresql database", nil)
		return nil, fmt.Errorf("username can't be empty")
	} else if len(name) <= 0 {
		utils.LogError("Database name can't be empty", "Register postgresql database", nil)
		return nil, fmt.Errorf("database name can't be empty")
	}
	requirements.Hostname = host
	requirements.User = user
	requirements.Name = name
	requirements.Password = password
	return requirements, nil
}

func RegisterDatabase(databaseType DatabaseType, name string, cron string, database interface{}) error {
	if database == nil {
		return fmt.Errorf("database is null")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory")
	}
	configFilePath := filepath.Join(homeDir, ".config", "bifrost_backups.json")

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file")
	}
	defer file.Close()

	config := Config{}

	json.NewDecoder(file).Decode(&config)

	newDatabase := &Database{
		Type: databaseType,
		Name: name,
		Cron: cron,
	}

	switch db := database.(type) {
	case *postgresql.PostgresqlRequirements:
		newDatabase.Postgresql = *db
	case *sqlite3.Sqlite3Requirements:
		newDatabase.Sqlite3 = *db
	default:
		return fmt.Errorf("unsupported database type")
	}

	config.Databases = append(config.Databases, *newDatabase)

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding JSON")
	}
	return nil
}
