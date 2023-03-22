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

package region

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var CloudRegionsCmd = &cobra.Command{
	Use:   "region",
	Short: "Manage cloud regions",
	Long:  "Manage cloud regions for your YBM clusters",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listCloudRegionsCmd = &cobra.Command{
	Use:   "list",
	Short: "List Cloud Provider Regions",
	Long:  `List Cloud Provider Regions`,
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
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
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
	CloudRegionsCmd.AddCommand(listCloudRegionsCmd)

	listCloudRegionsCmd.Flags().String("cloud-provider", "", "[REQUIRED] The cloud provider for which the regions have to be fetched. AWS or GCP.")
	listCloudRegionsCmd.MarkFlagRequired("cloud-provider")
}
