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

// pauseClusterCmd represents the cluster command
var pauseClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Pause clusters in YugabyteDB Managed",
	Long:  "Pause clusters in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, clusterIDOK, errMsg := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)
		if !clusterIDOK {
			fmt.Fprintf(os.Stderr, "Error when fetching cluster ID: %v\n", errMsg)
			return
		}

		_, _, err := apiClient.ClusterApi.PauseCluster(context.Background(), accountID, projectID, clusterID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while pausing the cluster %v: %v\n", clusterName, err.Error())
			return
		}

		fmt.Fprintf(os.Stdout, "The cluster %v is being paused\n", clusterName)
	},
}

func init() {
	pauseCmd.AddCommand(pauseClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pauseClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pauseClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pauseClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be paused")
	pauseClusterCmd.MarkFlagRequired("cluster-name")
}
