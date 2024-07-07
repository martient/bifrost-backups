package cmd

import (
	"os"
	"strings"

	"github.com/martient/bifrost-backup/pkg/setup"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
)

// generateCmd represents the load command
var registerDatabaseCmd = &cobra.Command{
	Use:   "register-database",
	Short: "Register a new database",
	Long:  `Register a new database to be backup`,
	Run: func(cmd *cobra.Command, args []string) {
		if disableUpdateCheck, _ := rootCmd.Flags().GetBool("disable-update-check"); !disableUpdateCheck {
			doConfirmAndSelfUpdate()
		}

		if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
			setup.InteractiveRegisterDatabase()
		} else {
			db_int, _ := cmd.Flags().GetInt64("type")
			db_type := setup.DatabaseType(db_int)
			switch db_type {
			case 1:
				host, _ := cmd.Flags().GetString("host")
				name, _ := cmd.Flags().GetString("name")
				user, _ := cmd.Flags().GetString("user")
				password, _ := cmd.Flags().GetString("password")
				registered, err := setup.RegisterPostgresqlDatabase(host, user, name, password)
				if err != nil {
					utils.LogError("Your database haven't been registerd: %s", "CLI", err)
					os.Exit(1)
				}
				cron, _ := cmd.Flags().GetString("cron")
				storagesStr, _ := cmd.Flags().GetString("storages")
				storages := strings.Split(storagesStr, ",")
				err = setup.RegisterDatabase(db_type, name, cron, storages, registered)
				if err != nil {
					utils.LogError("Saved failed: %s", "CLI", err)
					os.Exit(1)
				}
			case 2:
				path, _ := cmd.Flags().GetString("path")
				registered, err := setup.RegisterSqlite3Database(path)
				if err != nil {
					utils.LogError("Your database haven't been registerd: %s", "CLI", err)
					os.Exit(1)
				}
				cron, _ := cmd.Flags().GetString("cron")
				storagesStr, _ := cmd.Flags().GetString("storages")
				storages := strings.Split(storagesStr, ",")
				name, _ := cmd.Flags().GetString("name")
				err = setup.RegisterDatabase(db_type, name, cron, storages, registered)
				if err != nil {
					utils.LogError("Saved failed: %s", "CLI", err)
					os.Exit(1)
				}
			default:
				utils.LogWarning("Please choose between the available type of database with --type", "CLI")
				os.Exit(-1)
			}

		}
		// byteValue, _ := io.ReadAll(jsonFile)
		// result := environmentmanager.GenerateEnvFile(byteValue, newEnvFilePath, readOnlyEnvFilesPath)
		// if result != 0 {
		// 	os.Exit(1)
		// }
	},
}

func init() {
	rootCmd.AddCommand(registerDatabaseCmd)
	registerDatabaseCmd.Flags().BoolP("interactive", "i", false, "Use the interactive mode")
	registerDatabaseCmd.Flags().Int64("type", -1, "Database type (1: postgresql, 2: sqlite3)")
	registerDatabaseCmd.Flags().String("path", "", "Database path (sqlite3)")
	registerDatabaseCmd.Flags().String("host", "localhost", "Database host")
	registerDatabaseCmd.Flags().String("name", "", "Database name")
	registerDatabaseCmd.Flags().String("user", "", "Database user")
	registerDatabaseCmd.Flags().String("password", "", "Database user password")
	registerDatabaseCmd.Flags().String("cron", "0,30 * * * *", "Backup cron (default, one backup each 30 minutes)")
	registerDatabaseCmd.Flags().String("storages", "default", "Storage name ex:\"s3AWS, s3Azure, s3GCP\" (default, local storage of the home user .bifrost_backups)")
}
