/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var clusterName string
var allReplicaOpt []string
var test map[string]string

// Parse array of read replica string to string params
func parseReplicaOpts(replicaOpts []string) []ybmclient.ReadReplicaSpec {
	readReplicaSpecs := []ybmclient.ReadReplicaSpec{}

	for _, replicaOpt := range replicaOpts {
		// Default Values
		spec := ybmclient.ReadReplicaSpec{
			NodeInfo: ybmclient.ClusterNodeInfo{
				NumCores:   2,
				MemoryMb:   4096,
				DiskSizeGb: 10,
			},
			PlacementInfo: ybmclient.PlacementInfo{
				CloudInfo: ybmclient.CloudInfo{
					Code:   ybmclient.CLOUDENUM_AWS,
					Region: "us-west-2",
				},
				NumNodes: 1,
			},
		}

		for _, subOpt := range strings.Split(replicaOpt, ",") {
			kvp := strings.Split(subOpt, "=")
			key := kvp[0]
			val := kvp[1]
			n, _ := strconv.Atoi(val)
			switch key {
			case "num_cores":
				spec.NodeInfo.NumCores = int32(n)
			case "memory_mb":
				spec.NodeInfo.MemoryMb = int32(n)
			case "disk_size_gb":
				spec.NodeInfo.DiskSizeGb = int32(n)
			case "code":
				spec.PlacementInfo.CloudInfo.Code = ybmclient.CloudEnum(val)
			case "region":
				spec.PlacementInfo.CloudInfo.Region = val
			case "num_nodes":
				spec.PlacementInfo.NumNodes = int32(n)
			case "vpc_id":
				spec.PlacementInfo.VpcId = *ybmclient.NewNullableString(&val)
			case "num_replicas":
				numReplicas := int32(n)
				spec.PlacementInfo.NumReplicas = *ybmclient.NewNullableInt32(&numReplicas)
			case "multi_zone":
				isMultiZone, _ := strconv.ParseBool(val)
				spec.PlacementInfo.MultiZone = *ybmclient.NewNullableBool(&isMultiZone)
			}

		}
		readReplicaSpecs = append(readReplicaSpecs, spec)
	}
	return readReplicaSpecs
}

var getReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Get read replica in YugabyteDB Managed",
	Long:  "Get read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, projectID := getApiRequestInfo("", "")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		resp, r, err := apiClient.ReadReplicaApi.ListReadReplicas(context.Background(), accountID, projectID, clusterID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ReadReplicaApi.ListReadReplicas`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		prettyPrintJson(resp)
	},
}

var createReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Create read replica in YugabyteDB Managed",
	Long:  "Create read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, projectID := getApiRequestInfo("", "")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		readReplicaSpecs := parseReplicaOpts(allReplicaOpt)

		resp, r, err := apiClient.ReadReplicaApi.CreateReadReplica(context.Background(), accountID, projectID, clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ReadReplicaApi.CreateReadReplica``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		prettyPrintJson(resp)
	},
}

var updateReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Edit read replica in YugabyteDB Managed",
	Long:  "Edit read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, projectID := getApiRequestInfo("", "")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		readReplicaSpecs := parseReplicaOpts(allReplicaOpt)

		resp, r, err := apiClient.ReadReplicaApi.EditReadReplicas(context.Background(), accountID, projectID, clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ReadReplicaApi.EditReadReplicas``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		prettyPrintJson(resp)
	},
}

var deleteReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Delete read replica from YugabyteDB Managed",
	Long:  "Delete read replica from YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, projectID := getApiRequestInfo("", "")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		r, err := apiClient.ReadReplicaApi.DeleteReadReplica(context.Background(), accountID, projectID, clusterID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ReadReplicaApi.DeleteReadReplica``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		fmt.Fprintf(os.Stdout, "Success: deleted all read replicas deleted for cluster: %v\n", clusterName)
	},
}

func init() {

	getReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "The name of the cluster")
	getReadReplicaCmd.MarkFlagRequired("cluster")
	getReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	getCmd.AddCommand(getReadReplicaCmd)

	createReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "The name of the cluster")
	createReadReplicaCmd.MarkFlagRequired("cluster")
	createReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	createCmd.AddCommand(createReadReplicaCmd)

	updateReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "The name of the cluster")
	updateReadReplicaCmd.MarkFlagRequired("cluster")
	updateReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	updateCmd.AddCommand(updateReadReplicaCmd)

	deleteReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "The name of the cluster")
	deleteReadReplicaCmd.MarkFlagRequired("cluster")
	deleteReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	deleteCmd.AddCommand(deleteReadReplicaCmd)
}
