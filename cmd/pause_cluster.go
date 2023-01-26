/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// pauseClusterCmd represents the cluster command
var pauseClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Pause clusters in YugabyteDB Managed",
	Long:  "Pause clusters in YugabyteDB Managed",
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
		resp, r, err := authApi.PauseCluster(clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.PauseCluster`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}
		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}

		formatter.ClusterWrite(clustersCtx, []ybmclient.ClusterData{resp.GetData()})

		fmt.Printf("The cluster %s is being paused\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
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
