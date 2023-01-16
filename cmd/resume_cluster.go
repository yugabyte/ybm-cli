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

// resumeClusterCmd represents the cluster command
var resumeClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Pause clusters in YugabyteDB Managed",
	Long:  "Pause clusters in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, projectID := getApiRequestInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, clusterIDOK, errMsg := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)
		if !clusterIDOK {
			fmt.Fprintf(os.Stderr, "Error when fetching cluster ID: %v\n", errMsg)
			return
		}

		_, _, err := apiClient.ClusterApi.ResumeCluster(context.Background(), accountID, projectID, clusterID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while resuming the cluster %v\n", clusterName)
			return
		}

		fmt.Fprintf(os.Stdout, "The cluster %v is being resumed\n", clusterName)
	},
}

func init() {
	resumeCmd.AddCommand(resumeClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resumeClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resumeClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	resumeClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be resumed")
	resumeClusterCmd.MarkFlagRequired("cluster-name")
}
