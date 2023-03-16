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

package namespace

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

var NamespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Manage namespaces",
	Long:  "Manage namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listNamespaceCmd = &cobra.Command{
	Use:   "list",
	Short: "List namespaces",
	Long:  `List namespaces`,
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

		namespacesResp, resp, err := authApi.GetClusterNamespaces(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		namespaceData := namespacesResp.GetData()
		fmt.Println(namespaceData)

		// namespaceCtx := formatter.Context{
		// 	Output: os.Stdout,
		// 	Format: formatter.namespaceFormat(viper.GetString("output")),
		// }

		// formatter.CloudRegionWrite(namespaceCtx, namespaceData)
	},
}

func init() {
	NamespaceCmd.AddCommand(listNamespaceCmd)

	listNamespaceCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	listNamespaceCmd.MarkFlagRequired("cluster-name")
}
