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
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var syncGcpCmd = &cobra.Command{
	Use:   "sync",
	Short: "Trigger resync of GCP backup replication for a cluster",
	Long:  "Trigger resync of backup data for a cluster. Creates one-time transfer operations for existing transfer jobs to synchronize backup data to target GCS buckets.",
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

		spec := ybmclient.NewResyncGcpBackupReplicationSpec()

		r, err := authApi.TriggerGcpBackupReplicationResync(clusterId).ResyncGcpBackupReplicationSpec(*spec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Resync triggered for all backup replication configs in cluster %s\n", formatter.Colorize(ClusterName, formatter.GREEN_COLOR))
	},
}
