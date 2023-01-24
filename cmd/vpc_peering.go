/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func findVpcPeering(vpcPeerings []ybmclient.VpcPeeringData, name string) (ybmclient.VpcPeeringData, error) {
	for _, vpcPeering := range vpcPeerings {
		if vpcPeering.Spec.Name == name {
			return vpcPeering, nil
		}
	}
	return ybmclient.VpcPeeringData{}, errors.New("Unable to find VpcPeering " + name)
}

var getVpcPeeringCmd = &cobra.Command{
	Use:   "vpc-peering",
	Short: "Get VPC peerings in YugabyteDB Managed",
	Long:  "Get VPC peerings in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		resp, r, err := authApi.ListVpcPeerings().Execute()

		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.ListVpcPeerings``: %v\n", err)
			logrus.Errorf("Full HTTP response: %v\n", r)
			return
		}

		// if user filters by name, add it to the request
		vpcPeeringName, _ := cmd.Flags().GetString("name")
		if vpcPeeringName != "" {
			vpcPeering, findErr := findVpcPeering(resp.Data, vpcPeeringName)
			if findErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", findErr)
				return
			}
			prettyPrintJson(vpcPeering)
			return
		}

		prettyPrintJson(resp)
	},
}

var createVpcPeeringCmd = &cobra.Command{
	Use:   "vpc-peering",
	Short: "Create VPC peering in YugabyteDB Managed",
	Long:  "Create VPC peering in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		vpcPeeringName, _ := cmd.Flags().GetString("name")
		ybVpcName, _ := cmd.Flags().GetString("vpc-name")
		appCloud, _ := cmd.Flags().GetString("cloud")
		appProject, _ := cmd.Flags().GetString("project")
		appVpcId, _ := cmd.Flags().GetString("vpc-id")

		applicationVPCSpec := *ybmclient.NewCustomerVpcSpec(appVpcId, appProject, *ybmclient.NewVpcCloudInfo(ybmclient.CloudEnum(appCloud)))

		// Validations
		if appCloud == "AWS" {
			region, _ := cmd.Flags().GetString("region")
			if region == "" {
				fmt.Fprintf(os.Stderr, "Error: region is required for AWS\n")
				return
			}

			cidr, _ := cmd.Flags().GetString("cidr")
			if cidr == "" {
				fmt.Fprintf(os.Stderr, "Error: cidr is required for AWS\n")
				return
			}

			applicationVPCSpec.CloudInfo.SetRegion(region)
			applicationVPCSpec.SetCidr(cidr)
		}
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		vpcListRequest := authApi.ListSingleTenantVpcsByName(ybVpcName)
		resp, r, err := vpcListRequest.Execute()
		if err != nil {
			logrus.Errorf("Unable to find VPC with name %v. Error when calling `NetworkApi.ListSingleTenantVpcs``: %v\n", ybVpcName, err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}

		if resp.Data == nil || len(resp.Data) == 0 {
			logrus.Errorf("Error: VPC %s not found\n", ybVpcName)
			return
		}
		ybVpcId := resp.Data[0].Info.Id

		vpcPeeringSpec := *ybmclient.NewVpcPeeringSpec(ybVpcId, vpcPeeringName, applicationVPCSpec)
		vpcPeeringResp, response, err := authApi.CreateVpcPeering().VpcPeeringSpec(vpcPeeringSpec).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.CreateVpcPeering``: %v\n", err)
			logrus.Errorf("Full HTTP response: %v\n", response)
			return
		}

		prettyPrintJson(vpcPeeringResp)
	},
}

var deleteVpcPeeringCmd = &cobra.Command{
	Use:   "vpc-peering",
	Short: "Delete VPC peering in YugabyteDB Managed",
	Long:  "Delete VPC peering in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		vpcPeeringName, _ := cmd.Flags().GetString("name")

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		resp, r, err := authApi.ListVpcPeerings().Execute()

		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.ListVpcPeerings``: %v\n", err)
			logrus.Errorf("Full HTTP response: %v\n", r)
			return
		}

		// check vpcPeeringName exists
		vpcPeering, err := findVpcPeering(resp.Data, vpcPeeringName)
		if err != nil {
			logrus.Errorf("Error: %s\n", err)
			return
		}
		vpcPeeringId := vpcPeering.Info.Id

		response, err := authApi.DeleteVpcPeering(vpcPeeringId).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.ListVpcPeerings``: %v\n", err)
			logrus.Errorf("Full HTTP response: %v\n", response)
			return
		}

		logrus.Infof("VPC-peering %v was queued for termination.\n", vpcPeeringName)
	},
}

func init() {
	getCmd.AddCommand(getVpcPeeringCmd)
	getVpcPeeringCmd.Flags().String("name", "", "Name for the VPC peering")

	createCmd.AddCommand(createVpcPeeringCmd)
	createVpcPeeringCmd.Flags().String("name", "", "Name for the VPC peering")
	createVpcPeeringCmd.MarkFlagRequired("name")
	createVpcPeeringCmd.Flags().String("vpc-name", "", "Name of the VPC to peer")
	createVpcPeeringCmd.MarkFlagRequired("vpc-name")
	createVpcPeeringCmd.Flags().String("cloud", "", "Cloud of the VPC with which to peer")
	createVpcPeeringCmd.MarkFlagRequired("cloud")
	createVpcPeeringCmd.Flags().String("project", "", "Project of the VPC with which to peer")
	createVpcPeeringCmd.MarkFlagRequired("project")
	createVpcPeeringCmd.Flags().String("vpc-id", "", "ID of the VPC with which to peer")
	createVpcPeeringCmd.MarkFlagRequired("vpc-id")
	createVpcPeeringCmd.Flags().String("region", "", "Region of the VPC with which to peer")
	createVpcPeeringCmd.Flags().String("cidr", "", "CIDR of the VPC with which to peer")

	deleteCmd.AddCommand(deleteVpcPeeringCmd)
	deleteVpcPeeringCmd.Flags().String("name", "", "Name for the VPC peering")
}
