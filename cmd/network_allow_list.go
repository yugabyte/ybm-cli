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

var nalName string
var nalDescription string
var nalIpAddrs []string

var getNetworkAllowListCmd = &cobra.Command{
	Use:   "network_allow_list",
	Short: "Get network allow list in YugabyteDB Managed",
	Long:  `Get network allow list in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		// No option to filter by name :(
		resp, r, err := apiClient.NetworkApi.ListNetworkAllowLists(context.Background(), accountID, projectID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListNetworkAllowLists`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		if cmd.Flags().Changed("name") {
			for _, allowList := range resp.Data {
				if allowList.Spec.Name == nalName {
					prettyPrintJson(allowList)
					return
				}
			}
			fmt.Fprintf(os.Stderr, "NetworkAllowList <%s> does not exist \n", nalName)
			return
		}

		prettyPrintJson(resp)
	},
}

var createNetworkAllowListCmd = &cobra.Command{
	Use:   "network_allow_list",
	Short: "Create network allow lists in YugabyteDB Managed",
	Long:  `Create network allow lists in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		nalSpec := ybmclient.NetworkAllowListSpec{
			Name:        nalName,
			Description: nalDescription,
			AllowList:   nalIpAddrs,
		}

		resp, r, err := apiClient.NetworkApi.CreateNetworkAllowList(context.Background(), accountID, projectID).NetworkAllowListSpec(nalSpec).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.CreateNetworkAllowList``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		prettyPrintJson(resp)
	},
}

var deleteNetworkAllowListCmd = &cobra.Command{
	Use:   "network_allow_list",
	Short: "Delete network allow list from YugabyteDB Managed",
	Long:  `Delete network allow list from YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		readResp, readResponse, readErr := apiClient.NetworkApi.ListNetworkAllowLists(context.Background(), accountID, projectID).Execute()

		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListNetworkAllowLists`: %v\n", readErr)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", readResponse)
			return
		}
		var allowListID string

		for _, allowList := range readResp.Data {
			if allowList.Spec.Name == nalName {
				allowListID = allowList.Info.Id
				break
			}
		}

		if allowListID == "" {
			fmt.Fprintf(os.Stderr, "NetworkAllowList <%s> does not exist \n", nalName)
			return
		}

		r, err := apiClient.NetworkApi.DeleteNetworkAllowList(context.Background(), accountID, projectID, allowListID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.DeleteNetworkAllowList``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		fmt.Fprintf(os.Stdout, "Success: NetworkAllosList <%s> deleted\n", nalName)
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
