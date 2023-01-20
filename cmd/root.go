/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/internal/log"
)

var (
	cfgFile string
	// host    string
	// apiKey  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ybm",
	Short: "ybm is a CLI for YugabyteDB Managed",
	Long: `ybm is a CLI tool which  helps in managing database infrastructure on YugabyteDB Managed,
	the fully managed DBaaS offering of YugabteDB.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setDefaults() {
	viper.SetDefault("host", "devcloud.yugabyte.com")
	viper.SetDefault("output", "table")
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("debug", false)
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	setDefaults()
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ybm-cli.yaml)")
	rootCmd.PersistentFlags().StringP("host", "", "", "YBM Api hostname")
	rootCmd.PersistentFlags().StringP("apiKey", "a", "", "YBM Api Key")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Select the desired output format (table, json, pretty). Default to table")
	rootCmd.PersistentFlags().StringP("logLevel", "l", "", "Select the desired log level format(info). Default to info")
	rootCmd.PersistentFlags().Bool("debug", false, "Use debug mode, same as --logLevel debug")

	//Bind peristents flags to viper
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("apiKey", rootCmd.PersistentFlags().Lookup("apiKey"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("logLevel"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ybm-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ybm-cli")
	}

	//Will check every environment variable starting with YBM_
	viper.SetEnvPrefix("ybm")
	//Read all enviromnent variable that match YBM_ENVNAME
	viper.AutomaticEnv() // read in environment variables that match

	// Set log level
	log.SetLogLevel(viper.GetString("logLevel"), viper.GetBool("debug"))
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Debugf("Using config file:", viper.ConfigFileUsed())
	}

}
