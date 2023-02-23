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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

// getClusterCmd represents the cluster command
var getClusterCmd = &cobra.Command{
	Use:   "get",
	Short: "Get clusters",
	Long:  "Get clusters",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterListRequest := authApi.ListClusters()
		isGetByName := false
		// if user filters by name, add it to the request
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName != "" {
			clusterListRequest = clusterListRequest.Name(clusterName)
			isGetByName = true
		}

		resp, r, err := clusterListRequest.Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		if isGetByName && len(resp.GetData()) > 0 {
			fullClusterContext := *formatter.NewFullClusterContext()
			fullClusterContext.Output = os.Stdout
			fullClusterContext.Format = formatter.NewFullClusterFormat(viper.GetString("output"))
			fullClusterContext.SetFullCluster(*authApi, resp.GetData()[0])
			fullClusterContext.Write()
			return
		}
		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}
		formatter.ClusterWrite(clustersCtx, resp.GetData())
	},
}

func init() {
	ClusterCmd.AddCommand(getClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	getClusterCmd.Flags().String("cluster-name", "", "[OPTIONAL] The name of the cluster to get details.")
}
