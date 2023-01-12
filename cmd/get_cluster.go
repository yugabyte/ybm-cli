/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// getClusterCmd represents the cluster command
var getClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Get clusters in YugabyteDB Managed",
	Long:  "Get clusters in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		clusterListRequest := apiClient.ClusterApi.ListClusters(context.Background(), accountID, projectID)

		// if user filters by name, add it to the request
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName != "" {
			clusterListRequest = clusterListRequest.Name(clusterName)
		}

		resp, r, err := clusterListRequest.Execute()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.ListClusters``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}
		// response from `ListClusters`: ClusterListResponse
		prettyPrintJson(resp)
	},
}

func init() {
	getCmd.AddCommand(getClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	getClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to get details")
}
