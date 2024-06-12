package cmd

import (
	"os"

	"github.com/martient/bifrost-backup/pkg/setup"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
)

// generateCmd represents the load command
var registerStorageCmd = &cobra.Command{
	Use:   "register-storage",
	Short: "Register a new storage",
	Long:  `Register a new storage for the backup`,
	Run: func(cmd *cobra.Command, args []string) {
		if disableUpdateCheck, _ := rootCmd.Flags().GetBool("disable-update-check"); !disableUpdateCheck {
			doConfirmAndSelfUpdate()
		}

		if interactive, _ := cmd.Flags().GetBool("no-interactive"); !interactive {
			setup.InteractiveRegisterStorage()
		} else {
			storage_int, _ := cmd.Flags().GetInt64("type")
			storage_type := setup.StorageType(storage_int)
			switch storage_type {
			case 1:
				path, _ := cmd.Flags().GetString("path")
				registered, err := setup.RegisterLocalStorage(path)
				if err != nil {
					utils.LogError("Your storage haven't been registerd: %s", "CLI", err)
					os.Exit(1)
				}
				name, _ := cmd.Flags().GetString("name")
				err = setup.RegisterStorage(storage_type, name, registered)
				if err != nil {
					utils.LogError("Saved failed: %s", "CLI", err)
					os.Exit(1)
				}
			case 2:
				bucket_name, _ := cmd.Flags().GetString("bucket-name")
				access_key_id, _ := cmd.Flags().GetString("access-key-id")
				access_key_secret, _ := cmd.Flags().GetString("access-key-secret")
				endpoint, _ := cmd.Flags().GetString("endpoint")
				region, _ := cmd.Flags().GetString("region")
				registered, err := setup.RegisterS3Storage(bucket_name, access_key_id, access_key_secret, endpoint, region)
				if err != nil {
					utils.LogError("Your storage haven't been registerd: %s", "CLI", err)
					os.Exit(1)
				}
				name, _ := cmd.Flags().GetString("name")
				err = setup.RegisterStorage(storage_type, name, registered)
				if err != nil {
					utils.LogError("Saved failed: %s", "CLI", err)
					os.Exit(1)
				}
			default:
				utils.LogWarning("Please choose between the available type of storage with --type", "CLI")
				os.Exit(-1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(registerStorageCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	registerStorageCmd.Flags().BoolP("no-interactive", "i", false, "Use the interactive mode")
	registerStorageCmd.Flags().Int64("type", -1, "Database type")
	registerStorageCmd.Flags().String("name", "default", "Storage name")
	registerStorageCmd.Flags().String("path", "~/bifrost-backups", "Path for the output target folder in the local storage")

	registerStorageCmd.Flags().String("bucket-name", "", "Bucket name")
	registerStorageCmd.Flags().String("account-id", "", "Account Id")
	registerStorageCmd.Flags().String("access-key-id", "", "Access key Id")
	registerStorageCmd.Flags().String("access-key-secret", "", "Access key secret")
	registerStorageCmd.Flags().String("endpoint", "", "Endpoint")
	registerStorageCmd.Flags().String("region", "auto", "Region of storage")
}
