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
	Type          StorageType                           `yaml:"storage_type"`
	Name          string                                `yaml:"name"`
	CipherKey     string                                `yaml:"cypher_key"`
	RetentionDays int                                   `yaml:"retention_days" default:"21"`
	Compression   bool                                  `yaml:"compression" default:"true"`
	LocalStorage  localstorage.LocalStorageRequirements `yaml:"local_storage,omitempty"` // Make local_storage optional
	S3            s3.S3Requirements                     `yaml:"s3,omitempty"`            // Make s3 optional
}

type Database struct {
	Type       DatabaseType                      `yaml:"database_type"`
	Name       string                            `yaml:"name"`
	Postgresql postgresql.PostgresqlRequirements `yaml:"postgresql,omitempty"`
	Sqlite3    sqlite3.Sqlite3Requirements       `yaml:"sqlite3,omitempty"`
	Storages   []string                          `yaml:"storages"`
	Cron       string                            `yaml:"cron"`
}

type Config struct {
	Version   string     `yaml:"version"`
	Databases []Database `yaml:"databases,omitempty"`
	Storages  []Storage  `yaml:"storages,omitempty"`
}
