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

var getInstanceTypesCmd = &cobra.Command{
	Use:   "describe-instances",
	Short: "Get Instance Types in YugabyteDB Managed",
	Long:  `Get Instance Types in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
		cloudRegion, _ := cmd.Flags().GetString("region")
		tier, _ := cmd.Flags().GetString("tier")
		showDisabled, _ := cmd.Flags().GetBool("show-disabled")
		instanceTypesResp, resp, err := authApi.GetSupportedInstanceTypes(cloudProvider, tier, cloudRegion).ShowDisabled(showDisabled).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf("Error when calling `ClusterApi.GetSupportedInstanceTypes`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		instanceTypeData := instanceTypesResp.GetData()[cloudRegion]

		instanceTypeCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewInstanceTypeFormat(viper.GetString("output")),
		}

		formatter.InstanceTypeWrite(instanceTypeCtx, instanceTypeData)
	},
}

func init() {
	ClusterCmd.AddCommand(getInstanceTypesCmd)
	getInstanceTypesCmd.Flags().String("cloud-provider", "", "The cloud provider for which the regions have to be fetched. AWS or GCP.")
	getInstanceTypesCmd.MarkFlagRequired("cloud-provider")
	getInstanceTypesCmd.Flags().String("region", "", "The region in the cloud provider for which the instance types have to fetched.")
	getInstanceTypesCmd.MarkFlagRequired("region")
	getInstanceTypesCmd.Flags().String("tier", "PAID", "Tier. FREE or PAID. Default: PAID")
	getInstanceTypesCmd.Flags().Bool("show-disabled", false, "Whether to show disabled instance types. true or false. Default: false")

}
