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

package vpc

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	"github.com/yugabyte/ybm-cli/cmd/vpc/peering"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var VPCCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Manage VPCs",
	Long:  "Manage VPCs",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getVpcCmd = &cobra.Command{
	Use:   "get",
	Short: "Get VPC in YugabyteDB Managed",
	Long:  `Get VPC in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		listVpcCmd.Run(cmd, args)
		logrus.Warnln("\nThe command `ybm vpc get` is deprecated. Please use `ybm vpc list` instead.")
	},
}

// vpcCmd represents the vpc command
var listVpcCmd = &cobra.Command{
	Use:   "list",
	Short: "List VPCs in YugabyteDB Managed",
	Long:  "List VPCs in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		vpcName, _ := cmd.Flags().GetString("name")

		vpcListRequest := authApi.ListSingleTenantVpcsByName(vpcName)

		resp, r, err := vpcListRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		// response from `ListClusters`: ClusterListResponse
		vpcCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewVPCFormat(viper.GetString("output")),
		}

		formatter.VPCWrite(vpcCtx, resp.GetData())
	},
}

// TODO: make list VPC not show the IP ranges and other details and build a describe for showing that.

var createRegions []string
var createCidrs []string
var createVpcCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a VPC in YugabyteDB Managed",
	Long:  "Create a VPC in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		// Validations
		if !cmd.Flags().Changed("global-cidr") && !cmd.Flags().Changed("cidr") {
			logrus.Fatal("Either global-cidr or cidr must be specified")
		}

		if len(createRegions) != len(createCidrs) {
			logrus.Fatal("Number of regions and cidrs must be equal")
		}

		vpcName, _ := cmd.Flags().GetString("name")
		cloud, _ := cmd.Flags().GetString("cloud-provider")
		globalCidrRange, _ := cmd.Flags().GetString("global-cidr")

		// global CIDR only works with GCP
		if cloud != "GCP" && cmd.Flags().Changed("global-cidr") {
			logrus.Fatal("global-cidr is only supported for GCP")
		}

		// If non-global CIDR, validate that there are different regions specified
		regionMap := map[string]int{}
		vpcRegionSpec := []ybmclient.VpcRegionSpec{}

		if cmd.Flags().Changed("cidr") {
			for index, region := range createRegions {
				cidr := createCidrs[index]
				if valid, err := util.ValidateCIDR(cidr); !valid {
					logrus.Fatal(err)
				}
				spec := *ybmclient.NewVpcRegionSpecWithDefaults()
				regionMap[region] = index
				spec.SetRegion(region)
				spec.SetCidr(cidr)
				vpcRegionSpec = append(vpcRegionSpec, spec)
			}
			if len(regionMap) != len(createRegions) {
				logrus.Fatal("Regions must be unique")
			}
		}

		vpcSpec := *ybmclient.NewSingleTenantVpcSpec(vpcName, ybmclient.CloudEnum(cloud), vpcRegionSpec)
		if cmd.Flags().Changed("global-cidr") {
			if valid, err := util.ValidateCIDR(globalCidrRange); !valid {
				logrus.Fatal(err)
			}
			vpcSpec.SetParentCidr(globalCidrRange)
		}
		vpcRequest := *ybmclient.NewSingleTenantVpcRequest(vpcSpec)

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		resp, r, err := authApi.CreateVpc().SingleTenantVpcRequest(vpcRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		vpcID := resp.Data.GetInfo().Id
		vpcData := []ybmclient.SingleTenantVpcDataResponse{resp.GetData()}

		msg := fmt.Sprintf("The VPC %s is being created", formatter.Colorize(vpcName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(vpcID, "", ybmclient.TASKTYPEENUM_CREATE_VPC, []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The VPC %s has been created\n", formatter.Colorize(vpcName, formatter.GREEN_COLOR))

			vpcListRequest := authApi.ListSingleTenantVpcsByName(vpcName)
			respC, r, err := vpcListRequest.Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			vpcData = respC.GetData()
		} else {
			fmt.Println(msg)
		}
		vpcCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewVPCFormat(viper.GetString("output")),
		}
		formatter.VPCWrite(vpcCtx, vpcData)

	},
}

var deleteVpcCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a VPC in YugabyteDB Managed",
	Long:  "Delete a VPC in YugabyteDB Managed",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		vpcName, _ := cmd.Flags().GetString("name")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to delete %s: %s", "vpc", vpcName), viper.GetBool("force"))
		if err != nil {
			logrus.Fatal(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		vpcName, _ := cmd.Flags().GetString("name")
		if vpcName == "" {
			logrus.Fatal("name field is required")
		}
		vpcId, err := authApi.GetVpcIdByName(vpcName)
		if err != nil {
			logrus.Fatalf("could not fetch VPC ID: %s", err.Error())
		}
		r, err := authApi.DeleteVpc(vpcId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The VPC %s is being deleted", formatter.Colorize(vpcName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(vpcId, "", ybmclient.TASKTYPEENUM_DELETE_VPC, []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The VPC %s has been deleted\n", formatter.Colorize(vpcName, formatter.GREEN_COLOR))
			return
		}
		fmt.Println(msg)
	},
}

func init() {
	VPCCmd.AddCommand(peering.VPCPeeringCmd)

	VPCCmd.AddCommand(getVpcCmd)
	getVpcCmd.Flags().String("name", "", "[OPTIONAL] Name for the VPC.")

	VPCCmd.AddCommand(listVpcCmd)
	listVpcCmd.Flags().String("name", "", "[OPTIONAL] Name for the VPC.")

	VPCCmd.AddCommand(createVpcCmd)
	createVpcCmd.Flags().SortFlags = false
	createVpcCmd.Flags().String("name", "", "[REQUIRED] Name for the VPC.")
	createVpcCmd.MarkFlagRequired("name")
	createVpcCmd.Flags().String("cloud-provider", "", "[REQUIRED] Cloud provider for the VPC.")
	createVpcCmd.MarkFlagRequired("cloud-provider")
	createVpcCmd.Flags().String("global-cidr", "", "[OPTIONAL] Global CIDR for the VPC.")
	createVpcCmd.Flags().StringSliceVar(&createRegions, "region", []string{}, "[OPTIONAL] Region of the VPC.")
	createVpcCmd.Flags().StringSliceVar(&createCidrs, "cidr", []string{}, "[OPTIONAL] CIDR of the VPC.")
	createVpcCmd.MarkFlagsRequiredTogether("region", "cidr")
	createVpcCmd.MarkFlagsMutuallyExclusive("global-cidr", "cidr")

	VPCCmd.AddCommand(deleteVpcCmd)
	deleteVpcCmd.Flags().String("name", "", "[REQUIRED] Name for the VPC.")
	deleteVpcCmd.MarkFlagRequired("name")
	deleteVpcCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
}
