package cmd

import (
	"os"
	"path/filepath"

	"github.com/martient/bifrost-backup/pkg/setup"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
)

var exportConfigCmd = &cobra.Command{
	Use:   "export-config",
	Short: "Export configuration in unciphered format",
	Long:  `Export the configuration file with all sensitive information decrypted for backup purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			utils.LogError("Output path cannot be empty", "CLI", nil)
			return
		}

		// Create output directory if it doesn't exist
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0700); err != nil {
			utils.LogError("Failed to create output directory", "CLI", err)
			return
		}

		// Read current config
		config, err := setup.ReadConfigUnciphered()
		if err != nil {
			utils.LogError("Failed to read config", "CLI", err)
			return
		}

		// Write the unciphered config to the specified location
		if err := setup.WriteConfigUnciphered(outputPath, config); err != nil {
			utils.LogError("Failed to write unciphered config", "CLI", err)
			return
		}

		utils.LogInfo("Configuration successfully exported to %s", "CLI", outputPath)
	},
}

func init() {
	rootCmd.AddCommand(exportConfigCmd)
	exportConfigCmd.Flags().String("output", "", "Path where to save the unciphered configuration file")
	exportConfigCmd.MarkFlagRequired("output")
}
