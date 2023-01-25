/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

var getCloudRegionsCmd = &cobra.Command{
	Use:   "cloud_regions",
	Short: "Get Cloud Regions in YugabyteDB Managed",
	Long:  `Get Cloud Regions in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		cloudType, _ := cmd.Flags().GetString("cloud-type")
		cloudRegionsResp, resp, err := authApi.GetSupportedCloudRegions().Cloud(cloudType).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.GetSupportedCloudRegions`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", resp)
			return
		}

		prettyPrintJson(cloudRegionsResp)
	},
}

func init() {
	getCmd.AddCommand(getCloudRegionsCmd)
	getCloudRegionsCmd.Flags().String("cloud-type", "", "The cloud provider for which the regions have to be fetched. AWS or GCP.")
	getCloudRegionsCmd.MarkFlagRequired("cloud-type")

}
