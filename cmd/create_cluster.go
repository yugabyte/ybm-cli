/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// createClusterCmd represents the cluster command
var createClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Create a cluster in YB Managed",
	Long:  "Create a cluster in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		credentials, _ := cmd.Flags().GetStringToString("credentials")

		username := credentials["username"]
		password := credentials["password"]
		regionInfoMapList := []map[string]string{}
		if cmd.Flags().Changed("region-info") {
			regionInfoList, _ := cmd.Flags().GetStringArray("region-info")
			for _, regionInfoString := range regionInfoList {
				regionInfoMap := map[string]string{}
				for _, regionInfo := range strings.Split(regionInfoString, ",") {
					kvp := strings.Split(regionInfo, "=")
					key := kvp[0]
					val := kvp[1]
					switch key {
					case "region":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["region"] = val
						}
					case "num_nodes":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["num_nodes"] = val
						}
					case "vpc":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["vpc"] = val
						}
					}
				}

				if _, ok := regionInfoMap["region"]; !ok {
					logrus.Errorln("Region not specified in region info")
					return
				}
				if _, ok := regionInfoMap["num_nodes"]; !ok {
					logrus.Errorln("Number of nodes not specified in region info")
					return
				}

				regionInfoMapList = append(regionInfoMapList, regionInfoMap)
			}
		}

		if cmd.Flags().Changed("node-config") {
			nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
			if _, ok := nodeConfig["num_cores"]; !ok {
				logrus.Errorln("Number of cores not specified in node config")
				return
			}
		}

		clusterSpec, err := authApi.CreateClusterSpec(cmd, regionInfoMapList)
		if err != nil {
			logrus.Errorf("Error while creating cluster spec: %v", err)
			return
		}

		dbCredentials := ybmclient.NewCreateClusterRequestDbCredentials()
		dbCredentials.Ycql = ybmclient.NewDBCredentials(username, password)
		dbCredentials.Ysql = ybmclient.NewDBCredentials(username, password)

		createClusterRequest := ybmclient.NewCreateClusterRequest(*clusterSpec, *dbCredentials)

		resp, r, err := authApi.CreateCluster().CreateClusterRequest(*createClusterRequest).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.CreateCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v", r)
			return
		}

		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}

		formatter.ClusterWrite(clustersCtx, []ybmclient.ClusterData{resp.GetData()})

		fmt.Printf("The cluster %s is being created\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
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
	createClusterCmd.Flags().String("cluster-name", "", "Name of the cluster.")
	createClusterCmd.MarkFlagRequired("cluster-name")
	createClusterCmd.Flags().StringToString("credentials", nil, `Credentials to login to the cluster. Please provide key value pairs
	username=<user-name>,password=<password>.`)
	createClusterCmd.MarkFlagRequired("credentials")

	createClusterCmd.Flags().String("cloud-type", "", "The cloud provider where database needs to be deployed. AWS or GCP.")
	createClusterCmd.Flags().String("cluster-type", "", "Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED.")
	createClusterCmd.Flags().StringToInt("node-config", nil, "Configuration of the cluster nodes.")
	createClusterCmd.Flags().StringArray("region-info", []string{}, `Region information for the cluster. Please provide key value pairs
	region=<region-name>,num_nodes=<number-of-nodes>,vpc=<vpc-name> as the value. region and num_nodes are mandatory, vpc is optional.`)
	createClusterCmd.Flags().String("cluster-tier", "", "The tier of the cluster. FREE or PAID.")
	createClusterCmd.Flags().String("fault-tolerance", "", "The fault tolerance of the cluster. The possible values are NONE, ZONE and REGION.")
	createClusterCmd.Flags().String("database-track", "", "The database track of the cluster. Stable or Preview.")

}
