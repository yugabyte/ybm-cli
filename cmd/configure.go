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
	"fmt"
	"os"
	"strings"

	"github.com/yugabyte/ybm-cli/cmd/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate ybm CLI",
	Long:  "Authenticate the ybm CLI through this command by providing the API Key.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Enter API Key: ")
		var apiKey string
		var host string
		var data []byte
		data, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			logrus.Fatalln("Could not read apiKey: ", err)
		}
		apiKey = string(data)

		// Validate that apiKey is a valid JWT token and that the token is not expired
		if strings.TrimSpace(apiKey) == "" {
			logrus.Fatalln("ApiKey cannot be empty")
		}
		expired, err := util.IsJwtTokenExpired(apiKey)
		if err != nil {
			logrus.Fatalln("ApiKey is invalid")
		}
		if expired {
			logrus.Fatalln("ApiKey is expired")
		}

		viper.GetViper().Set("apikey", &apiKey)

		// If the feature flag is enabled, prompt the user for URL
		if util.IsFeatureFlagEnabled(util.CONFIGURE_URL) {
			fmt.Print("Enter Host (leave empty for default cloud.yugabyte.com): ")
			fmt.Scanln(&host)
			if strings.TrimSpace(host) == "" {
				host = "cloud.yugabyte.com"

			}
		} else {
			host = "cloud.yugabyte.com"
		}
		viper.GetViper().Set("host", &host)
		err = viper.WriteConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Fprintln(os.Stdout, "No config was found a new one will be created.")
				//Try to create the file
				err = viper.SafeWriteConfig()
				if err != nil {
					logrus.Fatalf("Error when writing new config file: %v", err)

				}
			} else {
				logrus.Fatalf("Error when writing config file: %v", err)
			}
		}
		logrus.Infof("Configuration file '%v' sucessfully updated.", viper.GetViper().ConfigFileUsed())
	},
}
