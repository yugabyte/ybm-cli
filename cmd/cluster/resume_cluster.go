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
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// resumeClusterCmd represents the cluster command
var resumeClusterCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a cluster in YugabyteDB Managed",
	Long:  "Resume a cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, r, err := authApi.ResumeCluster(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.ResumeCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		clusterData := []ybmclient.ClusterData{resp.GetData()}

		msg := fmt.Sprintf("The cluster %s is being resumed", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "RESUME_CLUSTER", []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The cluster %s has been resumed\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.ListClusters().Name(clusterName).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
			}
			clusterData = respC.GetData()
		} else {
			fmt.Println(msg)
		}

		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}

		formatter.ClusterWrite(clustersCtx, clusterData)
	},
}

func init() {
	ClusterCmd.AddCommand(resumeClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resumeClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resumeClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	resumeClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be resumed")
	resumeClusterCmd.MarkFlagRequired("cluster-name")
}
