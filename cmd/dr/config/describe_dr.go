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

package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var describeDrCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe DR",
	Long:  `Describe DR`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		drName, _ := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatalf("Could not get cluster data: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		drInfo, err := authApi.GetDrDetailsByName(drName)
		if err != nil {
			logrus.Fatal(err)
		}
		drId := drInfo.GetId()
		clusterId := drInfo.GetSourceClusterId()
		drResp, r, err := authApi.GetXClusterDr(clusterId, drId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		drCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDrFormat(viper.GetString("output")),
		}

		formatter.DrWrite(drCtx, []ybmclient.XClusterDrData{drResp.GetData()}, *authApi)
	},
}

func init() {
	ConfigCmd.AddCommand(describeDrCmd)
	describeDrCmd.Flags().String("config", "", "[REQUIRED] Name of the DR configuration.")
	describeDrCmd.MarkFlagRequired("config")
}
