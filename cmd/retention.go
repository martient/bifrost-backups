package cmd

import (
	localstorage "github.com/martient/bifrost-backups/pkg/local_storage"
	"github.com/martient/bifrost-backups/pkg/s3"
	"github.com/martient/bifrost-backups/pkg/setup"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
)

var retentionCmd = &cobra.Command{
	Use:   "retention",
	Short: "Execute the retention policy operation",
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

			for i := 0; i < len(database.Storages); i++ {
				storage, err := setup.ReadStorageConfig(database.Storages[i])
				if err != nil {
					utils.LogError("Something went wrong during the config reading: %s", "CLI", err)
					return
				}
				if !storage.ExecuteRetentionPolicy {
					utils.LogInfo("Rentention policy of %s as been skipped for %s", "CLI", database.Name, storage.Name)
					continue
				}
				switch storage.Type {
				case setup.LocalStorage:
					err = localstorage.ExecuteRetentionPolicy(storage.LocalStorage, storage.RetentionDays)
				case setup.S3:
					err = s3.ExecuteRetentionPolicy(storage.S3, storage.RetentionDays)
				}
				if err != nil {
					utils.LogError("Something went wrong during the backup(s) cleaning process: %s", "CLI", err)
					return
				}
				utils.LogInfo("Backup(s) of %s as been deleted successfully following the retention policy of %s", "CLI", database.Name, storage.Name)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(retentionCmd)
	retentionCmd.Flags().String("name", "", "Database name")
}
