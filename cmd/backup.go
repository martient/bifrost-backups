package cmd

import (
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
			utils.LogError("Something went wrong %s", "CLI", err)
			return
		}

		switch database.Type {
		case setup.Postgresql:
			postgresql.RunBackup(database.Postgresql)
		case setup.Sqlite3:
			// No implementation for Sqlite3 backup yet
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().String("name", "", "Database name")
}
