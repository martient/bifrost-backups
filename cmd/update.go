package cmd

import (
	"fmt"
	"os"

	"github.com/martient/bifrost-backups/pkg/updater"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
)

var (
	autoConfirm bool
	channel     string
	noRestart   bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and perform updates",
	Long:  `Check for new versions of Bifrost Backup and update if available`,
	Run: func(cmd *cobra.Command, args []string) {
		updateChannel := updater.StableChannel
		if channel == "beta" {
			updateChannel = updater.BetaChannel
		}

		up, err := updater.New(BEMversion, "martient/bifrost-backup", updateChannel)
		if err != nil {
			utils.LogError("Failed to initialize updater: %v", "Updater", err)
			os.Exit(1)
		}

		release, err := up.Check()
		if err != nil {
			utils.LogError("Failed to check for updates: %v", "Updater", err)
			os.Exit(1)
		}

		if release == nil {
			utils.LogInfo("No updates available", "Updater")
			return
		}

		utils.LogInfo("New version %s available", "Updater", release.Version)

		// Display changelog
		changelog := up.GetChangelog(release)
		if changelog != "" {
			fmt.Printf("\nChangelog:\n%s\n", changelog)
		}

		if !autoConfirm {
			fmt.Printf("Do you want to update to version %s? [y/N] ", release.Version)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				return
			}
		}

		if err := up.Update(release); err != nil {
			utils.LogError("Failed to update: %v", "Updater", err)
			os.Exit(1)
		}

		utils.LogInfo("Successfully updated to version %s", "Updater", release.Version)

		// If this point is reached, it means the update failed to restart
		// This can happen if --no-restart was used
		if noRestart {
			utils.LogInfo("Update complete. Please restart the application manually.", "Updater")
		}
	},
}

func doConfirmAndSelfUpdate() {

}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&autoConfirm, "yes", "y", false, "Automatically confirm update")
	updateCmd.Flags().StringVarP(&channel, "channel", "c", "stable", "Update channel (stable/beta)")
	updateCmd.Flags().BoolVar(&noRestart, "no-restart", false, "Do not restart the command after update")
}
