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

package endpoint

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"

	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var listEndpointCmd = &cobra.Command{
	Use:   "list",
	Short: "List network endpoints for a cluster",
	Long:  "List network endpoints for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s\n", ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterEndpoints, _, err := authApi.GetEndpointsForClusterByName(clusterName)
		if err != nil {
			logrus.Fatalf("Could not get cluster data: %s\n", ybmAuthClient.GetApiErrorDetails(err))
		}

		region, _ := cmd.Flags().GetString("region")
		if region != "" {
			clusterEndpoints = util.Filter(clusterEndpoints, func(endpoint ybmclient.Endpoint) bool {
				return endpoint.GetRegion() == region
			})
		}

		accessibility, _ := cmd.Flags().GetString("accessibility")
		if accessibility != "" {
			clusterEndpoints = util.Filter(clusterEndpoints, func(endpoint ybmclient.Endpoint) bool {
				return string(endpoint.GetAccessibilityType()) == accessibility
			})
		}

		if len(clusterEndpoints) == 0 {
			logrus.Fatalf("No endpoints found\n")
		}

		providers, err := authApi.ExtractProviderFromClusterName(clusterName)
		if err != nil {
			logrus.Fatalf("could not fetch provider for cluster %s : %s\n", clusterName, ybmAuthClient.GetApiErrorDetails(err))
		}

		endpointsCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewEndpointFormat(viper.GetString("output")),
		}
		formatter.EndpointWrite(endpointsCtx, clusterEndpoints, providers)
	},
}

func init() {
	EndpointCmd.AddCommand(listEndpointCmd)
	listEndpointCmd.Flags().String("accessibility-type", "", "[OPTIONAL] Accessibility of the endpoint. Valid options are PUBLIC, PRIVATE and PRIVATE_SERVICE_ENDPOINT.")
	listEndpointCmd.Flags().String("region", "", "[OPTIONAL] The region of the endpoint.")
}
