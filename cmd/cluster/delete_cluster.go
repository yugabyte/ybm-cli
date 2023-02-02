/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

// deleteClusterCmd represents the cluster command
var deleteClusterCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete cluster in YugabyteDB Managed",
	Long:  "Delete cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
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
			logrus.Errorf("Error when calling `ClusterApi.DeleteCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		msg := fmt.Sprintf("The cluster %s is being deleted", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "DELETE_CLUSTER", []string{"FAILED", "SUCCEEDED"}, msg, 1500)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
			}
			fmt.Printf("The cluster %s has been deleted\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
			return
		}
		fmt.Println(msg)
	},
}

func init() {
	ClusterCmd.AddCommand(deleteClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	deleteClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be deleted")
	deleteClusterCmd.MarkFlagRequired("cluster-name")
}
