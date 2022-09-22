/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func getVpcByName(apiClient openapi.APIClient, accountID string, projectID string, name string) openapi.ApiListSingleTenantVpcsRequest {
	vpcListRequest := apiClient.NetworkApi.ListSingleTenantVpcs(context.Background(), accountID, projectID)
	if name != "" {
		vpcListRequest = vpcListRequest.Name(name)
	}

	return vpcListRequest
}

// vpcCmd represents the vpc command
var getVpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Get VPCs in YugabyteDB Managed",
	Long:  "Get VPCs in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		vpcName, _ := cmd.Flags().GetString("name")

		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		vpcListRequest := getVpcByName(*apiClient, accountID, projectID, vpcName)
		resp, r, err := vpcListRequest.Execute()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListVpcs``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}
		// response from `ListClusters`: ClusterListResponse
		prettyPrintJson(resp)
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

		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		resp, response, err := apiClient.NetworkApi.CreateVpc(context.Background(), accountID, projectID).SingleTenantVpcRequest(vpcRequest).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.CreateVpc``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", response)
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
		vpcName, _ := cmd.Flags().GetString("name")
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		readResp, readResponse, readErr := apiClient.NetworkApi.ListSingleTenantVpcs(context.Background(), accountID, projectID).Name(vpcName).Execute()
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Unable to find VPC with name %v. Error when calling `NetworkApi.ListSingleTenantVpcs``: %v\n", vpcName, readErr)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", readResponse)
			return
		}
		respData := readResp.Data
		if len(respData) == 0 {
			fmt.Fprintf(os.Stderr, "Unable to find VPC with name %v. Error when calling `NetworkApi.ListSingleTenantVpcs``: %v\n", vpcName, readErr)
			return
		}
		vpcId := respData[0].Info.Id

		resp, err := apiClient.NetworkApi.DeleteVpc(context.Background(), accountID, projectID, vpcId).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.DeleteVpc``: %v\n", err)
			return
		}

		fmt.Fprintf(os.Stdout, "VPC %v was queued for termination.\n", vpcName)
		prettyPrintJson(resp)
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
