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

func getClusterId(apiClient ybmclient.APIClient, name string, accountID string, projectID string) string {
	cResp, cR, cErr := apiClient.ClusterApi.ListClusters(context.Background(), accountID, projectID).Name(clusterName).Execute()
	if cErr != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.ListClusters`: %v\n", cErr)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", cR)
	}
	cRespData := cResp.Data
	if len(cRespData) == 0 {
		fmt.Fprintf(os.Stderr, "Unable to find Cluster with name %v. Error when calling `ClusterApi.ListClusters`: %v\n", clusterName, cErr)
		return ""
	}
	return cRespData[0].Info.Id
}

var getReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Get read replica in YugabyteDB Managed",
	Long:  "Get read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		clusterID := getClusterId(*apiClient, clusterName, accountID, projectID)
		if clusterID == "" {
			return
		}

		resp, r, err := apiClient.ReadReplicaApi.ListReadReplicas(context.Background(), accountID, projectID, clusterID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListNetworkAllowLists`: %v\n", err)
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
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		clusterID := getClusterId(*apiClient, clusterName, accountID, projectID)
		if clusterID == "" {
			return
		}

		// var replicaData []map[string]string
		readReplicaSpecs := []ybmclient.ReadReplicaSpec{}

		for _, replicaOpt := range allReplicaOpt {
			// replicaMap := make(map[string]string)
			// replicaData = append(replicaData, replicaMap)
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
			// ./ybm create read_replica -r=num_cores=2,memory_mb=4096,disk_size_gb=10,code=GCP,region=us-west-1,num_nodes=1,vpc_id=7071b608-cc5e-4353-9321-cdeafd3ea2be
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

		fmt.Println(len(allReplicaOpt))
		fmt.Println(allReplicaOpt)

		fmt.Println(len(readReplicaSpecs))

		fmt.Printf("%+v\n", readReplicaSpecs)

		resp, r, err := apiClient.ReadReplicaApi.CreateReadReplica(context.Background(), accountID, projectID, clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ReadReplicaApi.CreateReadReplica``: %v\n", err)
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
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		readResp, readResponse, readErr := apiClient.NetworkApi.ListNetworkAllowLists(context.Background(), accountID, projectID).Execute()

		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.ListNetworkAllowLists`: %v\n", readErr)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", readResponse)
			return
		}

		allowList, findErr := findNetworkAllowList(readResp.Data, nalName)

		if findErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", findErr)
			return
		}

		r, err := apiClient.ReadReplicaApi.DeleteReadReplica(context.Background(), accountID, projectID, allowList.Info.Id).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `NetworkApi.DeleteNetworkAllowList``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		fmt.Fprintf(os.Stdout, "Success: NetworkAllosList <%s> deleted\n", nalName)
	},
}

func init() {
	// getNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "The name of the Network Allow List")
	// getCmd.AddCommand(getNetworkAllowListCmd)
	createReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "The name of the cluster")
	createNetworkAllowListCmd.MarkFlagRequired("cluster")

	createReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, "Options for a replica")
	createCmd.AddCommand(createReadReplicaCmd)

	createReadReplicaCmd.Flags().StringToStringVarP(&test, "region-info", "a", nil, `Region information for the cluster. Please provide key value pairs
	region=<region-name>,num_nodes=<number-of-nodes>,vpc_id=<vpc-id> as the value. region and num_nodes are mandatory, vpc_id is optional.`)
	// createNetworkAllowListCmd.Flags() // StringVarP(&nalName, "name", "n", "", "The name of the Network Allow List")

	// createNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "The name of the Network Allow List")
	// createNetworkAllowListCmd.MarkFlagRequired("name")
	// createNetworkAllowListCmd.Flags().StringVarP(&nalDescription, "description", "d", "", "Description of the Network Allow List")
	// createNetworkAllowListCmd.Flags().StringSliceVarP(&nalIpAddrs, "ip_addr", "i", []string{}, "IP addresses included in the Network Allow List")
	// createNetworkAllowListCmd.MarkFlagRequired("ip_addr")
	// createCmd.AddCommand(createNetworkAllowListCmd)

	// deleteNetworkAllowListCmd.Flags().StringVarP(&nalName, "name", "n", "", "The name of the Network Allow List")
	// deleteNetworkAllowListCmd.MarkFlagRequired("name")
	// deleteCmd.AddCommand(deleteNetworkAllowListCmd)
}
