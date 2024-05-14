package setup

import (
	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/sqlite3"
)

type DatabaseType int64

const (
	Postgresql DatabaseType = 1
	Sqlite3    DatabaseType = 2
)

type StorageType int64

const (
	LocalStorage StorageType = 1
	S3           StorageType = 2
)

type Storage struct {
	Type         StorageType `json:"storage_type"`
	Name         string
	LocalStorage localstorage.LocalStorageRequirements
}

type Database struct {
	Type       DatabaseType `json:"database_type"`
	Postgresql postgresql.PostgresqlRequirements
	Sqlite3    sqlite3.Sqlite3Requirements
	Cron       string `json:"cron"`
}

type Config struct {
	Version   string `json:"version"`
	Databases []Database
	Storages  []Storage
}
