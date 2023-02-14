// Copyright (c) YugaByte, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

var PseCmd = &cobra.Command{
	Use:   "private-service-endpoint",
	Short: "Manage Private Service Endpoints for a Cluster",
	Long:  "Manage Private Service Endpoints of a Cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getPseCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Private Service Endpoints for a Cluster",
	Long:  "Get Private Service Endpoints for a Cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
	},
}

var createPseCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Private Service Endpoints for a Cluster",
	Long:  "Create Private Service Endpoints for a Cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

	},
}

var updatePseCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Private Service Endpoints for a Cluster",
	Long:  "Update Private Service Endpoints for a Cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

	},
}

var deletePseCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete network allow list from YugabyteDB Managed",
	Long:  "Delete network allow list from YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

	},
}

func init() {
	PseCmd.AddCommand(getPseCmd)
	getPseCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	getPseCmd.MarkFlagRequired("cluster-name")
	getPseCmd.Flags().String("region", "", "[OPTIONAL] The region of the Private Service Endpoint.")

	PseCmd.AddCommand(createPseCmd)
	createPseCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	createPseCmd.MarkFlagRequired("cluster-name")
	createPseCmd.Flags().StringArray("region-info", []string{}, "[REQUIRED] The region information of the Private Service Endpoints. region-info=region=<region-name>,security-principles=<sp1>,>sp2>,<sp3>")
	createPseCmd.MarkFlagRequired("region-info")

	PseCmd.AddCommand(updatePseCmd)
	updatePseCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	updatePseCmd.MarkFlagRequired("cluster-name")
	updatePseCmd.Flags().StringArray("region-info", []string{}, "[REQUIRED] The region information of the Private Service Endpoints. region-info=region=<region-name>,security-principles=<sp1>,>sp2>,<sp3>")
	updatePseCmd.MarkFlagRequired("region-info")

	PseCmd.AddCommand(deletePseCmd)
	deletePseCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	deletePseCmd.MarkFlagRequired("cluster-name")
	deletePseCmd.Flags().String("region", "", "[REQUIRED] The region of the Private Service Endpoint.")
	deletePseCmd.MarkFlagRequired("region")
}
