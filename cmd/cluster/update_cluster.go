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

package cluster

import (
	"fmt"
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

// updateClusterCmd represents the cluster command
var updateClusterCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a cluster",
	Long:  "Update a cluster",
	Run: func(cmd *cobra.Command, args []string) {

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, r, err := authApi.GetCluster(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		originalSpec := resp.Data.GetSpec()
		trackID := originalSpec.SoftwareInfo.GetTrackId()
		trackName, err := authApi.GetTrackNameById(trackID)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		isNameChange := isNameUpdateOnly(cmd)
		populateFlags(cmd, originalSpec, trackName, authApi)

		regionInfoMapList := []map[string]string{}
		changedRegionInfo := cmd.Flags().Changed("region-info")

		if changedRegionInfo {
			regionInfoList, _ := cmd.Flags().GetStringArray("region-info")
			for _, regionInfoString := range regionInfoList {
				regionInfoMap := map[string]string{}
				for _, regionInfo := range strings.Split(regionInfoString, ",") {
					kvp := strings.Split(regionInfo, "=")
					if len(kvp) != 2 {
						logrus.Fatalln("Incorrect format in region info")
					}
					key := kvp[0]
					val := kvp[1]
					switch key {
					case "region":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["region"] = val
						}
					case "num-nodes":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["num-nodes"] = val
						}
					case "vpc":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["vpc"] = val
						}
					case "num-cores":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["num-cores"] = val
						}
					case "disk-size-gb":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["disk-size-gb"] = val
						}
					case "disk-iops":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["disk-iops"] = val
						}
					}
				}

				if _, ok := regionInfoMap["region"]; !ok {
					logrus.Fatalln("Region not specified in region info")
				}
				if _, ok := regionInfoMap["num-nodes"]; !ok {
					logrus.Fatalln("Number of nodes not specified in region info")
				}
				if _, ok := regionInfoMap["num-cores"]; !ok {
					logrus.Fatalln("Number of cores not specified in region info")
				}
				if _, ok := regionInfoMap["disk-size-gb"]; !ok {
					logrus.Fatalln("Disk size not specified in region info")
				}

				regionInfoMapList = append(regionInfoMapList, regionInfoMap)
			}
		}

		clusterSpec, err := authApi.EditClusterSpec(cmd, regionInfoMapList, clusterID)
		if err != nil {
			logrus.Fatalf("Error while creating cluster spec: %v", err)
		}

		clusterVersion := originalSpec.ClusterInfo.GetVersion()
		clusterSpec.ClusterInfo.SetVersion(clusterVersion)

		resp, r, err = authApi.EditCluster(clusterID).ClusterSpec(*clusterSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		clusterData := []ybmclient.ClusterData{resp.GetData()}

		msg := fmt.Sprintf("The cluster %s is being updated and will begin updating if there are changes", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") && !isNameChange {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_EDIT_CLUSTER, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The cluster %s has been updated\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.ListClusters().Name(clusterName).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			clusterData = respC.GetData()
		} else {
			fmt.Println(msg)
		}

		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}

		formatter.ClusterWrite(clustersCtx, clusterData)

	},
}

