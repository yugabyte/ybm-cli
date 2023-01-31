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
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// resumeClusterCmd represents the cluster command
var resumeClusterCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a cluster in YugabyteDB Managed",
	Long:  "Resume a cluster in YugabyteDB Managed",
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
		resp, r, err := authApi.ResumeCluster(clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.ResumeCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}

		formatter.ClusterWrite(clustersCtx, []ybmclient.ClusterData{resp.GetData()})

		msg := fmt.Sprintf("The cluster %s is being resumed", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "RESUME_CLUSTER", []string{"FAILED", "SUCCEEDED"}, msg, 240)
	},
}

func init() {
	ClusterCmd.AddCommand(resumeClusterCmd)

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
