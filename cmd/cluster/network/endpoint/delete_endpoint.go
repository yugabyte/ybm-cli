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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var deleteEndpointCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a private service endpoint",
	Long:  `Delete a private service endpoint.`,
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
			logrus.Debugln("Endpoint is a private service endpoint, attempting to delete")
			r, err := authApi.DeletePrivateServiceEndpoint(clusterId, endpointId).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Error when calling `ClusterApi.DeletePrivateServiceEndpoint`: %s\n", ybmAuthClient.GetApiErrorDetails(err))
			}

			msg := fmt.Sprintf("Deleted endpoint %s", endpointId)
			fmt.Println(msg)

		default:
			logrus.Fatalf("Endpoint is not a private service endpoint. Only private service endpoints are currently supported.\n")
		}

	},
}

func init() {
	EndpointCmd.AddCommand(deleteEndpointCmd)
	deleteEndpointCmd.Flags().String("endpoint-id", "", "[REQUIRED] THe ID of the endpoint")
	deleteEndpointCmd.MarkFlagRequired("endpoint-id")
}
