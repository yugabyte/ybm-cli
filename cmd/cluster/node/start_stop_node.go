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

package node

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var stopNodeCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a cluster node",
	Long:  "Stop a cluster node",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}
		nodeName, _ := cmd.Flags().GetString("node-name")
		nodeOpRequest := ybmclient.NewNodeOpRequest(nodeName, ybmclient.NODEOPENUM_STOP)

		resp, err := authApi.PerformNodeOperation(clusterId).NodeOpRequest(*nodeOpRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The node %s is being stopped", formatter.Colorize(nodeName, formatter.GREEN_COLOR))
		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_STOP_NODE, []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The node %s has been stopped\n", formatter.Colorize(nodeName, formatter.GREEN_COLOR))
		} else {
			fmt.Println(msg)
		}

	},
}

var startNodeCmd = &cobra.Command{
	Use:   "start",
	Short: "start a cluster node",
	Long:  "start a cluster node",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}
		nodeName, _ := cmd.Flags().GetString("node-name")
		nodeOpRequest := ybmclient.NewNodeOpRequest(nodeName, ybmclient.NODEOPENUM_START)

		resp, err := authApi.PerformNodeOperation(clusterId).NodeOpRequest(*nodeOpRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The node %s is being started", formatter.Colorize(nodeName, formatter.GREEN_COLOR))
		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_START_NODE, []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The node %s has been started\n", formatter.Colorize(nodeName, formatter.GREEN_COLOR))
		} else {
			fmt.Println(msg)
		}

	},
}

func init() {
	util.AddCommandIfFeatureFlag(NodeCmd, stopNodeCmd, util.NODE_OP)
	stopNodeCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster to get details.")
	stopNodeCmd.MarkFlagRequired("cluster-name")
	stopNodeCmd.Flags().String("node-name", "", "[REQUIRED] The name of the node to stop.")
	stopNodeCmd.MarkFlagRequired("node-name")

	util.AddCommandIfFeatureFlag(NodeCmd, startNodeCmd, util.NODE_OP)
	startNodeCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster to get details.")
	startNodeCmd.MarkFlagRequired("cluster-name")
	startNodeCmd.Flags().String("node-name", "", "[REQUIRED] The name of the node to stop.")
	startNodeCmd.MarkFlagRequired("node-name")
}
