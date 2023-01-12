/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func findVpcPeering(vpcPeerings []openapi.VpcPeeringData, name string) (openapi.VpcPeeringData, error) {
	for _, vpcPeering := range vpcPeerings {
		if vpcPeering.Spec.Name == name {
			return vpcPeering, nil
		}
	}
	return openapi.VpcPeeringData{}, errors.New("Unable to find VpcPeering " + name)
}

var getVpcPeeringCmd = &cobra.Command{
	Use:   "vpc-peering",
	Short: "Get VPC peerings in YugabyteDB Managed",
	Long:  "Get VPC peerings in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {

		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		resp, r, err := apiClient.NetworkApi.ListVpcPeerings(context.Background(), accountID, projectID).Execute()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListVpcPeerings``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
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

		applicationVPCSpec := *openapi.NewCustomerVpcSpec(appVpcId, appProject, *openapi.NewVpcCloudInfo(openapi.CloudEnum(appCloud)))

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

		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		// check ybVpcName exists
		vpcListRequest := getVpcByName(*apiClient, accountID, projectID, ybVpcName)
		resp, r, err := vpcListRequest.Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListVpcs``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		if resp.Data == nil || len(resp.Data) == 0 {
			fmt.Fprintf(os.Stderr, "Error: VPC %s not found\n", ybVpcName)
			return
		}
		ybVpcId := resp.Data[0].Info.Id

		vpcPeeringSpec := *openapi.NewVpcPeeringSpec(ybVpcId, vpcPeeringName, applicationVPCSpec)
		vpcPeeringResp, response, err := apiClient.NetworkApi.CreateVpcPeering(context.Background(), accountID, projectID).VpcPeeringSpec(vpcPeeringSpec).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.CreateVpcPeering``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", response)
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

		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		resp, r, err := apiClient.NetworkApi.ListVpcPeerings(context.Background(), accountID, projectID).Execute()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListVpcPeerings``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		// check vpcPeeringName exists
		vpcPeering, findErr := findVpcPeering(resp.Data, vpcPeeringName)
		if findErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", findErr)
			return
		}
		vpcPeeringId := vpcPeering.Info.Id

		response, err := apiClient.NetworkApi.DeleteVpcPeering(context.Background(), accountID, projectID, vpcPeeringId).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.DeleteVpcPeering``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", response)
		}

		fmt.Fprintf(os.Stdout, "VPC-peering %v was queued for termination.\n", vpcPeeringName)
		prettyPrintJson(resp)
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
