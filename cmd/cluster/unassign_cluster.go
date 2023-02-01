/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

// unassignClusterCmd represents the cluster command
var unassignClusterCmd = &cobra.Command{
	Use:   "unassign",
	Short: "Unassign resources(e.g. network allow lists) to clusters in YugabyteDB Managed",
	Long:  "Unassign resources(e.g. network allow lists) to clusters in YugabyteDB Managed",
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
		nalFound := false
		for _, nal := range networkAllowListListResp.Data {
			nalId := nal.Info.GetId()
			if nalId == newNetworkAllowListId {
				nalFound = true
			} else {
				allowListIds = append(allowListIds, nalId)
			}
		}
		if !nalFound {
			logrus.Errorf("The allow list %s is not associated with the cluster %s", formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		}

		_, r, err = authApi.EditClusterNetworkAllowLists(clusterId, allowListIds).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.EditClusterNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		fmt.Printf("The cluster %s is being unassigned the network allow list %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR), formatter.Colorize(newNetworkAllowListName, formatter.GREEN_COLOR))
	},
}

func init() {
	ClusterCmd.AddCommand(unassignClusterCmd)
	unassignClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to be unassignd")
	unassignClusterCmd.MarkFlagRequired("cluster-name")
	unassignClusterCmd.Flags().String("network-allow-list", "", "The name of the network allow list to be unassignd")
	// Marked as required for now since as of now network allow list is the only resource that can be unassigned
	unassignClusterCmd.MarkFlagRequired("network-allow-list")
}
