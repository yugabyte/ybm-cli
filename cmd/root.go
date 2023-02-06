/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/backup"
	"github.com/yugabyte/ybm-cli/cmd/cdcsink"
	"github.com/yugabyte/ybm-cli/cmd/cdcstream"
	"github.com/yugabyte/ybm-cli/cmd/cluster"
	"github.com/yugabyte/ybm-cli/cmd/nal"
	"github.com/yugabyte/ybm-cli/cmd/readreplica"
	"github.com/yugabyte/ybm-cli/cmd/vpc"
	"github.com/yugabyte/ybm-cli/cmd/vpcpeering"
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
	viper.SetDefault("no-color", false)
	viper.SetDefault("wait", false)
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
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colors in output , default to false")
	rootCmd.PersistentFlags().Bool("wait", false, "Wait until the task is completed, otherwise it will exit immediately, default to false")

	//Bind peristents flags to viper
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("apiKey", rootCmd.PersistentFlags().Lookup("apiKey"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("logLevel"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("wait", rootCmd.PersistentFlags().Lookup("wait"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(cluster.ClusterCmd)
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(cdcsink.CDCSinkCmd)
	rootCmd.AddCommand(cdcstream.CDCStreamCmd)
	rootCmd.AddCommand(nal.NalCmd)
	rootCmd.AddCommand(readreplica.ReadReplicaCmd)
	rootCmd.AddCommand(vpc.VPCCmd)
	rootCmd.AddCommand(vpcpeering.VPCPeeringCmd)
	rootCmd.AddCommand(configureCmd)

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
	//Set Logrus formatter options
	log.SetFormatter()
	// Set log level
	log.SetLogLevel(viper.GetString("logLevel"), viper.GetBool("debug"))
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}

}
