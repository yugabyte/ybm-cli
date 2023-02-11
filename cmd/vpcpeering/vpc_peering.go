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

package vpcpeering

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var VPCPeeringCmd = &cobra.Command{
	Use:   "vpc-peering",
	Short: "Manage VPC Peerings",
	Long:  "Manage VPC Peerings",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func findVpcPeering(vpcPeerings []ybmclient.VpcPeeringData, name string) (ybmclient.VpcPeeringData, error) {
	for _, vpcPeering := range vpcPeerings {
		if vpcPeering.Spec.Name == name {
			return vpcPeering, nil
		}
	}
	return ybmclient.VpcPeeringData{}, errors.New("Unable to find VpcPeering " + name)
}

var getVpcPeeringCmd = &cobra.Command{
	Use:   "get",
	Short: "Get VPC peerings in YugabyteDB Managed",
	Long:  "Get VPC peerings in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.ListVpcPeerings().Execute()
		if err != nil {
			logrus.Errorf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `NetworkApi.ListVpcPeerings`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		vpcPeeringCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewVPCPeeringFormat(viper.GetString("output")),
		}

		// if user filters by name, add it to the request
		vpcPeeringName, _ := cmd.Flags().GetString("name")
		if vpcPeeringName != "" {
			vpcPeering, findErr := findVpcPeering(resp.Data, vpcPeeringName)
			if findErr != nil {
				logrus.Fatalf("Error: %s\n", findErr)
			}
			formatter.VPCPeeringWrite(vpcPeeringCtx, []ybmclient.VpcPeeringData{vpcPeering})
			return
		}

		formatter.VPCPeeringWrite(vpcPeeringCtx, resp.GetData())
	},
}

