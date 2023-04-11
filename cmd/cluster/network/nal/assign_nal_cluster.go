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

package nal

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// assignClusterCmd represents the cluster command
var assignClusterCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign resources(e.g. network allow lists) to clusters",
	Long:  "Assign resources(e.g. network allow lists) to clusters",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		newNetworkAllowListName, _ := cmd.Flags().GetString("network-allow-list")
		newNetworkAllowListId, err := authApi.GetNetworkAllowListIdByName(newNetworkAllowListName)
		if err != nil {
			logrus.Fatal(err)
		}

		networkAllowListListResp, r, err := authApi.ListClusterNetworkAllowLists(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		allowListIds := []string{}
		allowListIds = append(allowListIds, newNetworkAllowListId)
		for _, nal := range networkAllowListListResp.Data {
			allowListIds = append(allowListIds, nal.Info.GetId())
		}

		_, r, err = authApi.EditClusterNetworkAllowLists(clusterId, allowListIds).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The network allow list %s is being assigned to the cluster %s", formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_EDIT_ALLOW_LIST, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The network allow list %s has been assigned to the cluster %s\n", formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		} else {
			fmt.Println(msg)
		}
	},
}

func init() {
	AllowListCmd.AddCommand(assignClusterCmd)
	assignClusterCmd.Flags().String("network-allow-list", "", "[REQUIRED] The name of the network allow list to be assigned.")
	assignClusterCmd.MarkFlagRequired("network-allow-list")
}
