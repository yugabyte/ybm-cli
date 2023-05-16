// Licensed to Yugabyte, Inc. under one or more contributor license
// agreements. See the NOTICE file distributed with this work for
// additional information regarding copyright ownership. Yugabyte
// licenses this file to you under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/backup"
	"github.com/yugabyte/ybm-cli/cmd/cdc"
	"github.com/yugabyte/ybm-cli/cmd/cluster"
	"github.com/yugabyte/ybm-cli/cmd/nal"
	"github.com/yugabyte/ybm-cli/cmd/region"
	"github.com/yugabyte/ybm-cli/cmd/role"
	"github.com/yugabyte/ybm-cli/cmd/signup"
	"github.com/yugabyte/ybm-cli/cmd/tools"
	"github.com/yugabyte/ybm-cli/cmd/util"
	"github.com/yugabyte/ybm-cli/cmd/vpc"

	"github.com/yugabyte/ybm-cli/internal/log"
	"github.com/yugabyte/ybm-cli/internal/releases"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ybm",
	Short: "ybm - Effortlessly manage your DB infrastructure on YugabyteDB Managed (DBaaS) from command line!",
	Long:  `ybm - Effortlessly manage your DB infrastructure on YugabyteDB Managed (DBaaS) from command line!`,

	Run: func(cmd *cobra.Command, args []string) {
		myFigure := figure.NewFigure("ybm", "", true)
		myFigure.Print()
		logrus.Printf("\n")
		cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if strings.HasPrefix(cmd.CommandPath(), "ybm completion") {
			return
		}
		releases.PrintUpgradeMessageIfNeeded()

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setDefaults() {
	viper.SetDefault("host", "cloud.yugabyte.com")
	viper.SetDefault("output", "table")
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("debug", false)
	viper.SetDefault("no-color", false)
	viper.SetDefault("wait", false)
	viper.SetDefault("timeout", time.Duration(7*24*time.Hour))
	viper.SetDefault("lastVersionAvailable", "v0.0.0")
	viper.SetDefault("lastCheckedTime", 0)
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.EnableCaseInsensitive = true
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	setDefaults()
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ybm-cli.yaml)")
	rootCmd.PersistentFlags().StringP("apiKey", "a", "", "YBM Api Key")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Select the desired output format (table, json, pretty). Default to table")
	rootCmd.PersistentFlags().StringP("logLevel", "l", "", "Select the desired log level format(info). Default to info")
	rootCmd.PersistentFlags().Bool("debug", false, "Use debug mode, same as --logLevel debug")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colors in output , default to false")
	rootCmd.PersistentFlags().Bool("wait", false, "Wait until the task is completed, otherwise it will exit immediately, default to false")
	rootCmd.PersistentFlags().Duration("timeout", 7*24*time.Hour, "Wait command timeout, example: 5m, 1h.")

	//Bind peristents flags to viper
	viper.BindPFlag("apiKey", rootCmd.PersistentFlags().Lookup("apiKey"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("logLevel"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("wait", rootCmd.PersistentFlags().Lookup("wait"))
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

	// Make host configurable only if the CONFIGURE_URL feature flag is set to true
	if util.IsFeatureFlagEnabled(util.CONFIGURE_URL) {
		rootCmd.PersistentFlags().StringP("host", "", "", "YBM Api hostname")
		viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	}

	rootCmd.AddCommand(cluster.ClusterCmd)
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(nal.NalCmd)
	rootCmd.AddCommand(vpc.VPCCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(signup.SignUpCmd)
	rootCmd.AddCommand(region.CloudRegionsCmd)
	rootCmd.AddCommand(role.RoleCmd)
	util.AddCommandIfFeatureFlag(rootCmd, tools.ToolsCmd, util.TOOLS)
	util.AddCommandIfFeatureFlag(rootCmd, cdc.CdcCmd, util.CDC)

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