var createVpcPeeringCmd = &cobra.Command{
	Use:   "create",
	Short: "Create VPC peering in YugabyteDB Managed",
	Long:  "Create VPC peering in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		vpcPeeringName, _ := cmd.Flags().GetString("name")
		ybVpcName, _ := cmd.Flags().GetString("yb-vpc-name")
		appCloud, _ := cmd.Flags().GetString("cloud-provider")

		var applicationVPCSpec *ybmclient.CustomerVpcSpec

		// Validating and keeping the flow similar to the UI flow
		if appCloud == "AWS" {
			appAccountID, _ := cmd.Flags().GetString("app-vpc-account-id")
			if appAccountID == "" {
				logrus.Fatal("Could not create VPC peering: app-vpc-account-id is required for AWS.")
				return
			}
			appVpcID, _ := cmd.Flags().GetString("app-vpc-id")
			if appVpcID == "" {
				logrus.Fatal("Could not create VPC peering: app-vpc-id is required for AWS.")
				return
			}
			appVpcRegion, _ := cmd.Flags().GetString("app-vpc-region")
			if appVpcRegion == "" {
				logrus.Fatal("Could not create VPC peering: app-vpc-region is required for AWS.")
				return
			}

			appVpcCidr, _ := cmd.Flags().GetString("app-vpc-cidr")
			if appVpcCidr == "" {
				logrus.Fatal("Could not create VPC peering: app-vpc-cidr is required for AWS.")
				return
			}
			applicationVPCSpec = ybmclient.NewCustomerVpcSpec(appVpcID, appAccountID, *ybmclient.NewVpcCloudInfo(ybmclient.CloudEnum(appCloud)))
			applicationVPCSpec.CloudInfo.SetRegion(appVpcRegion)
			applicationVPCSpec.SetCidr(appVpcCidr)

		} else if appCloud == "GCP" {
			appProjectID, _ := cmd.Flags().GetString("app-vpc-project-id")
			if appProjectID == "" {
				logrus.Fatalf("Could not create VPC peering: app-vpc-project-id is required for GCP.")
			}
			appVpcName, _ := cmd.Flags().GetString("app-vpc-name")
			if appVpcName == "" {
				logrus.Fatalf("Could not create VPC peering: app-vpc-name is required for GCP.")
			}

			applicationVPCSpec = ybmclient.NewCustomerVpcSpec(appVpcName, appProjectID, *ybmclient.NewVpcCloudInfo(ybmclient.CloudEnum(appCloud)))

			// app vpc cidr is optional for GCP
			appVpcCidr, _ := cmd.Flags().GetString("app-vpc-cidr")
			if appVpcCidr != "" {
				applicationVPCSpec.SetCidr(appVpcCidr)
			}

		} else {
			logrus.Fatal("Could not create VPC peering: The cloud provider must be either GCP or AWS.")
		}

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		ybVpcId, err := authApi.GetVpcIdByName(ybVpcName)
		if err != nil {
			logrus.Fatalf("Unable to find VPC with name %v. Error: %v", ybVpcName, err)
		}

		ybVpcResp, resp, err := authApi.GetSingleTenantVpc(ybVpcId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf("Error when calling `GetSingleTenantVpc`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		ybVpcCloud := string(ybVpcResp.Data.Spec.GetCloud())

		if appCloud != ybVpcCloud {
			logrus.Error("The Yugabyte DB VPC and application VPC must be in the same cloud.")
			return
		}

		vpcPeeringSpec := *ybmclient.NewVpcPeeringSpec(ybVpcId, vpcPeeringName, *applicationVPCSpec)
		vpcPeeringResp, response, err := authApi.CreateVpcPeering().VpcPeeringSpec(vpcPeeringSpec).Execute()
		if err != nil {
			logrus.Errorf("Full HTTP response: %v", response)
			logrus.Fatalf("Error when calling `NetworkApi.CreateVpcPeering`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		vpcPeeringID := vpcPeeringResp.GetData().Info.Id

		msg := fmt.Sprintf("The VPC Peering %s is being created", formatter.Colorize(vpcPeeringName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(ybVpcId, "", "CREATE_VPC_PEERING", []string{"FAILED", "SUCCEEDED"}, msg, 1800)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The VPC Peering %s has been created\n", formatter.Colorize(vpcPeeringName, formatter.GREEN_COLOR))

			vpcPeeringResp, response, err = authApi.GetVpcPeering(vpcPeeringID).Execute()
			if err != nil {
				logrus.Errorf("Full HTTP response: %v", response)
				logrus.Fatalf("Error when calling `NetworkApi.ListVpcPeerings`: %s", ybmAuthClient.GetApiErrorDetails(err))
			}

		} else {
			fmt.Println(msg)
		}

		vpcPeeringCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewVPCPeeringFormat(viper.GetString("output")),
		}

		formatter.VPCPeeringWrite(vpcPeeringCtx, []ybmclient.VpcPeeringData{vpcPeeringResp.GetData()})
	},
}

var deleteVpcPeeringCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete VPC peering in YugabyteDB Managed",
	Long:  "Delete VPC peering in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		vpcPeeringName, _ := cmd.Flags().GetString("name")

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.ListVpcPeerings().Execute()
		if err != nil {
			logrus.Errorf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `NetworkApi.ListVpcPeerings`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		// check vpcPeeringName exists
		vpcPeering, err := findVpcPeering(resp.Data, vpcPeeringName)
		if err != nil {
			logrus.Fatalf("Error: %s\n", err)
		}
		vpcPeeringId := vpcPeering.Info.Id
		ybvpcID := vpcPeering.Spec.InternalYugabyteVpcId

		response, err := authApi.DeleteVpcPeering(vpcPeeringId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf("Error when calling `NetworkApi.ListVpcPeerings`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		msg := fmt.Sprintf("VPC peering %s is being terminated", formatter.Colorize(vpcPeeringName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(ybvpcID, "", "DELETE_VPC_PEERING", []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("VPC peering %s has been terminated.\n", formatter.Colorize(vpcPeeringName, formatter.GREEN_COLOR))
			return
		}
		fmt.Println(msg)
	},
}

func init() {
	VPCPeeringCmd.AddCommand(getVpcPeeringCmd)
	getVpcPeeringCmd.Flags().String("name", "", "[REQUIRED] Name for the VPC peering.")
	getVpcPeeringCmd.MarkFlagRequired("name")

	VPCPeeringCmd.AddCommand(createVpcPeeringCmd)
	createVpcPeeringCmd.Flags().String("name", "", "[REQUIRED] Name for the VPC peering.")
	createVpcPeeringCmd.MarkFlagRequired("name")
	createVpcPeeringCmd.Flags().String("yb-vpc-name", "", "[REQUIRED] Name of the YugabyteDB Managed VPC.")
	createVpcPeeringCmd.MarkFlagRequired("yb-vpc-name")
	createVpcPeeringCmd.Flags().String("cloud-provider", "", "[REQUIRED] Cloud of the VPC with which to peer. AWS or GCP.")
	createVpcPeeringCmd.MarkFlagRequired("cloud-provider")
	createVpcPeeringCmd.Flags().String("app-vpc-name", "", "[OPTIONAL] Name of the application VPC. Required for GCP. Not applicable for AWS.")
	createVpcPeeringCmd.Flags().String("app-vpc-project-id", "", "[OPTIONAL] Project ID of the application VPC. Required for GCP. Not applicable for AWS.")
	createVpcPeeringCmd.Flags().String("app-vpc-cidr", "", "[OPTIONAL] CIDR of the application VPC. Required for AWS. Optional for GCP.")
	createVpcPeeringCmd.Flags().String("app-vpc-account-id", "", "[OPTIONAL] Account ID of the application VPC. Required for AWS. Not applicable for GCP.")
	createVpcPeeringCmd.Flags().String("app-vpc-id", "", "[OPTIONAL] ID of the application VPC. Required for AWS. Not applicable for GCP.")
	createVpcPeeringCmd.Flags().String("app-vpc-region", "", "[OPTIONAL] Region of the application VPC. Required for AWS. Not applicable for GCP.")

	VPCPeeringCmd.AddCommand(deleteVpcPeeringCmd)
	deleteVpcPeeringCmd.Flags().String("name", "", "[REQUIRED] Name for the VPC peering.")
	deleteVpcPeeringCmd.MarkFlagRequired("name")
}
