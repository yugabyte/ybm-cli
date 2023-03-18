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
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var describeEndpointCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a network endpoint for a cluster",
	Long:  `Describe a network endpoint for a cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterListRequest := authApi.ListClusters()
		// user filters by name, add it to the request
		clusterListRequest = clusterListRequest.Name(clusterName)

		resp, r, err := clusterListRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) == 0 {
			logrus.Fatalf("Cluster not found")
		}

		clusterId := resp.GetData()[0].Info.Id
		clusterEndpoints := resp.GetData()[0].Info.ClusterEndpoints
		jsonEndpoints, _ := json.Marshal(clusterEndpoints)
		logrus.Debugf("Found endpoints: %v", string(jsonEndpoints))
		endpointId, _ := cmd.Flags().GetString("endpoint-id")
		clusterEndpoints = util.Filter(clusterEndpoints, func(endpoint ybmclient.Endpoint) bool {
			return endpoint.Id == endpointId || endpoint.GetPseId() == endpointId
		})

		if len(clusterEndpoints) == 0 {
			logrus.Fatalf("Endpoint not found")
		}

		// We currently support fetching just Private Service Endpoints
		switch clusterEndpoints[0].GetAccessibilityType() {

		case ybmclient.ACCESSIBILITYTYPE_PRIVATE_SERVICE_ENDPOINT:
			logrus.Debugln("Endpoint is a private service endpoint, getting more data")
			pseGetResponse, r, err := authApi.GetPrivateServiceEndpoint(clusterId, endpointId).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Error when calling `ClusterApi.GetPrivateServiceEndpoint`: %s", ybmAuthClient.GetApiErrorDetails(err))
			}
			if viper.GetString("output") == "table" {
				psEndpointContext := *formatter.NewPSEndpointContext()
				psEndpointContext.Output = os.Stdout
				psEndpointContext.Format = formatter.NewPSEndpointFormat(viper.GetString("output"))
				psEndpointContext.SetFullPSEndpoint(*authApi, pseGetResponse.GetData(), clusterEndpoints[0])
				psEndpointContext.Write()
				return
			}

			psEndpointContext := formatter.Context{
				Output: os.Stdout,
				Format: formatter.NewPSEndpointFormat(viper.GetString("output")),
			}
			formatter.PSEndpointWrite(psEndpointContext, pseGetResponse.GetData(), clusterEndpoints[0])

			break

		default:
			logrus.Fatalf("Endpoint is not a private service endpoint. Only private service endpoints are currently supported.")
		}
	},
}

func init() {
	EndpointCmd.AddCommand(describeEndpointCmd)
	describeEndpointCmd.Flags().String("endpoint-id", "", "[REQUIRED] The ID of the endpoint")
}
