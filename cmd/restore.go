package cmd

import (
	"bytes"

	"github.com/martient/bifrost-backup/pkg/crypto"
	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/s3"
	"github.com/martient/bifrost-backup/pkg/setup"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
)

// generateCmd represents the load command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Execute the restoration operation",
	Run: func(cmd *cobra.Command, args []string) {
		if disableUpdateCheck, _ := rootCmd.Flags().GetBool("disable-update-check"); !disableUpdateCheck {
			doConfirmAndSelfUpdate()
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil && name == "" {
			utils.LogError("name can't be empty", "CLI", err)
			return
		}

		database, err := setup.ReadDatabaseConfig(name)
		if err != nil {
			utils.LogError("Something went wrong during the config reading: %s", "CLI", err)
			return
		}
		storage_name, _ := cmd.Flags().GetString("storage-name")
		backup_name, _ := cmd.Flags().GetString("backup-name")
		var result *bytes.Buffer
		var cipher_key string

		for i := 0; i < len(database.Storages); i++ {
			storage, err := setup.ReadStorageConfig(database.Storages[i])
			if err != nil {
				utils.LogError("Something went wrong during the config reading: %s", "CLI", err)
				return
			}
			if storage_name != "" && storage_name == storage.Name {
				switch storage.Type {
				case setup.LocalStorage:
					cipher_key = storage.CipherKey
					result, err = localstorage.PullBackup(storage.LocalStorage, backup_name)
				case setup.S3:
					cipher_key = storage.CipherKey
					result, err = s3.PullBackup(storage.S3, backup_name)
				default:
					utils.LogError("Unsupported storage type used during the restore process...", "CLI", nil)
					return
				}
				if err != nil {
					utils.LogError("Something went wrong during the retrieving process: %s", "CLI", err)
					return
				}
				utils.LogInfo("Backup of %s successfully retrieved with %s", "CLI", database.Name, storage.Name)
			}
		}

		if result == nil {
			utils.LogWarning("The backup fetched seems to be empty...", "CLI")
			return
		}

		decipher_result, err := crypto.Decipher([]byte(cipher_key), result.Bytes())
		if err != nil {
			utils.LogError("Something went wrong during the encryption process: %s", "CLI", err)
			return
		}
		switch database.Type {
		case setup.Postgresql:
			err = postgresql.RunRestoration(database.Postgresql, decipher_result)
		case setup.Sqlite3:
			// No implementation for Sqlite3 backup yet
		}

		if err != nil {
			utils.LogError("Something went wrong during the restoring process: %s", "CLI", err)
			return
		}
		utils.LogInfo("Backup of %s successfully restored", "CLI", database.Name)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().String("name", "", "Database name")
	restoreCmd.Flags().String("storage-name", "", "You must define a specific storage otherwise it gonna take the first found (not handled yet)")
	restoreCmd.Flags().String("backup-name", "", "Backup name on your storage solution (not handled yet)")
}
