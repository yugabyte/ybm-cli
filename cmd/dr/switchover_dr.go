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

package dr

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var switchoverDrCmd = &cobra.Command{
	Use:   "switchover",
	Short: "Switchover DR for a cluster",
	Long:  `Switchover DR for a cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		drName, _ := cmd.Flags().GetString("dr-name")
		if err != nil {
			logrus.Fatalf("Could not get cluster data: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		drInfo, err := authApi.GetDrDetailsByName(drName)
		if err != nil {
			logrus.Fatal(err)
		}
		drId := drInfo.GetId()
		clusterId := drInfo.GetSourceClusterId()

		response, err := authApi.SwitchoverXClusterDr(clusterId, drId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("Switchover is in progress for the DR %s ", formatter.Colorize(drName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_DR_SWITCHOVER, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("Switchover for DR config %s is successful\n", formatter.Colorize(drName, formatter.GREEN_COLOR))

			drGetResp, r, err := authApi.GetXClusterDr(clusterId, drId).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			drCtx := formatter.Context{
				Output: os.Stdout,
				Format: formatter.NewDrFormat(viper.GetString("output")),
			}

			formatter.DrWrite(drCtx, []ybmclient.XClusterDrData{drGetResp.GetData()}, *authApi)
		} else {
			fmt.Println(msg)
		}

	},
}

func init() {
	DrCmd.AddCommand(switchoverDrCmd)
	switchoverDrCmd.Flags().String("dr-name", "", "[REQUIRED] Name of the DR configuration.")
	switchoverDrCmd.MarkFlagRequired("dr-name")
}
