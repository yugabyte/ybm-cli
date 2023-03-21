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
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var createEndpointCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new endpoint",
	Long:  `Create a new endpoint`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterData := getCluster(cmd, authApi)
		accessibilityType, _ := cmd.Flags().GetString("accessibility-type")
		reg, _ := cmd.Flags().GetString("region")

		switch accessibilityType {
		case string(ybmclient.ACCESSIBILITYTYPE_PRIVATE_SERVICE_ENDPOINT):
			logrus.Debugln("Endpoint is a private service endpoint")
			if cmd.Flags().Changed("security-principals") {
				logrus.Debugln("Security principals are set, attempting to create")
			} else {
				logrus.Fatalln("Security principals are not set and are mandatory for Private Service Endpoints.")
			}
			securityPrincipalsString, _ := cmd.Flags().GetString("security-principals")
			securityPrincipalsList := strings.Split(securityPrincipalsString, ",")

			allClusterRegions := clusterData.Info.ClusterRegionInfoDetails
			desiredRegions := util.Filter(allClusterRegions, func(region ybmclient.ClusterRegionInfoDetails) bool {
				return region.Region == reg
			})

			if len(desiredRegions) == 0 {
				logrus.Fatalf("No region found for cluster %s with name %s", clusterData.Spec.Name, reg)
			}
			if len(desiredRegions) > 1 {
				logrus.Fatalf("Multiple regions found for cluster %s with name %s", clusterData.Spec.Name, reg)
			}

			regionArnMap := make(map[string][]string)
			regionArnMap[desiredRegions[0].Id] = securityPrincipalsList
			createPseSpec := authApi.CreatePrivateServiceEndpointSpec(regionArnMap)

			createPseRequest := authApi.CreatePrivateServiceEndpoint(clusterData.Info.Id)
			createPseRequest.PrivateServiceEndpointSpec(createPseSpec[0])

			createResp, r, err := createPseRequest.Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Could not create private service endpoint: %s", err.Error())
			}

			psEps := util.Filter(createResp.GetData(), func(ep ybmclient.PrivateServiceEndpointRegionData) bool {
				return ep.GetSpec().ClusterRegionInfoId == desiredRegions[0].Id
			})

			if len(psEps) == 0 {
				logrus.Fatalf("No private service endpoint found for cluster %s with region %s", clusterData.Spec.Name, reg)
			}

			msg := fmt.Sprintf("Created private service endpoint in region %v", reg)
			fmt.Println(msg)

		default:
			logrus.Fatalf("Endpoint is not a private service endpoint. Only private service endpoints are currently supported.")

		}

	},
}

func init() {
	EndpointCmd.AddCommand(createEndpointCmd)
	createEndpointCmd.Flags().String("accessibility-type", "", "[REQUIRED] The accessibility of the endpoint.")
	createEndpointCmd.MarkFlagRequired("accessibility-type")
	createEndpointCmd.Flags().String("region", "", "[REQUIRED] The region of the endpoint.")
	createEndpointCmd.MarkFlagRequired("region")
	createEndpointCmd.Flags().String("security-principals", "", "[OPTIONAL] The security principals of the endpoint.")
}
