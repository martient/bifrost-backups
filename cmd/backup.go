package cmd

import (
	"bytes"

	localstorage "github.com/martient/bifrost-backup/pkg/local_storage"
	"github.com/martient/bifrost-backup/pkg/postgresql"
	"github.com/martient/bifrost-backup/pkg/setup"
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

		var result *bytes.Buffer

		switch database.Type {
		case setup.Postgresql:
			result, err = postgresql.RunBackup(database.Postgresql)
		case setup.Sqlite3:
			// No implementation for Sqlite3 backup yet
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
			switch storage.Type {
			case setup.LocalStorage:
				err = localstorage.StoreBackup(storage.LocalStorage, result)
			case setup.S3:
				// No implementation for Sqlite3 backup yet
			}
			if err != nil {
				utils.LogError("Something went wrong during the storing process: %s", "CLI", err)
				return
			}
			utils.LogInfo("Backup of %s successfully stored with %s", "CLI", database.Name, storage.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().String("name", "", "Database name")
}
