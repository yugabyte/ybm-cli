// Licensed to Yugabyte, Inc. under one or more contributor license
// agreements. See the NOTICE file distributed with this work for
// additional information regarding copyright ownership. Yugabyte
// licenses this file to you under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

var ClusterName string
var allReplicaOpt []string

var ReadReplicaCmd = &cobra.Command{
	Use:   "read-replica",
	Short: "Manage Read Replicas",
	Long:  "Manage Read Replicas",
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
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			switch key {
			case "num-cores":
				//Avoid potential integer overflow see gosec
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					spec.NodeInfo.NumCores = int32(n)
				}
			case "disk-size-gb":
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
			case "num-nodes":
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
			case "num-replicas":
				if n > 0 && n <= math.MaxInt32 {
					/* #nosec G109 */
					numReplicas := int32(n)
					spec.PlacementInfo.NumReplicas = *ybmclient.NewNullableInt32(&numReplicas)
				}
			case "multi-zone":
				isMultiZone, err := strconv.ParseBool(val)
				if err != nil {
					return nil, err
				}
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
	Short: "Get read replica",
	Long:  "Get read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		listReadReplicaCmd.Run(cmd, args)
		logrus.Warnln("\nThe command `ybm cluster read-replica get` is deprecated. Please use `ybm cluster read-replica list` instead.")
	},
}

var listReadReplicaCmd = &cobra.Command{
	Use:   "list",
	Short: "List read replicas",
	Long:  "List read replicas in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, r, err := authApi.ListReadReplicas(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ReadReplicaApi.ListReadReplicas`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		printReadReplicaOutput(resp)

	},
}

var createReadReplicaCmd = &cobra.Command{
	Use:   "create",
	Short: "Create read replica",
	Long:  "Create read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		vpcId, err := authApi.GetClusterVpcById(clusterID)
		if err != nil {
			logrus.Fatalf("Error while fetching the VPC ID of the primary cluster: %s\n", ybmAuthClient.GetApiErrorDetails(err))
		}
		if vpcId == "" {
			logrus.Fatal("The cluster must be deployed in a dedicated VPC to create read replicas")
		}

		primaryClusterCloud, err := authApi.GetClusterCloudById(clusterID)
		if err != nil {
			logrus.Fatalf("Error while fetching the cloud provider of the primary cluster: %s\n", ybmAuthClient.GetApiErrorDetails(err))
		}

		readReplicaSpecs, err := parseReplicaOpts(authApi, allReplicaOpt, primaryClusterCloud, vpcId)
		if err != nil {
			logrus.Fatalf("Error while parsing read replica options: %s", ybmAuthClient.GetApiErrorDetails(err))
			return
		}

		resp, r, err := authApi.CreateReadReplica(clusterID).ReadReplicaSpec(readReplicaSpecs).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ReadReplicaApi.CreateReadReplica`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("Read Replica is being created for cluster %s", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_CREATE_READ_REPLICA, []string{"FAILED", "SUCCEEDED"}, msg, 1800)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("Read Replica has been created for cluster %s.\n", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

			resp, r, err = authApi.ListReadReplicas(clusterID).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Error when calling `ReadReplicaApi.ListReadReplicas`: %s", ybmAuthClient.GetApiErrorDetails(err))
			}
		} else {
			fmt.Println(msg)
		}
		printReadReplicaOutput(resp)
	},
}

var updateReadReplicaCmd = &cobra.Command{
	Use:   "update",
	Short: "Edit read replica",
	Long:  "Edit read replica in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatal(err)
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
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ReadReplicaApi.EditReadReplicas`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		msg := fmt.Sprintf("Read Replica is being updated for cluster %s", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_EDIT_READ_REPLICA, []string{"FAILED", "SUCCEEDED"}, msg, 1800)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("Read Replica has been updated for cluster %s.\n", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

			resp, r, err = authApi.ListReadReplicas(clusterID).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf("Error when calling `ReadReplicaApi.ListReadReplicas`: %s", ybmAuthClient.GetApiErrorDetails(err))
			}
		} else {
			fmt.Println(msg)
		}
		printReadReplicaOutput(resp)
	},
}

var deleteReadReplicaCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete read replica",
	Long:  "Delete read replica from YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		r, err := authApi.DeleteReadReplica(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ReadReplicaApi.DeleteReadReplica`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		msg := fmt.Sprintf("Read Replica is being deleted for cluster %s", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_DELETE_READ_REPLICA, []string{"FAILED", "SUCCEEDED"}, msg, 1800)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("All Read Replica has been deleted for cluster %s.\n", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))
			return
		}
		fmt.Printf("All Read Replica has been deleted for cluster %s.\n", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

	},
}

func init() {
	ReadReplicaCmd.AddCommand(getReadReplicaCmd)

	ReadReplicaCmd.AddCommand(listReadReplicaCmd)

	ReadReplicaCmd.AddCommand(createReadReplicaCmd)
	createReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `[OPTIONAL] Region information for the cluster. Please provide key value pairs num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,code=<GCP or AWS>,region=<region>,num-nodes=<num-nodes>,vpc=<vpc-name>,num-replicas=<num-replicas>,multi-zone=<multi-zone>.`)

	ReadReplicaCmd.AddCommand(updateReadReplicaCmd)
	updateReadReplicaCmd.Flags().StringArrayVarP(&allReplicaOpt, "replica", "r", []string{}, `[OPTIONAL] Region information for the cluster. Please provide key value pairs num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,code=<GCP or AWS>,region=<region>,num-nodes=<num-nodes>,vpc=<vpc-name>,num-replicas=<num-replicas>,multi-zone=<multi-zone>.`)

	ReadReplicaCmd.AddCommand(deleteReadReplicaCmd)
}
