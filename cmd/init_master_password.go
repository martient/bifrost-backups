package cmd

import (
	"syscall"

	"github.com/martient/bifrost-backup/pkg/setup"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var initMasterPasswordCmd = &cobra.Command{
	Use:   "init-master-password",
	Short: "Initialize or change the master password",
	Long:  `Set or change the master password used for exporting configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read current config
		config, err := setup.ReadConfigUnciphered()
		if err != nil {
			utils.LogError("Failed to read config", "CLI", err)
			return
		}

		// If master password is already set, verify old password first
		if config.MasterHash != "" {
			utils.LogInfo("Enter current master password: ", "CLI")
			currentPassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				utils.LogError("Failed to read password", "CLI", err)
				return
			}

			if err := setup.ValidateMasterPassword(&config, string(currentPassword)); err != nil {
				utils.LogError("Invalid current master password", "CLI", err)
				return
			}
		}

		// Prompt for new password
		utils.LogInfo("\nEnter new master password: ", "CLI")
		password1, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			utils.LogError("Failed to read password", "CLI", err)
			return
		}

		utils.LogInfo("\nConfirm new master password: ", "CLI")
		password2, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			utils.LogError("Failed to read password", "CLI", err)
			return
		}

		if string(password1) != string(password2) {
			utils.LogError("Passwords do not match", "CLI", nil)
			return
		}

		// Set new master password
		if err := setup.SetMasterPassword(&config, string(password1)); err != nil {
			utils.LogError("Failed to set master password", "CLI", err)
			return
		}

		// Update the config with the new master password
		if err := setup.UpdateConfig(config); err != nil {
			utils.LogError("Failed to update config", "CLI", err)
			return
		}

		utils.LogInfo("\nMaster password successfully set", "CLI")
	},
}

func init() {
	rootCmd.AddCommand(initMasterPasswordCmd)
}
