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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var getCloudRegionsCmd = &cobra.Command{
	Use:   "describe-regions",
	Short: "Get Cloud Regions in YugabyteDB Managed",
	Long:  `Get Cloud Regions in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
		cloudRegionsResp, resp, err := authApi.GetSupportedCloudRegions().Cloud(cloudProvider).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf("Error when calling `ClusterApi.GetSupportedCloudRegions`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		cloudRegionData := cloudRegionsResp.GetData()

		cloudRegionCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewCloudRegionFormat(viper.GetString("output")),
		}

		formatter.CloudRegionWrite(cloudRegionCtx, cloudRegionData)
	},
}

func init() {
	ClusterCmd.AddCommand(getCloudRegionsCmd)
	getCloudRegionsCmd.Flags().String("cloud-provider", "", "The cloud provider for which the regions have to be fetched. AWS or GCP.")
	getCloudRegionsCmd.MarkFlagRequired("cloud-provider")

}
