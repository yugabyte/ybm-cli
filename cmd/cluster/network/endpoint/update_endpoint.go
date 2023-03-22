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
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var updateEndpointCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a network endpoint for a cluster",
	Long:  `Update a network endpoint for a cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s\n", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		endpointId, _ := cmd.Flags().GetString("endpoint-id")
		clusterEndpoint, clusterId, err := authApi.GetEndpointByIdForClusterByName(clusterName, endpointId)
		if err != nil {
			logrus.Fatalf("Error when calling `ClusterApi.GetEndpointByIdForClusterByName`: %s\n", ybmAuthClient.GetApiErrorDetails(err))
		}

		// We currently support fetching just Private Service Endpoints
		switch clusterEndpoint.GetAccessibilityType() {

		case ybmclient.ACCESSIBILITYTYPE_PRIVATE_SERVICE_ENDPOINT:
			if !cmd.Flags().Changed("security-principals") {
				logrus.Fatalf("security-principals is required for private service endpoints\n")
			}

			pseGetResponse, r, err := authApi.GetPrivateServiceEndpoint(clusterId, endpointId).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Error when calling `ClusterApi.GetPrivateServiceEndpoint`: %s\n", ybmAuthClient.GetApiErrorDetails(err))
			}

			securityPrincipalsString, _ := cmd.Flags().GetString("security-principals")
			securityPrincipalsList := strings.Split(securityPrincipalsString, ",")

			regionArnMap := make(map[string][]string)
			regionArnMap[pseGetResponse.Data.Spec.ClusterRegionInfoId] = securityPrincipalsList

			// we create a spec that has a single element
			pseSpec := authApi.CreatePrivateServiceEndpointRegionSpec(regionArnMap)

			// we pass the only element in the spec to the update endpoint call
			updateResp, r, err := authApi.EditPrivateServiceEndpoint(clusterId, endpointId).PrivateServiceEndpointRegionSpec(pseSpec[0]).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Error when calling `ClusterApi.EditPrivateServiceEndpoint`: %s\n", ybmAuthClient.GetApiErrorDetails(err))
			}

			msg := fmt.Sprintf("Updated endpoint %s", updateResp.Data.Info.Id)
			fmt.Println(msg)

		default:
			logrus.Fatalf("Endpoint is not a private service endpoint. Only private service endpoints are currently supported.\n")
		}

	},
}

func init() {
	EndpointCmd.AddCommand(updateEndpointCmd)
	updateEndpointCmd.Flags().String("endpoint-id", "", "[REQUIRED] The ID of the endpoint")
	updateEndpointCmd.MarkFlagRequired("endpoint-id")
	updateEndpointCmd.Flags().String("security-principals", "", "[OPTIONAL] The list of security principals that have access to this endpoint. Required for private service endpoints.  Accepts a comma separated list. E.g.: `arn:aws:iam::account_id1:root,arn:aws:iam::account_id2:root`")
}
