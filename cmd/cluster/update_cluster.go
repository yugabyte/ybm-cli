/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
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
	Short: "Update a cluster in YB Managed",
	Long:  "Update a cluster in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {

		clusterName, _ := cmd.Flags().GetString("cluster-name")
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
		resp, r, err := authApi.GetCluster(clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.GetCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		originalSpec := resp.Data.GetSpec()
		trackID := originalSpec.SoftwareInfo.GetTrackId()
		trackName, err := authApi.GetTrackNameById(trackID)
		if err != nil {
			logrus.Errorf("Error when calling `getTrackName`: %s", ybmAuthClient.GetApiErrorDetails(err))
			return
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
						logrus.Errorln("Incorrect format in region info")
						return
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
					logrus.Errorln("Region not specified in region info")
					return
				}
				if _, ok := regionInfoMap["num-nodes"]; !ok {
					logrus.Errorln("Number of nodes not specified in region info")
					return
				}

				regionInfoMapList = append(regionInfoMapList, regionInfoMap)
			}
		}

		if cmd.Flags().Changed("node-config") {
			nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
			if _, ok := nodeConfig["num-cores"]; !ok {
				logrus.Error("Number of cores not specified in node config\n")
				return
			}
		}

		clusterSpec, err := authApi.CreateClusterSpec(cmd, regionInfoMapList)
		if err != nil {
			logrus.Errorf("Error while creating cluster spec: %v", err)
			return
		}

		clusterVersion := originalSpec.ClusterInfo.GetVersion()
		clusterSpec.ClusterInfo.SetVersion(clusterVersion)

		resp, r, err = authApi.EditCluster(clusterID).ClusterSpec(*clusterSpec).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `ClusterApi.UpdateCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}
		clusterData := []ybmclient.ClusterData{resp.GetData()}

		msg := fmt.Sprintf("The cluster %s is being updated", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "EDIT_CLUSTER", []string{"FAILED", "SUCCEEDED"}, msg, 1500)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
			}
			fmt.Printf("The cluster %s has been updated\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.ListClusters().Name(clusterName).Execute()
			if err != nil {
				logrus.Errorf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
				logrus.Debugf("Full HTTP response: %v", r)
				return
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
	updateClusterCmd.Flags().String("cluster-name", "", "Name of the cluster.")
	updateClusterCmd.MarkFlagRequired("cluster-name")
	updateClusterCmd.Flags().String("cloud-type", "", "The cloud provider where database needs to be deployed. AWS or GCP.")
	updateClusterCmd.Flags().String("cluster-type", "", "Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED.")
	updateClusterCmd.Flags().StringToInt("node-config", nil, "Configuration of the cluster nodes. Please provide key value pairs num-cores=<num-cores>,disk-size-gb=<disk-size-gb> as the value.  num-cores is mandatory, disk-size-gb is optional.")
	updateClusterCmd.Flags().StringArray("region-info", []string{}, `Region information for the cluster. Please provide key value pairs
	region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name> as the value. region and num-nodes are mandatory, vpc is optional.`)
	updateClusterCmd.Flags().String("cluster-tier", "", "The tier of the cluster. FREE or PAID.")
	updateClusterCmd.Flags().String("fault-tolerance", "", "The fault tolerance of the cluster. The possible values are NONE, ZONE and REGION.")
	updateClusterCmd.Flags().String("database-track", "", "The database track of the cluster. Stable or Preview.")

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
					logrus.Errorf("Error when calling `getVpcName`: %s", ybmAuthClient.GetApiErrorDetails(err))
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
