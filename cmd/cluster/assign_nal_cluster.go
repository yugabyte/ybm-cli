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

// assignClusterCmd represents the cluster command
var assignClusterCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign resources(e.g. network allow lists) to clusters in YugabyteDB Managed",
	Long:  "Assign resources(e.g. network allow lists) to clusters in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}
		newNetworkAllowListName, _ := cmd.Flags().GetString("network-allow-list")
		newNetworkAllowListId, err := authApi.GetNetworkAllowListIdByName(newNetworkAllowListName)
		if err != nil {
			logrus.Error(err)
			return
		}

		networkAllowListListResp, r, err := authApi.ListClusterNetworkAllowLists(clusterId).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.ListClusterNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		allowListIds := []string{}
		allowListIds = append(allowListIds, newNetworkAllowListId)
		for _, nal := range networkAllowListListResp.Data {
			allowListIds = append(allowListIds, nal.Info.GetId())
		}

		_, r, err = authApi.EditClusterNetworkAllowLists(clusterId, allowListIds).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.EditClusterNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		msg := fmt.Sprintf("The cluster %s is being assigned the network allow list %s", formatter.Colorize(clusterName, formatter.GREEN_COLOR), formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, "CLUSTER", "EDIT_ALLOW_LIST", []string{"FAILED", "SUCCEEDED"}, msg, 600)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
			}
			fmt.Printf("The cluster %s has been assigned the network allow list %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR), formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR))

		} else {
			fmt.Println(msg)
		}
	},
}

func init() {
	ClusterCmd.AddCommand(assignClusterCmd)
	assignClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be assignd")
	assignClusterCmd.MarkFlagRequired("cluster-name")
	assignClusterCmd.Flags().String("network-allow-list", "", "The name of the network allow list to be assignd")
	// Marked as required for now since as of now network allow list is the only resource that can be assigned
	assignClusterCmd.MarkFlagRequired("network-allow-list")
}
