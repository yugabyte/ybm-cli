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
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var nalName string
var nalDescription string
var nalIpAddrs []string

func findNetworkAllowList(nals []ybmclient.NetworkAllowListData, name string) (ybmclient.NetworkAllowListData, error) {
	for _, allowList := range nals {
		if allowList.Spec.Name == nalName {
			return allowList, nil
		}
	}
	return ybmclient.NetworkAllowListData{}, errors.New("Unable to find NetworkAllowList " + name)
}

var getNetworkAllowListCmd = &cobra.Command{
	Use:   "network_allow_list",
	Short: "Get network allow list in YugabyteDB Managed",
	Long:  `Get network allow list in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		var respFilter []ybmclient.NetworkAllowListData
		// No option to filter by name :(
		resp, r, err := authApi.ListNetworkAllowLists().Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.ListNetworkAllowLists`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}

		respFilter = resp.GetData()
		if cmd.Flags().Changed("name") {
			allowList, err := findNetworkAllowList(resp.Data, nalName)

			if err != nil {
				logrus.Error(err)
				return
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
	Use:   "network_allow_list",
	Short: "Create network allow lists in YugabyteDB Managed",
	Long:  `Create network allow lists in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		nalSpec := ybmclient.NetworkAllowListSpec{
			Name:        nalName,
			Description: nalDescription,
			AllowList:   nalIpAddrs,
		}

		resp, r, err := authApi.CreateNetworkAllowList().NetworkAllowListSpec(nalSpec).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.ListNetworkAllowLists`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
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
	Use:   "network_allow_list",
	Short: "Delete network allow list from YugabyteDB Managed",
	Long:  `Delete network allow list from YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.ListNetworkAllowLists().Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.ListNetworkAllowLists`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}

		allowList, err := findNetworkAllowList(resp.Data, nalName)
		if err != nil {
			logrus.Error(err)
			return
		}

		r, err = authApi.DeleteNetworkAllowList(allowList.Info.Id).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `NetworkApi.DeleteNetworkAllowList`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}
		fmt.Printf("NetworkAllowList %s successfully deleted\n", formatter.Colorize(nalName, formatter.GREEN_COLOR))
	},
}

func init() {
	getNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "The name of the Network Allow List")
	getCmd.AddCommand(getNetworkAllowListCmd)

	createNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "The name of the Network Allow List")
	createNetworkAllowListCmd.MarkFlagRequired("name")
	createNetworkAllowListCmd.Flags().StringVarP(&nalDescription, "description", "d", "", "Description of the Network Allow List")
	createNetworkAllowListCmd.Flags().StringSliceVarP(&nalIpAddrs, "ip_addr", "i", []string{}, "IP addresses included in the Network Allow List")
	createNetworkAllowListCmd.MarkFlagRequired("ip_addr")
	createCmd.AddCommand(createNetworkAllowListCmd)

	deleteNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "The name of the Network Allow List")
	deleteNetworkAllowListCmd.MarkFlagRequired("name")
	deleteCmd.AddCommand(deleteNetworkAllowListCmd)
}
