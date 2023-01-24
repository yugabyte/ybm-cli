/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// vpcCmd represents the vpc command
var getVpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Get VPCs in YugabyteDB Managed",
	Long:  "Get VPCs in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		vpcName, _ := cmd.Flags().GetString("name")

		vpcListRequest := authApi.ListSingleTenantVpcsByName(vpcName)
		resp, r, err := vpcListRequest.Execute()

		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.ListSingleTenantVpcs`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}
		// response from `ListClusters`: ClusterListResponse
		vpcCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewVPCFormat(viper.GetString("output")),
		}

		formatter.VPCWrite(vpcCtx, resp.GetData())
	},
}

var createRegions []string
var createCidrs []string
var createVpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Create a VPC in YugabyteDB Managed",
	Long:  "Create a VPC in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		// Validations
		if !cmd.Flags().Changed("global-cidr") && !cmd.Flags().Changed("cidr") {
			fmt.Fprintln(os.Stderr, "Either global-cidr or cidr must be specified")
			os.Exit(1)
		}

		if len(createRegions) != len(createCidrs) {
			fmt.Fprintln(os.Stderr, "Number of regions and cidrs must be equal")
			os.Exit(1)
		}

		vpcName, _ := cmd.Flags().GetString("name")
		cloud, _ := cmd.Flags().GetString("cloud")
		globalCidrRange, _ := cmd.Flags().GetString("global-cidr")

		// global CIDR only works with GCP
		if cloud != "GCP" && cmd.Flags().Changed("global-cidr") {
			fmt.Fprintln(os.Stderr, "global-cidr is only supported for GCP")
			os.Exit(1)
		}

		// If non-global CIDR, validate that there are different regions specified
		regionMap := map[string]int{}
		vpcRegionSpec := []ybmclient.VpcRegionSpec{}

		if cmd.Flags().Changed("cidr") {
			for index, region := range createRegions {
				cidr := createCidrs[index]
				spec := *ybmclient.NewVpcRegionSpecWithDefaults()
				regionMap[region] = index
				spec.SetRegion(region)
				spec.SetCidr(cidr)
				vpcRegionSpec = append(vpcRegionSpec, spec)
			}
			if len(regionMap) != len(createRegions) {
				fmt.Fprintln(os.Stderr, "Regions must be unique")
				os.Exit(1)
			}
		}

		vpcSpec := *ybmclient.NewSingleTenantVpcSpec(vpcName, ybmclient.CloudEnum(cloud), vpcRegionSpec)
		if cmd.Flags().Changed("global-cidr") {
			vpcSpec.SetParentCidr(globalCidrRange)
		}
		vpcRequest := *ybmclient.NewSingleTenantVpcRequest(vpcSpec)

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		resp, r, err := authApi.CreateVpc().SingleTenantVpcRequest(vpcRequest).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.CreateVpc`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var deleteVpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Delete a VPC in YugabyteDB Managed",
	Long:  "Delete a VPC in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		vpcName, _ := cmd.Flags().GetString("name")
		if vpcName == "" {
			logrus.Error("name field is required")
			os.Exit(1)
		}
		vpcId, err := authApi.GetVpcIdByName(vpcName)
		if err != nil {
			logrus.Errorf("could not fetch VPC ID: ", err.Error())
			return
		}
		_, err = authApi.DeleteVpc(vpcId).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.DeleteVpc``: %v\n", err)
			return
		}

		logrus.Infof("VPC %s was queued for termination", vpcName)
	},
}

func init() {
	getCmd.AddCommand(getVpcCmd)
	getVpcCmd.Flags().String("name", "", "Name for the VPC")

	createCmd.AddCommand(createVpcCmd)
	createVpcCmd.Flags().String("name", "", "Name for the VPC")
	createVpcCmd.MarkFlagRequired("name")
	createVpcCmd.Flags().String("cloud", "", "Cloud provider for the VPC")
	createVpcCmd.MarkFlagRequired("cloud")
	createVpcCmd.Flags().String("global-cidr", "", "Global CIDR for the VPC")
	createVpcCmd.Flags().StringSliceVar(&createRegions, "region", []string{}, "")
	createVpcCmd.Flags().StringSliceVar(&createCidrs, "cidr", []string{}, "")
	createVpcCmd.MarkFlagsRequiredTogether("region", "cidr")
	createVpcCmd.MarkFlagsMutuallyExclusive("global-cidr", "cidr")

	deleteCmd.AddCommand(deleteVpcCmd)
	deleteVpcCmd.Flags().String("name", "", "Name for the VPC")
	deleteVpcCmd.MarkFlagRequired("name")
}
