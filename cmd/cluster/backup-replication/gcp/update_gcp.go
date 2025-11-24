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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

var updateGcpCmd = &cobra.Command{
	Use:   "update",
	Short: "Update GCP backup replication configuration for a cluster",
	Long:  "Update GCP backup replication configuration for all backup regions in the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterId, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		regionalTargets, err := parseAndBuildRegionTargets(cmd, authApi, clusterId)
		if err != nil {
			logrus.Fatalf("%s\n", err)
		}

		err = modifyBackupReplicationAndDisplay(
			authApi,
			clusterId,
			regionalTargets,
			"GCP backup replication for cluster %s is being updated",
			"GCP backup replication has been updated for cluster %s",
		)
		if err != nil {
			logrus.Fatalf("%s\n", err)
		}
	},
}

func init() {
	updateGcpCmd.Flags().StringArray("region-target", []string{}, `[REQUIRED] Specify region and bucket pairs for backup replication. Format: region=<region-name>,bucket-name=<bucket-name>. Must be specified for each backup region in the cluster.`)
	updateGcpCmd.MarkFlagRequired("region-target")
}
