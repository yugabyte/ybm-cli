/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

// getClusterCmd represents the cluster command
var getClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Get clusters in YugabyteDB Managed",
	Long:  "Get clusters in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterListRequest := authApi.ListClusters()

		// if user filters by name, add it to the request
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName != "" {
			clusterListRequest = clusterListRequest.Name(clusterName)
		}

		resp, r, err := clusterListRequest.Execute()

		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}
		formatter.ClusterWrite(clustersCtx, resp.GetData())
	},
}

func init() {
	getCmd.AddCommand(getClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	getClusterCmd.Flags().String("cluster-name", "", "The name of the cluster to get details")
}
