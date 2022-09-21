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

// clustersCmd represents the clusters command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Get clusters in YugabyteDB Managed",
	Long:  `Get clusters in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		resp, r, err := apiClient.ClusterApi.ListClusters(context.Background(), accountID, projectID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.ListClusters``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}
		// response from `ListClusters`: ClusterListResponse
		prettyPrintJson(resp)
	},
}

func init() {
	getCmd.AddCommand(clusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clustersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clustersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
