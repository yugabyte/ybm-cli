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

// clustersCmd represents the clusters command
var clustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "List clusters in YugabyteDB Managed",
	Long:  `List clusters in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		configuration := ybmclient.NewConfiguration()
		//Configure the client

		configuration.Host = os.Getenv("YBM_HOST")
		configuration.Scheme = "https"
		apiClient := ybmclient.NewAPIClient(configuration)
		// authorize user with api key
		apiKey := os.Getenv("YBM_API_KEY")
		apiClient.GetConfig().AddDefaultHeader("Authorization", "Bearer "+apiKey)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		resp, r, err := apiClient.ClusterApi.ListClusters(context.Background(), accountID, projectID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.ListClusters``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}
		// response from `ListClusters`: ClusterListResponse
		fmt.Fprintf(os.Stdout, "Response from `ClusterApi.ListClusters`: %v\n", resp)
	},
}

func init() {
	listCmd.AddCommand(clustersCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clustersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clustersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
