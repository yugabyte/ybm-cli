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

var listNetworkAllowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List network allow lists in YugabyteDB Managed",
	Long:  "List network allow lists in YugabyteDB Managed",
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
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		respFilter = resp.GetData()
		// TODO: should we even allow this parameter to be specified when you do list?
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

// TODO: decide if we need a describe

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
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
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
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to delete %s: %s", "network-allow-list", nalName), viper.GetBool("force"))
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

		resp, r, err := authApi.ListNetworkAllowLists().Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		allowList, err := util.FindNetworkAllowList(resp.Data, nalName)
		if err != nil {
			logrus.Error(err)
			return
		}

		r, err = authApi.DeleteNetworkAllowList(allowList.Info.Id).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		fmt.Printf("NetworkAllowList %s successfully deleted\n", formatter.Colorize(nalName, formatter.GREEN_COLOR))
	},
}

func init() {
	NalCmd.AddCommand(listNetworkAllowListCmd)
	listNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "[OPTIONAL] The name of the Network Allow List.")

	NalCmd.AddCommand(createNetworkAllowListCmd)
	createNetworkAllowListCmd.Flags().SortFlags = false
	createNetworkAllowListCmd.Flags().StringSliceVarP(&nalIpAddrs, "ip-addr", "i", []string{}, "[REQUIRED] IP addresses included in the Network Allow List.")
	createNetworkAllowListCmd.MarkFlagRequired("ip-addr")
	createNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "[REQUIRED] The name of the Network Allow List.")
	createNetworkAllowListCmd.MarkFlagRequired("name")
	createNetworkAllowListCmd.Flags().StringVarP(&nalDescription, "description", "d", "", "[OPTIONAL] Description of the Network Allow List.")

	NalCmd.AddCommand(deleteNetworkAllowListCmd)
	deleteNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "[REQUIRED] The name of the Network Allow List.")
	deleteNetworkAllowListCmd.MarkFlagRequired("name")
	deleteNetworkAllowListCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
}
