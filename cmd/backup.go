package cmd

import (
	"bytes"
	"encoding/base64"

	"github.com/martient/bifrost-backup/pkg/crypto"
	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/s3"
	"github.com/martient/bifrost-backup/pkg/setup"
	"github.com/martient/bifrost-backup/pkg/sqlite3"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
)

// generateCmd represents the load command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Execute the backup operation",
	Run: func(cmd *cobra.Command, args []string) {
		if disableUpdateCheck, _ := rootCmd.Flags().GetBool("disable-update-check"); !disableUpdateCheck {
			doConfirmAndSelfUpdate()
		}

		var names []string

		name, err := cmd.Flags().GetString("name")
		if name == "" {
			utils.LogInfo("getting backups source", "CLI", err)
			fetched_names, err := setup.GetDatabaseConfigName()
			if err != nil {
				utils.LogError("Something went wrong during the config reading: %s", "CLI", err)
				return
			} else if len(fetched_names) == 0 {
				utils.LogWarning("No backup source found", "CLI")
				return
			}
			names = append(names, fetched_names...)
		} else {
			names = append(names, name)
		}

		for i := 0; i < len(names); i++ {
			name = names[i]

			database, err := setup.ReadDatabaseConfig(name)
			if err != nil {
				utils.LogError("Something went wrong during the config reading: %s", "CLI", err)
				return
			}

			var result *bytes.Buffer

			switch database.Type {
			case setup.Postgresql:
				result, err = postgresql.RunBackup(database.Postgresql)
			case setup.Sqlite3:
				result, err = sqlite3.RunBackup(database.Sqlite3)
			}

			if err != nil {
				utils.LogError("Something went wrong during the backuping process: %s", "CLI", err)
				return
			}
			for i := 0; i < len(database.Storages); i++ {
				storage, err := setup.ReadStorageConfig(database.Storages[i])
				if err != nil {
					utils.LogError("Something went wrong during the config reading: %s", "CLI", err)
					return
				}
				cipher_key, err := base64.StdEncoding.DecodeString(storage.CipherKey)
				if err != nil {
					utils.LogError("Something went wrong during the convertion of the cipher key process: %s", "CLI", err)
					return
				}
				cipher_result, err := crypto.Cipher(cipher_key, result.Bytes())
				if err != nil {
					utils.LogError("Something went wrong during the encryption process: %s", "CLI", err)
					return
				}
				switch storage.Type {
				case setup.LocalStorage:
					err = localstorage.StoreBackup(storage.LocalStorage, cipher_result)
				case setup.S3:
					err = s3.StoreBackup(storage.S3, result)
				}
				if err != nil {
					utils.LogError("Something went wrong during the storing process: %s", "CLI", err)
					return
				}
				utils.LogInfo("Backup of %s successfully stored with %s", "CLI", database.Name, storage.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().String("name", "", "Database name")
}
