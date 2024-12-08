package cmd

import (
	"github.com/martient/bifrost-backups/pkg/updater"
	"github.com/martient/golang-utils/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	BEMversion    string
	noEncryption  bool
	cfgFile       string
	disableUpdate bool
	updateChannel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Bifrost-backup",
	Short: "Backup manager",
	Long:  `Backup solution for Postgresql, Sqlite3`,
	// examples and usage of using your application. For example:

	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(versionFormated string, version string) {
	rootCmd.Version = versionFormated
	BEMversion = version
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf   .}}{{end}} {{printf  .Version}}`)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bifrost-backup.yaml)")
	rootCmd.PersistentFlags().BoolVar(&disableUpdate, "disable-update-check", false, "Disable automatic update check")
	rootCmd.PersistentFlags().StringVar(&updateChannel, "update-channel", "stable", "Update channel (stable/beta)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "Auto accept manual question y/n")
	rootCmd.PersistentFlags().BoolVar(&noEncryption, "no-encryption", false, "Disable configuration encryption (not recommended for production)")

	// If update check is not disabled, check for updates
	if !disableUpdate {
		go checkForUpdates()
	}
}

func initConfig() {
	// if jsonConfigFile != "" {
	// 	viper.SetConfigFile(jsonConfigFile)
	// }
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		utils.LogInfo("Using config file: %s", "CLI", viper.ConfigFileUsed())
	}
}

func checkForUpdates() {
	channel := updater.StableChannel
	if updateChannel == "beta" {
		channel = updater.BetaChannel
	}

	up, err := updater.New(BEMversion, "martient/bifrost-backup", channel)
	if err != nil {
		utils.LogError("Failed to initialize updater: %v", "Updater", err)
		return
	}

	release, err := up.Check()
	if err != nil {
		utils.LogError("Failed to check for updates: %v", "Updater", err)
		return
	}

	if release != nil {
		utils.LogInfo("New version %s available. Run 'bifrost-backup update' to upgrade.", "Updater", release.Version)
	}
}
