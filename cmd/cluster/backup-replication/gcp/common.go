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

package gcp

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

func getClusterBackupRegions(authApi *ybmAuthClient.AuthApiClient, clusterId string) (map[string]bool, error) {
	clusterResp, resp, err := authApi.GetCluster(clusterId).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", resp)
		return nil, fmt.Errorf("%s", ybmAuthClient.GetApiErrorDetails(err))
	}

	clusterData := clusterResp.GetData()
	clusterInfo := clusterData.GetInfo()
	regionInfoDetails := clusterInfo.GetClusterRegionInfoDetails()
	if len(regionInfoDetails) == 0 {
		return nil, fmt.Errorf("no backup regions found for cluster %s", ClusterName)
	}

	backupRegions := make(map[string]bool)
	for _, regionInfoDetail := range regionInfoDetails {
		if regionInfoDetail.GetBackupRegion() {
			region := regionInfoDetail.GetRegion()
			if region != "" {
				backupRegions[region] = true
			}
		}
	}

	if len(backupRegions) == 0 {
		return nil, fmt.Errorf("no backup regions found for cluster %s", ClusterName)
	}

	return backupRegions, nil
}

func parseAndBuildRegionTargets(cmd *cobra.Command, authApi *ybmAuthClient.AuthApiClient, clusterId string) ([]ybmclient.GcpBackupReplicationRegionTarget, error) {
	backupRegions, err := getClusterBackupRegions(authApi, clusterId)
	if err != nil {
		return nil, err
	}

	changedRegionTarget := cmd.Flags().Changed("region-target")
	if !changedRegionTarget {
		return nil, fmt.Errorf("--region-target must be specified for each backup region in the cluster")
	}

	regionTargetList, _ := cmd.Flags().GetStringArray("region-target")
	regionTargetMap := make(map[string]string)

	for _, regionTargetString := range regionTargetList {
		regionTargetMapItem := map[string]string{}
		for _, regionTarget := range strings.Split(regionTargetString, ",") {
			regionTarget = strings.TrimSpace(regionTarget)
			if regionTarget == "" {
				continue
			}
			kvp := strings.Split(regionTarget, "=")
			if len(kvp) != 2 {
				return nil, fmt.Errorf("incorrect format in region target")
			}
			key := strings.TrimSpace(kvp[0])
			val := strings.TrimSpace(kvp[1])
			switch key {
			case "region":
				if len(val) != 0 {
					regionTargetMapItem["region"] = val
				}
			case "bucket-name":
				if len(val) != 0 {
					regionTargetMapItem["bucket-name"] = val
				}
			}
		}

		if _, ok := regionTargetMapItem["region"]; !ok {
			return nil, fmt.Errorf("region not specified in region target")
		}
		if _, ok := regionTargetMapItem["bucket-name"]; !ok {
			return nil, fmt.Errorf("bucket name not specified in region target")
		}

		region := regionTargetMapItem["region"]
		bucketName := regionTargetMapItem["bucket-name"]

		if _, exists := backupRegions[region]; !exists {
			return nil, fmt.Errorf("region '%s' is not a backup region in cluster %s", region, ClusterName)
		}

		if _, duplicate := regionTargetMap[region]; duplicate {
			return nil, fmt.Errorf("duplicate region target provided for region '%s'", region)
		}

		regionTargetMap[region] = bucketName
	}

	if len(regionTargetMap) != len(backupRegions) {
		missingRegions := []string{}
		for region := range backupRegions {
			if _, provided := regionTargetMap[region]; !provided {
				missingRegions = append(missingRegions, region)
			}
		}
		return nil, fmt.Errorf("--region-target must be provided for all backup regions. Missing regions: %s", strings.Join(missingRegions, ", "))
	}

	regionalTargets := make([]ybmclient.GcpBackupReplicationRegionTarget, 0, len(regionTargetMap))
	for region, bucketName := range regionTargetMap {
		target := ybmclient.NewGcpBackupReplicationRegionTarget(region)
		target.SetTarget(bucketName)
		regionalTargets = append(regionalTargets, *target)
	}

	return regionalTargets, nil
}

func modifyBackupReplicationAndDisplay(
	authApi *ybmAuthClient.AuthApiClient,
	clusterId string,
	regionalTargets []ybmclient.GcpBackupReplicationRegionTarget,
	operationMsg string,
	successMsg string,
) error {
	spec := ybmclient.NewGcpBackupReplicationSpec(regionalTargets)

	_, r, err := authApi.ModifyGcpBackupReplication(clusterId).GcpBackupReplicationSpec(*spec).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		return fmt.Errorf("%s", ybmAuthClient.GetApiErrorDetails(err))
	}

	msg := fmt.Sprintf(operationMsg, formatter.Colorize(ClusterName, formatter.GREEN_COLOR))
	if viper.GetBool("wait") {
		returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_MODIFY_GCP_BACKUP_REPLICATION, []string{"FAILED", "SUCCEEDED"}, msg)
		if err != nil {
			return fmt.Errorf("error when getting task status: %s", err)
		}
		if returnStatus != "SUCCEEDED" {
			return fmt.Errorf("operation failed with error: %s", returnStatus)
		}
		fmt.Printf(successMsg+"\n\n", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))
	} else {
		fmt.Println(msg)
	}

	updatedConfigResp, resp, err := authApi.GetGcpBackupReplicationConfig(clusterId).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", resp)
		return fmt.Errorf("%s", ybmAuthClient.GetApiErrorDetails(err))
	}

	backupReplicationCtx := formatter.Context{
		Output: os.Stdout,
		Format: formatter.Format(viper.GetString("output")),
	}

	err = formatter.BackupReplicationWrite(backupReplicationCtx, updatedConfigResp.GetData(), false)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}
