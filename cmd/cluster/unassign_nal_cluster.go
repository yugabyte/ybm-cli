// Copyright (c) YugaByte, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

// unassignClusterCmd represents the cluster command
var unassignClusterCmd = &cobra.Command{
	Use:   "unassign",
	Short: "Unassign resources(e.g. network allow lists) to clusters in YugabyteDB Managed",
	Long:  "Unassign resources(e.g. network allow lists) to clusters in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}
		newNetworkAllowListName, _ := cmd.Flags().GetString("network-allow-list")
		newNetworkAllowListId, err := authApi.GetNetworkAllowListIdByName(newNetworkAllowListName)
		if err != nil {
			logrus.Error(err)
			return
		}

		networkAllowListListResp, r, err := authApi.ListClusterNetworkAllowLists(clusterId).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.ListClusterNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		allowListIds := []string{}
		nalFound := false
		for _, nal := range networkAllowListListResp.Data {
			nalId := nal.Info.GetId()
			if nalId == newNetworkAllowListId {
				nalFound = true
			} else {
				allowListIds = append(allowListIds, nalId)
			}
		}
		if !nalFound {
			logrus.Errorf("The allow list %s is not associated with the cluster %s", formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))
			return
		}

		_, r, err = authApi.EditClusterNetworkAllowLists(clusterId, allowListIds).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.EditClusterNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		msg := fmt.Sprintf("The cluster %s is being unassigned the network allow list %s", formatter.Colorize(clusterName, formatter.GREEN_COLOR), formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, "CLUSTER", "EDIT_ALLOW_LIST", []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
			}
			fmt.Printf("The cluster %s has been unassigned the network allow list %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR), formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR))

		} else {
			fmt.Println(msg)
		}

	},
}

func init() {
	ClusterCmd.AddCommand(unassignClusterCmd)
	unassignClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be unassignd")
	unassignClusterCmd.MarkFlagRequired("cluster-name")
	unassignClusterCmd.Flags().String("network-allow-list", "", "The name of the network allow list to be unassignd")
	// Marked as required for now since as of now network allow list is the only resource that can be unassigned
	unassignClusterCmd.MarkFlagRequired("network-allow-list")
}
