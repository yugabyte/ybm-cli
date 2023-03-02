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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, r, err := authApi.GetCluster(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.GetCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		originalSpec := resp.Data.GetSpec()
		trackID := originalSpec.SoftwareInfo.GetTrackId()
		trackName, err := authApi.GetTrackNameById(trackID)
		if err != nil {
			logrus.Fatalf("Error when calling `getTrackName`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		populateFlags(cmd, originalSpec, trackName, authApi)

		regionInfoMapList := []map[string]string{}
		if cmd.Flags().Changed("region-info") {
			regionInfoList, _ := cmd.Flags().GetStringArray("region-info")
			regionInfoList = strings.Split(regionInfoList[0], "|")
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
					}
				}

				if _, ok := regionInfoMap["region"]; !ok {
					logrus.Fatalln("Region not specified in region info")
				}
				if _, ok := regionInfoMap["num-nodes"]; !ok {
					logrus.Fatalln("Number of nodes not specified in region info")
				}

				regionInfoMapList = append(regionInfoMapList, regionInfoMap)
			}
		}

		if cmd.Flags().Changed("node-config") {
			nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
			if _, ok := nodeConfig["num-cores"]; !ok {
				logrus.Fatal("Number of cores not specified in node config\n")
			}
		}

		clusterSpec, err := authApi.CreateClusterSpec(cmd, regionInfoMapList)
		if err != nil {
			logrus.Fatalf("Error while creating cluster spec: %v", err)
		}

		clusterVersion := originalSpec.ClusterInfo.GetVersion()
		clusterSpec.ClusterInfo.SetVersion(clusterVersion)

		resp, r, err = authApi.EditCluster(clusterID).ClusterSpec(*clusterSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.UpdateCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		clusterData := []ybmclient.ClusterData{resp.GetData()}

		msg := fmt.Sprintf("The cluster %s is being updated", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, "EDIT_CLUSTER", []string{"FAILED", "SUCCEEDED"}, msg, 1800)
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
				logrus.Fatalf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
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
	updateClusterCmd.Flags().String("cloud-type", "", "[OPTIONAL] The cloud provider where database needs to be deployed. AWS or GCP.")
	updateClusterCmd.Flags().String("cluster-type", "", "[OPTIONAL] Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED.")
	updateClusterCmd.Flags().StringToInt("node-config", nil, "[OPTIONAL] Configuration of the cluster nodes. Please provide key value pairs num-cores=<num-cores>,disk-size-gb=<disk-size-gb> as the value.  num-cores is mandatory, disk-size-gb is optional.")
	updateClusterCmd.Flags().StringArray("region-info", []string{}, `[OPTIONAL] Region information for the cluster. Please provide key value pairs, region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name> as the value. If provided, region and num-nodes are mandatory, vpc is optional.`)
	updateClusterCmd.Flags().String("cluster-tier", "", "[OPTIONAL] The tier of the cluster. Sandbox or Dedicated.")
	updateClusterCmd.Flags().String("fault-tolerance", "", "[OPTIONAL] The fault tolerance of the cluster. The possible values are NONE, ZONE and REGION.")
	updateClusterCmd.Flags().String("database-version", "", "[OPTIONAL] The database version of the cluster. Stable or Preview.")

}

func populateFlags(cmd *cobra.Command, originalSpec ybmclient.ClusterSpec, trackName string, authApi *ybmAuthClient.AuthApiClient) {
	if !cmd.Flags().Changed("cloud-type") {
		cmd.Flag("cloud-type").Value.Set(string(originalSpec.CloudInfo.GetCode()))
		cmd.Flag("cloud-type").Changed = true
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
	if !cmd.Flags().Changed("node-config") {
		nodeConfig := ""
		if diskSizeGb, ok := originalSpec.ClusterInfo.NodeInfo.GetDiskSizeGbOk(); ok {
			nodeConfig += "disk-size-gb=" + strconv.Itoa(int(*diskSizeGb))
		}
		if numCores, ok := originalSpec.ClusterInfo.NodeInfo.GetNumCoresOk(); ok {
			nodeConfig += ",num-cores=" + strconv.Itoa(int(*numCores))
		}
		cmd.Flag("node-config").Value.Set(nodeConfig)
		cmd.Flag("node-config").Changed = true

	}
	regionInfoList := ""
	numRegions := len(originalSpec.ClusterRegionInfo)
	if !cmd.Flags().Changed("region-info") {
		for index, clusterRegionInfo := range originalSpec.ClusterRegionInfo {
			regionInfo := ""
			if region, ok := clusterRegionInfo.PlacementInfo.CloudInfo.GetRegionOk(); ok && region != nil {
				regionInfo += "region=" + *region
			}
			//logrus.Errorln(clusterRegionInfo.PlacementInfo.GetNumNodes())
			if numNodes, ok := clusterRegionInfo.PlacementInfo.GetNumNodesOk(); ok && numNodes != nil {
				regionInfo += ",num-nodes=" + strconv.Itoa(int(*numNodes))
			}

			if vpcID, ok := clusterRegionInfo.PlacementInfo.GetVpcIdOk(); ok && vpcID != nil {
				vpcName, err := authApi.GetVpcNameById(*vpcID)
				if err != nil {
					logrus.Fatalf("Error when calling `getVpcName`: %s", ybmAuthClient.GetApiErrorDetails(err))
					return
				}
				regionInfo += ",vpc=" + vpcName
			}
			regionInfoList += regionInfo
			if index < numRegions-1 {
				regionInfoList += "|"
			}
		}
		cmd.Flag("region-info").Value.Set(regionInfoList)
		cmd.Flag("region-info").Changed = true

	}

}
