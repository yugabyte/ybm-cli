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

// deleteClusterCmd represents the cluster command
var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Delete cluster in YugabyteDB Managed",
	Long:  "Delete cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, projectID := getApiRequestInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, clusterIDOK, errMsg := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)
		if !clusterIDOK {
			fmt.Fprintf(os.Stderr, "Error when fetching cluster ID: %v\n", errMsg)
			return
		}
		r, err := apiClient.ClusterApi.DeleteCluster(context.Background(), accountID, projectID, clusterID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.ListClusters``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		fmt.Fprintf(os.Stdout, "The cluster %v is scheduled for deletion\n", clusterName)

	},
}

func init() {
	deleteCmd.AddCommand(deleteClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	deleteClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be deleted")
	deleteClusterCmd.MarkFlagRequired("cluster-name")
}
