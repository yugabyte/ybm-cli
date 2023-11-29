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

package policy

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var PolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage backup policy of a cluster",
	Long:  "Manage backup policy of a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listPolicyCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup policies",
	Long:  "List backup policies for cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		listBackupPoliciesRequest := authApi.ListBackupPolicies(clusterID)

		resp, r, err := listBackupPoliciesRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		policyCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewBackupPolicyFormat(viper.GetString("output")),
		}
		if len(resp.GetData()) < 1 {
			logrus.Println("No backup policies found for the given cluster\n")
			return
		}
		formatter.BackupPolicyListWrite(policyCtx, resp.GetData())
	},
}

func init() {
	PolicyCmd.AddCommand(listPolicyCmd)
	listPolicyCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster to fetch backup policiess.")
	listPolicyCmd.MarkFlagRequired("cluster-name")
}
