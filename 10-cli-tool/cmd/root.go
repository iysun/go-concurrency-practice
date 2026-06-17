package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd is the parent for all subcommands.
// Running the binary with no subcommand prints help.
var rootCmd = &cobra.Command{
	Use:   "devtool",
	Short: "A collection of developer utilities",
	Long:  `devtool bundles several handy utilities: port scanner, file finder, batch renamer.`,
}

// Execute is called from main(). Any error causes os.Exit(1).
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flag: applies to rootCmd and all its children
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.devtool.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")

	// Bind cobra flag to viper so both --verbose and DEVTOOL_VERBOSE env var work
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads the config file and ENV variables.
// TODO: expand to support per-project config discovery (walk up directory tree)
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.SetConfigName(".devtool")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("DEVTOOL") // DEVTOOL_VERBOSE, DEVTOOL_TIMEOUT, etc.
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
