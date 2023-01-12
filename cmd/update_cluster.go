/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// updateClusterCmd represents the cluster command
var updateClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Update a cluster in YB Managed",
	Long:  "Update a cluster in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {

		ctx := context.Background()

		apiClient, _ := getApiClient(ctx, cmd)
		accountID, _, _ := getAccountID(ctx, apiClient)
		projectID, _, _ := getProjectID(ctx, apiClient, accountID)

		clusterName, _ := cmd.Flags().GetString("cluster-name")

		clusterID, clusterIDOK, errMsg := getClusterID(ctx, apiClient, accountID, projectID, clusterName)
		if !clusterIDOK {
			fmt.Fprintf(os.Stderr, "Error while fetching cluster ID: %v\n", errMsg)
			return
		}

		resp, r, err := apiClient.ClusterApi.GetCluster(ctx, accountID, projectID, clusterID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.GetCluster``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}
		originalSpec := resp.Data.GetSpec()
		trackID := originalSpec.SoftwareInfo.GetTrackId()
		trackName, trackNameOK, errMsg := getTrackName(ctx, apiClient, accountID, trackID)
		if !trackNameOK {
			fmt.Fprintf(os.Stderr, "Error when calling `getTrackName``: %v\n", errMsg)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}
		populateFlags(cmd, originalSpec, trackName)

		regionInfoList := []map[string]string{}

		if cmd.Flags().Changed("region-info") {
			regionInfo, _ := cmd.Flags().GetStringToString("region-info")
			if _, ok := regionInfo["region"]; !ok {
				fmt.Fprintf(os.Stderr, "Region not specified in region info\n")
				return
			}
			if _, ok := regionInfo["num_nodes"]; !ok {
				fmt.Fprintf(os.Stderr, "Number of nodes not specified in region info\n")
				return
			}
			regionInfoList = append(regionInfoList, regionInfo)
		}

		if cmd.Flags().Changed("node-config") {
			nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
			if _, ok := nodeConfig["num_cores"]; !ok {
				fmt.Fprintf(os.Stderr, "Number of cores not specified in node config\n")
				return
			}
		}

		clusterSpec, clusterOK, errMsg := createClusterSpec(ctx, apiClient, cmd, accountID, regionInfoList)
		if !clusterOK {
			fmt.Fprintf(os.Stderr, "Error while creating cluster spec: %v\n", errMsg)
			return
		}

		clusterVersion := originalSpec.ClusterInfo.GetVersion()
		clusterSpec.ClusterInfo.SetVersion(clusterVersion)

		resp, r, err = apiClient.ClusterApi.EditCluster(ctx, accountID, projectID, clusterID).ClusterSpec(*clusterSpec).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.UpdateCluster``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}
		// response from `CreateCluster`: ClusterResponse
		fmt.Fprintf(os.Stdout, "Response from `ClusterApi.UpdateCluster`: %v\n", resp)
		fmt.Fprintf(os.Stdout, "The cluster %v is being updated\n", clusterName)
	},
}

func init() {
	updateCmd.AddCommand(updateClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	updateClusterCmd.Flags().String("cluster-name", "", "Name of the cluster.")
	updateClusterCmd.MarkFlagRequired("cluster-name")
	updateClusterCmd.Flags().String("cloud-type", "", "The cloud provider where database needs to be deployed. AWS or GCP.")
	updateClusterCmd.Flags().String("cluster-type", "", "Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED.")
	updateClusterCmd.Flags().StringToInt("node-config", nil, "Configuration of the cluster nodes.")
	updateClusterCmd.Flags().StringToString("region-info", nil, `Region information for the cluster. Please provide key value pairs
	region=<region-name>,num_nodes=<number-of-nodes>,vpc_id=<vpc-id> as the value. region and num_nodes are mandatory, vpc_id is optional.`)
	updateClusterCmd.Flags().String("cluster-tier", "", "The tier of the cluster. FREE or PAID.")
	updateClusterCmd.Flags().String("fault-tolerance", "", "The fault tolerance of the cluster. The possible values are NONE, ZONE and REGION.")
	updateClusterCmd.Flags().String("database-track", "", "The database track of the cluster. Stable or Preview.")

}

func populateFlags(cmd *cobra.Command, originalSpec ybmclient.ClusterSpec, trackName string) {
	if !cmd.Flags().Changed("cloud-type") {
		cmd.Flag("cloud-type").Value.Set(string(originalSpec.CloudInfo.GetCode()))
		cmd.Flag("cloud-type").Changed = true
	}
	if !cmd.Flags().Changed("cluster-type") {
		cmd.Flag("cluster-type").Value.Set(string(originalSpec.ClusterInfo.GetClusterType()))
		cmd.Flag("cluster-type").Changed = true
	}
	if !cmd.Flags().Changed("cluster-tier") {
		cmd.Flag("cluster-tier").Value.Set(string(originalSpec.ClusterInfo.GetClusterTier()))
		cmd.Flag("cluster-tier").Changed = true
	}
	if !cmd.Flags().Changed("fault-tolerance") {
		cmd.Flag("fault-tolerance").Value.Set(string(originalSpec.ClusterInfo.GetFaultTolerance()))
		cmd.Flag("fault-tolerance").Changed = true
	}
	if !cmd.Flags().Changed("database-track") {
		cmd.Flag("database-track").Value.Set(trackName)
		cmd.Flag("database-track").Changed = true
	}
	if !cmd.Flags().Changed("node-config") {
		nodeConfig := ""
		if diskSizeGb, ok := originalSpec.ClusterInfo.NodeInfo.GetDiskSizeGbOk(); ok {
			nodeConfig += "disk_size_gb=" + strconv.Itoa(int(*diskSizeGb))
		}
		if numCores, ok := originalSpec.ClusterInfo.NodeInfo.GetNumCoresOk(); ok {
			nodeConfig += ",num_cores=" + strconv.Itoa(int(*numCores))
		}
		cmd.Flag("node-config").Value.Set(nodeConfig)
		cmd.Flag("node-config").Changed = true

	}
	if !cmd.Flags().Changed("region-info") {
		regionInfo := ""
		if region, ok := originalSpec.ClusterRegionInfo[0].PlacementInfo.CloudInfo.GetRegionOk(); ok {
			regionInfo += "region=" + *region
		}
		if numNodes, ok := originalSpec.ClusterRegionInfo[0].PlacementInfo.GetNumNodesOk(); ok {
			regionInfo += ",num_nodes=" + strconv.Itoa(int(*numNodes))
		}
		if vpcID, ok := originalSpec.ClusterRegionInfo[0].PlacementInfo.GetVpcIdOk(); ok {
			regionInfo += ",vpc_id=" + *vpcID
		}
		cmd.Flag("region-info").Value.Set(regionInfo)
		cmd.Flag("region-info").Changed = true
	}

}
