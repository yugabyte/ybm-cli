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
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022- Yugabyte, Inc.

package nal

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var nalName string
var nalDescription string
var nalIpAddrs []string

var NalCmd = &cobra.Command{
	Use:   "network-allow-list",
	Short: "Manage Network Allow Lists",
	Long:  "Manage Network ALlow Lists",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getNetworkAllowListCmd = &cobra.Command{
	Use:   "get",
	Short: "Get network allow list in YugabyteDB Managed",
	Long:  "Get network allow list in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		var respFilter []ybmclient.NetworkAllowListData
		// No option to filter by name :(
		resp, r, err := authApi.ListNetworkAllowLists().Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `NetworkApi.ListNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		respFilter = resp.GetData()
		if cmd.Flags().Changed("name") {
			allowList, err := util.FindNetworkAllowList(resp.Data, nalName)
			if err != nil {
				logrus.Fatal(err)
			}

			respFilter = []ybmclient.NetworkAllowListData{allowList}
		}

		nalCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewNetworkAllowListFormat(viper.GetString("output")),
		}

		formatter.NetworkAllowListWrite(nalCtx, respFilter)
	},
}

var createNetworkAllowListCmd = &cobra.Command{
	Use:   "create",
	Short: "Create network allow lists in YugabyteDB Managed",
	Long:  "Create network allow lists in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		nalSpec := ybmclient.NetworkAllowListSpec{
			Name:        nalName,
			Description: nalDescription,
			AllowList:   nalIpAddrs,
		}

		resp, r, err := authApi.CreateNetworkAllowList().NetworkAllowListSpec(nalSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `NetworkApi.ListNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		nalCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewNetworkAllowListFormat(viper.GetString("output")),
		}
		respFilter := []ybmclient.NetworkAllowListData{resp.GetData()}

		formatter.NetworkAllowListWrite(nalCtx, respFilter)

		fmt.Printf("NetworkAllowList %s successful created\n", formatter.Colorize(nalName, formatter.GREEN_COLOR))
	},
}

var deleteNetworkAllowListCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete network allow list from YugabyteDB Managed",
	Long:  "Delete network allow list from YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.ListNetworkAllowLists().Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `NetworkApi.ListNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		allowList, err := util.FindNetworkAllowList(resp.Data, nalName)
		if err != nil {
			logrus.Error(err)
			return
		}

		r, err = authApi.DeleteNetworkAllowList(allowList.Info.Id).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `NetworkApi.DeleteNetworkAllowList`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		fmt.Printf("NetworkAllowList %s successfully deleted\n", formatter.Colorize(nalName, formatter.GREEN_COLOR))
	},
}

func init() {
	NalCmd.AddCommand(getNetworkAllowListCmd)
	getNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "[OPTIONAL] The name of the Network Allow List.")

	NalCmd.AddCommand(createNetworkAllowListCmd)
	createNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "[REQUIRED] The name of the Network Allow List.")
	createNetworkAllowListCmd.MarkFlagRequired("name")
	createNetworkAllowListCmd.Flags().StringVarP(&nalDescription, "description", "d", "", "[OPTIONAL] Description of the Network Allow List.")
	createNetworkAllowListCmd.Flags().StringSliceVarP(&nalIpAddrs, "ip-addr", "i", []string{}, "[REQUIRED] IP addresses included in the Network Allow List.")
	createNetworkAllowListCmd.MarkFlagRequired("ip-addr")

	NalCmd.AddCommand(deleteNetworkAllowListCmd)
	deleteNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "[REQUIRED] The name of the Network Allow List.")
	deleteNetworkAllowListCmd.MarkFlagRequired("name")
}
