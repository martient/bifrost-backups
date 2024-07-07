package setup

import (
	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/s3"
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
	Type          StorageType                           `json:"storage_type"`
	Name          string                                `json:"name"`
	CipherKey     string                                `json:"cypher_key"`
	RetentionDays int                                   `json:"retention_days" default:"21"`
	LocalStorage  localstorage.LocalStorageRequirements `json:"local_storage"`
	S3            s3.S3Requirements                     `json:"s3"`
}

type Database struct {
	Type       DatabaseType                      `json:"database_type"`
	Name       string                            `json:"name"`
	Postgresql postgresql.PostgresqlRequirements `json:"postgresql"`
	Sqlite3    sqlite3.Sqlite3Requirements       `json:"sqlite3"`
	Storages   []string                          `json:"storages"`
	Cron       string                            `json:"cron"`
}

type Config struct {
	Version   string     `json:"version"`
	Databases []Database `json:"databases"`
	Storages  []Storage  `json:"storages"`
}
