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

var getInstanceTypesCmd = &cobra.Command{
	Use:   "instance_types",
	Short: "Get Instance Types in YugabyteDB Managed",
	Long:  `Get Instance Types in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		cloudType, _ := cmd.Flags().GetString("cloud-type")
		cloudRegion, _ := cmd.Flags().GetString("cloud-region")
		tier, _ := cmd.Flags().GetString("tier")
		showDisabled, _ := cmd.Flags().GetBool("show-disabled")
		instanceTypesResp, resp, err := authApi.GetSupportedInstanceTypes(cloudType, tier, cloudRegion).ShowDisabled(showDisabled).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.GetSupportedInstanceTypes`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", resp)
			return
		}

		prettyPrintJson(instanceTypesResp)
	},
}

func init() {
	getCmd.AddCommand(getInstanceTypesCmd)
	getInstanceTypesCmd.Flags().String("cloud-type", "", "The cloud provider for which the regions have to be fetched. AWS or GCP.")
	getInstanceTypesCmd.MarkFlagRequired("cloud-type")
	getInstanceTypesCmd.Flags().String("cloud-region", "", "Cloud Region.")
	getInstanceTypesCmd.MarkFlagRequired("cloud-region")
	getInstanceTypesCmd.Flags().String("tier", "PAID", "Tier. FREE or PAID. Default: PAID")
	getInstanceTypesCmd.Flags().Bool("show-disabled", false, "Whether to show disabled instance types. true or false. Default: false")

}
