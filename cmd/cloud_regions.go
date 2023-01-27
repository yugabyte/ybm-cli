/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var getCloudRegionsCmd = &cobra.Command{
	Use:   "cloud_regions",
	Short: "Get Cloud Regions in YugabyteDB Managed",
	Long:  `Get Cloud Regions in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
		cloudRegionsResp, resp, err := authApi.GetSupportedCloudRegions().Cloud(cloudProvider).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.GetSupportedCloudRegions`: %v", err)
			logrus.Debugf("Full HTTP response: %v", resp)
			return
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
	getCmd.AddCommand(getCloudRegionsCmd)
	getCloudRegionsCmd.Flags().String("cloud-provider", "", "The cloud provider for which the regions have to be fetched. AWS or GCP.")
	getCloudRegionsCmd.MarkFlagRequired("cloud-provider")

}
