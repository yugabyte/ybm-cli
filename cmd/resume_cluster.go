/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

// resumeClusterCmd represents the cluster command
var resumeClusterCmd = &cobra.Command{
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
		_, r, err := authApi.ResumeCluster(clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.ResumeCluster`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}

		logrus.Infof("The cluster %v is being resumed", clusterName)
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
