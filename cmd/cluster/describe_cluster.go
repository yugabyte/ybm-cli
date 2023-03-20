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

var describeClusterCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a cluster",
	Long:  "Describe a cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterListRequest := authApi.ListClusters()
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterListRequest = clusterListRequest.Name(clusterName)

		resp, r, err := clusterListRequest.Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) > 0 && viper.GetString("output") == "table" {
			fullClusterContext := *formatter.NewFullClusterContext()
			fullClusterContext.Output = os.Stdout
			fullClusterContext.Format = formatter.NewFullClusterFormat(viper.GetString("output"))
			fullClusterContext.SetFullCluster(*authApi, resp.GetData()[0])
			fullClusterContext.Write()
			return
		}
		fmt.Println("No clusters found")
	},
}

func init() {
	ClusterCmd.AddCommand(describeClusterCmd)
	describeClusterCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster to get details.")
	describeClusterCmd.MarkFlagRequired("cluster-name")
}
