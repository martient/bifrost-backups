package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/setup/interactives"
	"github.com/martient/bifrost-backup/pkg/sqlite3"
	"github.com/martient/golang-utils/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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

var (
	configFilePath string
	config         Config
	configMutex    sync.RWMutex
	version        = "1.0"
)

func init() {
	generateDefaultConfig(version)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("error getting home directory: %w", err))
	}
	configFilePath = filepath.Join(homeDir, ".config", "bifrost_backups.yaml")
	loadConfig()
}

func loadConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Errorf("error opening config file: %w", err))
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		panic(fmt.Errorf("error decoding config file: %w", err))
	}
}

func RegisterDatabase(databaseType DatabaseType, name string, cron string, storages []string, requirements interface{}) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	newDatabase := &Database{
		Type:     databaseType,
		Name:     name,
		Cron:     cron,
		Storages: storages,
	}

	switch req := requirements.(type) {
	case *postgresql.PostgresqlRequirements:
		newDatabase.Postgresql = *req
	case *sqlite3.Sqlite3Requirements:
		newDatabase.Sqlite3 = *req
	default:
		return fmt.Errorf("unsupported database type: %T", requirements)
	}

	alreadyExist := false
	for i, ite := range config.Databases {
		if ite.Name == newDatabase.Name {
			config.Databases[i] = *newDatabase
			alreadyExist = true
			utils.LogInfo("Database %s already register, it as been updated", "REGISTER DATABASE", ite.Name)
			break
		}
	}
	if !alreadyExist {
		config.Databases = append(config.Databases, *newDatabase)
	}

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "error opening config file")
	}
	defer file.Close()

	if err := yaml.NewEncoder(file).Encode(config); err != nil {
		return errors.Wrap(err, "error encoding config file")
	}

	return nil
}
