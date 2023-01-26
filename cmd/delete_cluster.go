/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

// deleteClusterCmd represents the cluster command
var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Delete cluster in YugabyteDB Managed",
	Long:  "Delete cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}

		r, err := authApi.DeleteCluster(clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.DeleteCluster`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}

		fmt.Printf("The cluster %s is scheduled for deletion", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

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