func init() {
	ClusterCmd.AddCommand(updateClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	updateClusterCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster.")
	updateClusterCmd.MarkFlagRequired("cluster-name")
	updateClusterCmd.Flags().String("new-name", "", "[OPTIONAL] The new name to be given to the cluster.")
	updateClusterCmd.Flags().String("cloud-provider", "", "[OPTIONAL] The cloud provider where database needs to be deployed. AWS, AZURE or GCP.")
	updateClusterCmd.Flags().String("cluster-type", "", "[OPTIONAL] Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED.")
	updateClusterCmd.Flags().StringArray("region-info", []string{}, `Region information for the cluster, provided as key-value pairs. Arguments are region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name>,num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,disk-iops=<disk-iops> (AWS only). region, num-nodes, num-cores, disk-size-gb are required. Specify one --region-info flag for each region in the cluster.`)
	updateClusterCmd.MarkFlagRequired("region-info")
	updateClusterCmd.Flags().String("cluster-tier", "", "[OPTIONAL] The tier of the cluster. Sandbox or Dedicated.")
	updateClusterCmd.Flags().String("fault-tolerance", "", "[OPTIONAL] Fault tolerance of the cluster. The possible values are NONE, NODE, ZONE, or REGION. Default NONE.")
	updateClusterCmd.Flags().String("database-version", "", "[OPTIONAL] The database version of the cluster. Production, Innovation, Preview, or 'Early Access'.")

}

func isNameUpdateOnly(cmd *cobra.Command) bool {
	return !cmd.Flags().Changed("cloud-provider") && !cmd.Flags().Changed("cluster-type") && !cmd.Flags().Changed("cluster-tier") &&
		!cmd.Flags().Changed("fault-tolerance") && !cmd.Flags().Changed("database-version") &&
		!cmd.Flags().Changed("region-info") && cmd.Flags().Changed("new-name")
}

func populateFlags(cmd *cobra.Command, originalSpec ybmclient.ClusterSpec, trackName string, authApi *ybmAuthClient.AuthApiClient) {
	if !cmd.Flags().Changed("cloud-provider") {
		cmd.Flag("cloud-provider").Value.Set(string(originalSpec.CloudInfo.GetCode()))
		cmd.Flag("cloud-provider").Changed = true
	}
	if !cmd.Flags().Changed("cluster-type") {
		cmd.Flag("cluster-type").Value.Set(string(originalSpec.ClusterInfo.GetClusterType()))
		cmd.Flag("cluster-type").Changed = true
	}
	if !cmd.Flags().Changed("cluster-tier") {
		clusterTier := string(originalSpec.ClusterInfo.GetClusterTier())
		clusterTierCli := "Sandbox"
		if clusterTier == "PAID" {
			clusterTierCli = "Dedicated"
		}
		cmd.Flag("cluster-tier").Value.Set(clusterTierCli)
		cmd.Flag("cluster-tier").Changed = true
	}
	if !cmd.Flags().Changed("fault-tolerance") {
		cmd.Flag("fault-tolerance").Value.Set(string(originalSpec.ClusterInfo.GetFaultTolerance()))
		cmd.Flag("fault-tolerance").Changed = true
	}
	if !cmd.Flags().Changed("database-version") {
		cmd.Flag("database-version").Value.Set(trackName)
		cmd.Flag("database-version").Changed = true
	}
	regionInfoList := []string{}
	if !cmd.Flags().Changed("region-info") {
		for _, clusterRegionInfo := range originalSpec.ClusterRegionInfo {
			regionInfo := ""
			if region, ok := clusterRegionInfo.PlacementInfo.CloudInfo.GetRegionOk(); ok && region != nil {
				regionInfo += "region=" + *region
			}
			if numNodes, ok := clusterRegionInfo.PlacementInfo.GetNumNodesOk(); ok && numNodes != nil {
				regionInfo += ",num-nodes=" + strconv.Itoa(int(*numNodes))
			}

			if vpcID, ok := clusterRegionInfo.PlacementInfo.GetVpcIdOk(); ok && vpcID != nil {
				vpcName, err := authApi.GetVpcNameById(*vpcID)
				if err != nil {
					logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
					return
				}
				regionInfo += ",vpc=" + vpcName
			}

			if nodeInfo, ok := clusterRegionInfo.GetNodeInfoOk(); ok {
				if numCores, ok_ := nodeInfo.GetNumCoresOk(); ok_ {
					regionInfo += ",num-cores=" + strconv.Itoa(int(*numCores))
				}
				if diskSizeGb, ok_ := nodeInfo.GetDiskSizeGbOk(); ok_ {
					regionInfo += ",disk-size-gb=" + strconv.Itoa(int(*diskSizeGb))
				}
				if diskIops, ok_ := nodeInfo.GetDiskIopsOk(); ok_ && diskIops != nil {
					regionInfo += ",disk-iops=" + strconv.Itoa(int(*diskIops))
				}
			}
			regionInfoList = append(regionInfoList, regionInfo)
		}
		cmd.Flags().Set("region-info", strings.Join(regionInfoList, " "))
		cmd.Flag("region-info").Changed = true

	}

}
