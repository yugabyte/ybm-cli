/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// createClusterCmd represents the cluster command
var createClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Create a cluster in YB Managed",
	Long:  "Create a cluster in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {

		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		credentials, _ := cmd.Flags().GetStringToString("credentials")

		username := credentials["username"]
		password := credentials["password"]

		clusterSpec := ybmclient.NewClusterSpecWithDefaults()
		clusterSpec.SetName(clusterName)
		clusterSpec.SetCloudInfo(*ybmclient.NewCloudInfoWithDefaults())
		clusterSpec.SetClusterInfo(*ybmclient.NewClusterInfoWithDefaults())
		clusterSpec.ClusterInfo.SetNodeInfo(*ybmclient.NewClusterNodeInfoWithDefaults())
		clusterSpec.SetNetworkInfo(*ybmclient.NewNetworkingWithDefaults())
		clusterSpec.SetSoftwareInfo(*ybmclient.NewSoftwareInfoWithDefaults())

		dbCredentials := ybmclient.NewCreateClusterRequestDbCredentials()
		dbCredentials.Ycql = ybmclient.NewDBCredentials(username, password)
		dbCredentials.Ysql = ybmclient.NewDBCredentials(username, password)

		createClusterRequest := ybmclient.NewCreateClusterRequest(*clusterSpec, *dbCredentials)

		resp, r, err := apiClient.ClusterApi.CreateCluster(context.Background(), accountID, projectID).CreateClusterRequest(*createClusterRequest).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.CreateCluster``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}
		// response from `CreateCluster`: ClusterResponse
		fmt.Fprintf(os.Stdout, "Response from `ClusterApi.CreateCluster`: %v\n", resp)
		fmt.Fprintf(os.Stdout, "The cluster %v is being creted\n", clusterName)
	},
}

func init() {
	createCmd.AddCommand(createClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createClusterCmd.Flags().String("cluster-name", "", "Name of the cluster")
	createClusterCmd.MarkFlagRequired("cluster-name")
	createClusterCmd.Flags().StringToString("credentials", nil, "Credentials to login to the cluster")
	createClusterCmd.MarkFlagRequired("credentials")
	createClusterCmd.Flags().String("cloud-type", "", "The cloud provider where database needs to be deployed. AWS or GCP")
	createClusterCmd.Flags().String("cluster-type", "", "Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED")
	createClusterCmd.Flags().StringToInt("node-config", nil, "Configuration of the cluster nodes")
	createClusterCmd.Flags().StringToString("region-info", nil, "Region information for the cluster")
	createClusterCmd.Flags().String("cluster-tier", "", "The tier of the cluster. FREE or PAID")

}
