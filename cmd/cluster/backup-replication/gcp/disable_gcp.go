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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var disableGcpCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable GCP backup replication for a cluster",
	Long:  "Disable GCP backup replication for all backup regions in the cluster",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to disable GCP backup replication for cluster: %s", ClusterName), viper.GetBool("force"))
		if err != nil {
			logrus.Fatal(err)
		}
	},
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

		backupRegions, err := getClusterBackupRegions(authApi, clusterId)
		if err != nil {
			logrus.Fatalf("%s\n", err)
		}

		regionTargets := make([]ybmclient.GcpBackupReplicationRegionTarget, 0, len(backupRegions))
		for region := range backupRegions {
			target := ybmclient.NewGcpBackupReplicationRegionTarget(region)
			regionTargets = append(regionTargets, *target)
		}

		err = modifyBackupReplicationAndDisplay(
			authApi,
			clusterId,
			regionTargets,
			"GCP backup replication for cluster %s is being disabled",
			"GCP backup replication has been disabled for cluster %s",
		)
		if err != nil {
			logrus.Fatalf("%s\n", err)
		}
	},
}

func init() {
	disableGcpCmd.Flags().BoolP("force", "f", false, "Bypass the confirmation prompt for non-interactive usage")
}
