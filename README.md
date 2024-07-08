# 🌉 Bifrost-backups

## 📚 Overview

Bifrost-backups is a powerful tool designed to simplify backup setup in production environments. Set it up once, and let it handle your backup needs effortlessly.

## 🧠 Logic

### Backup Process
1. Database (PostgreSQL, SQLite3, etc.) → Dump
2. Dump → Cipher
3. Ciphered Dump → Upload
4. Apply Retention Policy
5. Cleanup

### Restoration Process
1. Search registered storage
2. Download latest backup (or specified version)
3. Decipher backup
4. Restore database

### Retention Policy
Clean up backups older than the defined retention period (default: 21 days, configurable per storage)

## 🚀 Getting Started

```shell
> bifrost-backups
Backup solution for PostgreSQL, SQLite3

Usage:
  bifrost-backups [command]

Available Commands:
  backup            Execute the backup operation
  help              Help about any command
  register-database Register a new database
  register-storage  Register a new storage
  restore           Execute the restoration operation
  retention         Execute the retention policy operation
  update            Check if a new version is available

Flags:
      --disable-update-check   Disable auto update checking before execution
  -h, --help                   Help for bifrost-backups
  -t, --toggle                 Help message for toggle
  -v, --version                Version for bifrost-backups
  -y, --yes                    Auto accept manual questions (y/n)

Use "bifrost-backups [command] --help" for more information about a command.
```

### Usage

#### Backup

```shell
> bifrost-backups backup [flags]

Flags:
  -h, --help          Help for backup
      --name string   Database name
```

Examples:
- All registered databases: `bifrost-backups backup`
- Specific database: `bifrost-backups backup --name dev`

#### Restore

```shell
> bifrost-backups restore [flags]

Flags:
      --backup-name string    Backup name on your storage solution
  -h, --help                  Help for restore
      --name string           Database name
      --storage-name string   Specify a storage (uses the first found if not specified)
```

Examples:
- Restore using the first found storage: `bifrost-backups restore --name dev`
- Restore using a specific storage: `bifrost-backups restore --name dev --storage-name default`

#### Retention

```shell
> bifrost-backups retention [flags]

Flags:
  -h, --help          Help for retention
      --name string   Database name
```

Example:
- Execute retention policy: `bifrost-backups retention --name dev`

#### Register Database

```shell
> bifrost-backups register-database [flags]

Flags:
      --cron string       Backup cron (default "0,30 * * * *")
  -h, --help              Help for register-database
      --host string       Database host (default "localhost")
  -i, --interactive       Use interactive mode
      --name string       Database name
      --password string   Database user password
      --path string       Database path (SQLite3)
      --storages string   Storage names, e.g., "s3AWS,s3Azure,s3GCP" (default "default")
      --type int          Database type (1: PostgreSQL, 2: SQLite3)
      --user string       Database user
```

Examples:
- PostgreSQL: `bifrost-backups register-database --type 1 --name dev --user dev --password mySuperC@mplexPassWord --host 0.0.0.0 --storages default,s3AWS`
- SQLite3: `bifrost-backups register-database --type 2 --name dev --path ~/project/db/project.db`

#### Register Storage

```shell
> bifrost-backups register-storage [flags]

Flags:
      --access-key-id string       Access key ID
      --access-key-secret string   Access key secret
      --account-id string          Account ID
      --bucket-name string         Bucket name
      --cipher-key string          Custom cipher key (AES256 32bits) or leave empty to generate one
      --endpoint string            Endpoint
  -h, --help                       Help for register-storage
      --name string                Storage name (default "default")
  -i, --no-interactive             Use interactive mode
      --path string                Path for local storage output (default "~/bifrost-backups")
      --region string              Storage region (default "auto")
      --retention int              Backup retention period in days (default 21)
      --type int                   Storage type
```

Examples:
- S3: `bifrost-backups register-storage --type 1 --name s3AWS --access-key-id myAccessKey --access-key-secret mySecretKey --bucket-name myBucketName --region myRegion`
- Local storage: `bifrost-backups register-storage --type 2 --name localStorage --path ~/bifrost-backups`

## 🤝 Contributing

We welcome contributions from everyone! Here's how you can contribute:

1. Fork the repository
2. Create a new branch for your feature or bug fix
3. Make your changes and commit them
4. Push your changes to your forked repository
5. Create a pull request to the 'develop' branch of the main repository

Please see the [CONTRIBUTING](CONTRIBUTING) file for more details.

## 📝 License

Bifrost-backups is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0). See the [LICENSE](LICENSE) file for more details.
