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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func getCluster(cmd *cobra.Command, authApi *ybmAuthClient.AuthApiClient) ybmclient.ClusterData {
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

	return resp.GetData()[0]
}

func getEndpoints(cmd *cobra.Command, authApi *ybmAuthClient.AuthApiClient) ([]ybmclient.Endpoint, string) {
	clusterData := getCluster(cmd, authApi)

	clusterId := clusterData.Info.Id
	clusterEndpoints := clusterData.Info.ClusterEndpoints
	jsonEndpoints, _ := json.Marshal(clusterEndpoints)
	logrus.Debugf("Found endpoints: %v", string(jsonEndpoints))

	return clusterEndpoints, clusterId
}

func getEndpointById(cmd *cobra.Command, authApi *ybmAuthClient.AuthApiClient) ([]ybmclient.Endpoint, string, string) {
	clusterEndpoints, clusterId := getEndpoints(cmd, authApi)
	endpointId, _ := cmd.Flags().GetString("endpoint-id")
	clusterEndpoints = util.Filter(clusterEndpoints, func(endpoint ybmclient.Endpoint) bool {
		return endpoint.Id == endpointId || endpoint.GetPseId() == endpointId
	})

	if len(clusterEndpoints) == 0 {
		logrus.Fatalf("Endpoint not found")
	}

	return clusterEndpoints, clusterId, endpointId
}
