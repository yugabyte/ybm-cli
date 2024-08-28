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

package connection_pooling

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

var ClusterName string

var ConnectionPoolingCmd = &cobra.Command{
	Use:   "connection-pooling",
	Short: "Manage Connection Pooling",
	Long:  "Manage Connection Pooling",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var enableConnectionPoolingCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable Connection Pooling",
	Long:  "Enable Connection Pooling",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		connectionPoolingOpSpec := ybmclient.NewConnectionPoolingOpSpec(ybmclient.CONNECTIONPOOLINGOPENUM_ENABLE)

		resp, err := authApi.PerformConnectionPoolingOperation(clusterId).ConnectionPoolingOpSpec(*connectionPoolingOpSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("Connection Pooling for cluster %s is being enabled", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_ENABLE_CONNECTION_POOLING, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("Connection Pooling has been enabled on cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		} else {
			fmt.Println(msg)
		}
	},
}

var disableConnectionPoolingCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable Connection Pooling",
	Long:  "Disable Connection Pooling",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		connectionPoolingOpSpec := ybmclient.NewConnectionPoolingOpSpec(ybmclient.CONNECTIONPOOLINGOPENUM_DISABLE)

		resp, err := authApi.PerformConnectionPoolingOperation(clusterId).ConnectionPoolingOpSpec(*connectionPoolingOpSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("Connection Pooling for cluster %s is being disabled", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_DISABLE_CONNECTION_POOLING, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("Connection Pooling has been disabled on cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		} else {
			fmt.Println(msg)
		}
	},
}

func init() {
	util.AddCommandIfFeatureFlag(ConnectionPoolingCmd, enableConnectionPoolingCmd, util.CONNECTION_POOLING)

	util.AddCommandIfFeatureFlag(ConnectionPoolingCmd, disableConnectionPoolingCmd, util.CONNECTION_POOLING)
}