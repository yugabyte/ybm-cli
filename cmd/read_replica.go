/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var clusterName string
var allReplicaOpt []string

// Parse array of read replica string to string params
func parseReplicaOpts(authApi *ybmAuthClient.AuthApiClient, replicaOpts []string) ([]ybmclient.ReadReplicaSpec, error) {
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
				//Avoid potential integer overflow see gosec
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					spec.NodeInfo.NumCores = int32(n)
				}
			case "memory_mb":
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					spec.NodeInfo.MemoryMb = int32(n)
				}
			case "disk_size_gb":
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					spec.NodeInfo.DiskSizeGb = int32(n)
				}
			case "code":
				spec.PlacementInfo.CloudInfo.Code = ybmclient.CloudEnum(val)
			case "region":
				spec.PlacementInfo.CloudInfo.Region = val
			case "num_nodes":
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					spec.PlacementInfo.NumNodes = int32(n)
				}
			case "vpc":
				vpcName := val
				vpcID, err := authApi.GetVpcIdByName(vpcName)
				if err != nil {
					logrus.Error(err)
					return nil, err
				}
				spec.PlacementInfo.VpcId = *ybmclient.NewNullableString(&vpcID)
			case "num_replicas":
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					numReplicas := int32(n)
					spec.PlacementInfo.NumReplicas = *ybmclient.NewNullableInt32(&numReplicas)
				}
			case "multi_zone":
				isMultiZone, _ := strconv.ParseBool(val)
				spec.PlacementInfo.MultiZone = *ybmclient.NewNullableBool(&isMultiZone)
			}

		}
		readReplicaSpecs = append(readReplicaSpecs, spec)
	}
	return readReplicaSpecs, nil
}

var getReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Get read replica in YugabyteDB Managed",
	Long:  "Get read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}
		resp, r, err := authApi.ListReadReplicas(clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ReadReplicaApi.ListReadReplicas`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}
		prettyPrintJson(resp)
	},
}

var createReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Create read replica in YugabyteDB Managed",
	Long:  "Create read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}
		readReplicaSpecs, err := parseReplicaOpts(authApi, allReplicaOpt)
		if err != nil {
			logrus.Errorf("Error while parsing read replica options: %v", err)
			return
		}

		resp, r, err := authApi.CreateReadReplica(clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ReadReplicaApi.CreateReadReplica`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var updateReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Edit read replica in YugabyteDB Managed",
	Long:  "Edit read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}
		readReplicaSpecs, err := parseReplicaOpts(authApi, allReplicaOpt)
		if err != nil {
			logrus.Errorf("Error while parsing read replica options: %v", err)
			return
		}

		resp, r, err := authApi.EditReadReplicas(clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ReadReplicaApi.EditReadReplicas`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var deleteReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "Delete read replica from YugabyteDB Managed",
	Long:  "Delete read replica from YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}
		r, err := authApi.DeleteReadReplica(clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ReadReplicaApi.DeleteReadReplica`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}
		fmt.Printf("All read replica sucessfully deleted for cluster %s \n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

	},
}

func init() {
	getReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	getReadReplicaCmd.MarkFlagRequired("cluster-name")
	getReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	getCmd.AddCommand(getReadReplicaCmd)

	createReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	createReadReplicaCmd.MarkFlagRequired("cluster-name")
	createReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	createCmd.AddCommand(createReadReplicaCmd)

	updateReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	updateReadReplicaCmd.MarkFlagRequired("cluster-name")
	updateReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	updateCmd.AddCommand(updateReadReplicaCmd)

	deleteReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	deleteReadReplicaCmd.MarkFlagRequired("cluster-name")
	deleteReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)
	deleteCmd.AddCommand(deleteReadReplicaCmd)
}
