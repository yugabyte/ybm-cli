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
	Short: "List endpoints for a cluster",
	Long:  "List endpoints for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterListRequest := authApi.ListClusters()
		// if user filters by name, add it to the request
		clusterListRequest = clusterListRequest.Name(clusterName)

		resp, r, err := clusterListRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) == 0 {
			logrus.Fatalf("Cluster not found")
		}

		clusterEndpoints := resp.GetData()[0].Info.ClusterEndpoints

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
			logrus.Fatalf("No endpoints found")
		}

		endpointsCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewEndpointFormat(viper.GetString("output")),
		}
		formatter.EndpointWrite(endpointsCtx, clusterEndpoints)
	},
}

func init() {
	EndpointCmd.AddCommand(listEndpointCmd)
	listEndpointCmd.Flags().String("accessibility", "", "[OPTIONAL] Accessibility of the endpoint")
	listEndpointCmd.Flags().String("region", "", "[OPTIONAL] Region of the endpoint")
}
