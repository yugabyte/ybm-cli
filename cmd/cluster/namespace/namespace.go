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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var NamespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Manage Cluster Namespaces",
	Long:  "Manage Cluster namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listNamespaceCmd = &cobra.Command{
	Use:   "list",
	Short: "List namespaces for a cluster",
	Long:  "List namespaces on your YugabyteDB Aeon cluster",
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

		resp, r, err := authApi.GetClusterNamespaces(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) == 0 {
			logrus.Fatalf("No namespaces found.\n")
		}

		namespaceCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewNamespaceFormat(viper.GetString("output")),
		}
		formatter.NamespaceWrite(namespaceCtx, resp.GetData())
	},
}

func init() {
	NamespaceCmd.AddCommand(listNamespaceCmd)
}
