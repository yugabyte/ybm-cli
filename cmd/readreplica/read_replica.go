/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package readreplica

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var clusterName string
var allReplicaOpt []string

var ReadReplicaCmd = &cobra.Command{
	Use:   "read_replica",
	Short: "read_replica",
	Long:  "Read Replica commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func getDefaultSpec(primaryClusterCloud ybmclient.CloudEnum, vpcId string) ybmclient.ReadReplicaSpec {
	n := int32(1)
	numReplicas := ybmclient.NewNullableInt32(&n)
	spec := ybmclient.ReadReplicaSpec{
		NodeInfo: ybmclient.ClusterNodeInfo{
			NumCores: 2,
		},
		PlacementInfo: ybmclient.PlacementInfo{
			CloudInfo: ybmclient.CloudInfo{
				Code:   primaryClusterCloud,
				Region: "us-west2",
			},
			VpcId:       *ybmclient.NewNullableString(&vpcId),
			NumNodes:    1,
			NumReplicas: *numReplicas,
		},
	}

	return spec
}

func setMemoryAndDisk(authApi *ybmAuthClient.AuthApiClient, spec *ybmclient.ReadReplicaSpec) error {
	cloud := string(spec.PlacementInfo.CloudInfo.Code)
	tier := "PAID"
	region := spec.PlacementInfo.CloudInfo.Region
	numCores := spec.NodeInfo.NumCores
	memoryMb, err := authApi.GetFromInstanceType("memory", cloud, tier, region, numCores)
	if err != nil {
		return err
	}
	spec.NodeInfo.MemoryMb = memoryMb
	if spec.NodeInfo.DiskSizeGb == 0 {
		diskSizeGb, err := authApi.GetFromInstanceType("disk", cloud, tier, region, numCores)
		if err != nil {
			return err
		}
		spec.NodeInfo.DiskSizeGb = diskSizeGb
	}
	return nil
}

// Parse array of read replica string to string params
func parseReplicaOpts(authApi *ybmAuthClient.AuthApiClient, replicaOpts []string, primaryClusterCloud ybmclient.CloudEnum, vpcId string) ([]ybmclient.ReadReplicaSpec, error) {
	readReplicaSpecs := []ybmclient.ReadReplicaSpec{}

	defaultSpec := getDefaultSpec(primaryClusterCloud, vpcId)

	for _, replicaOpt := range replicaOpts {

		spec := getDefaultSpec(primaryClusterCloud, vpcId)

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
			case "disk_size_gb":
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					spec.NodeInfo.DiskSizeGb = int32(n)
				}
			case "code":
				if string(primaryClusterCloud) != val {
					return nil, fmt.Errorf("all the read replicas must be in the same cloud provider as the primary cluster")
				}
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
		if err := setMemoryAndDisk(authApi, &spec); err != nil {
			return nil, err
		}
		readReplicaSpecs = append(readReplicaSpecs, spec)
	}

	if len(readReplicaSpecs) == 0 {
		if err := setMemoryAndDisk(authApi, &defaultSpec); err != nil {
			return nil, err
		}
		readReplicaSpecs = append(readReplicaSpecs, defaultSpec)
	}

	return readReplicaSpecs, nil
}

func printReadReplicaOutput(resp ybmclient.ReadReplicaListResponse) {
	readReplicaCtx := formatter.Context{
		Output: os.Stdout,
		Format: formatter.NewReadReplicaFormat(viper.GetString("output")),
	}

	formatter.ReadReplicaWrite(readReplicaCtx, resp.Data.GetSpec(), resp.Data.Info.GetEndpoints())
}

var getReadReplicaCmd = &cobra.Command{
	Use:   "get",
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

		printReadReplicaOutput(resp)

	},
}

var createReadReplicaCmd = &cobra.Command{
	Use:   "create",
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

		vpcId, err := authApi.GetClusterVpcById(clusterID)
		if err != nil {
			logrus.Errorf("Error while fetching the VPC ID of the primary cluster: %s\n", ybmAuthClient.GetApiErrorDetails(err))
			return
		}
		if vpcId == "" {
			logrus.Error("The cluster must be deployed in a dedicated VPC to create read replicas")
			return
		}

		primaryClusterCloud, err := authApi.GetClusterCloudById(clusterID)
		if err != nil {
			logrus.Errorf("Error while fetching the cloud provider of the primary cluster: %s\n", ybmAuthClient.GetApiErrorDetails(err))
			return
		}

		readReplicaSpecs, err := parseReplicaOpts(authApi, allReplicaOpt, primaryClusterCloud, vpcId)
		if err != nil {
			logrus.Errorf("Error while parsing read replica options: %s", ybmAuthClient.GetApiErrorDetails(err))
			return
		}

		resp, r, err := authApi.CreateReadReplica(clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ReadReplicaApi.CreateReadReplica`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		printReadReplicaOutput(resp)
	},
}

var updateReadReplicaCmd = &cobra.Command{
	Use:   "update",
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
		vpcId, err := authApi.GetClusterVpcById(clusterID)
		if err != nil {
			logrus.Errorf("Error while fetching the VPC ID of the primary cluster: %s\n", ybmAuthClient.GetApiErrorDetails(err))
			return
		}
		if vpcId == "" {
			logrus.Error("The cluster must be deployed in a dedicated VPC to create read replicacs")
			return
		}
		primaryClusterCloud, err := authApi.GetClusterCloudById(clusterID)
		if err != nil {
			logrus.Errorf("Error while fetching the cloud provider of the primary cluster: %s\n", ybmAuthClient.GetApiErrorDetails(err))
			return
		}
		readReplicaSpecs, err := parseReplicaOpts(authApi, allReplicaOpt, primaryClusterCloud, vpcId)
		if err != nil {
			logrus.Errorf("Error while parsing read replica options: %s", ybmAuthClient.GetApiErrorDetails(err))
			return
		}

		resp, r, err := authApi.EditReadReplicas(clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ReadReplicaApi.EditReadReplicas`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		printReadReplicaOutput(resp)
	},
}

var deleteReadReplicaCmd = &cobra.Command{
	Use:   "delete",
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
		fmt.Printf("All read replica sucessfully scheduled for deletion for cluster %s \n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

	},
}

func init() {
	ReadReplicaCmd.AddCommand(getReadReplicaCmd)
	getReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	getReadReplicaCmd.MarkFlagRequired("cluster-name")

	ReadReplicaCmd.AddCommand(createReadReplicaCmd)
	createReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	createReadReplicaCmd.MarkFlagRequired("cluster-name")
	createReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region_num_cores>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc=<vpc_name>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)

	ReadReplicaCmd.AddCommand(updateReadReplicaCmd)
	updateReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	updateReadReplicaCmd.MarkFlagRequired("cluster-name")
	updateReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `Region information for the cluster. Please provide key value pairs num_cores=<region_num_cores>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc=<vpc_name>,num_replicas=<num_replicas>,multi_zone=<multi_zone>`)

	ReadReplicaCmd.AddCommand(deleteReadReplicaCmd)
	deleteReadReplicaCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "The name of the cluster")
	deleteReadReplicaCmd.MarkFlagRequired("cluster-name")

}
