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
	"github.com/spf13/cobra"
	audit_log_exporter "github.com/yugabyte/ybm-cli/cmd/cluster/audit-log-exporter"
	backupreplication "github.com/yugabyte/ybm-cli/cmd/cluster/backup-replication"
	"github.com/yugabyte/ybm-cli/cmd/cluster/cert"
	connectionpooling "github.com/yugabyte/ybm-cli/cmd/cluster/connection-pooling"
	encryption "github.com/yugabyte/ybm-cli/cmd/cluster/encryption"
	log_exporter "github.com/yugabyte/ybm-cli/cmd/cluster/log-exporter"
	"github.com/yugabyte/ybm-cli/cmd/cluster/namespace"
	"github.com/yugabyte/ybm-cli/cmd/cluster/network"
	"github.com/yugabyte/ybm-cli/cmd/cluster/node"
	pitrconfig "github.com/yugabyte/ybm-cli/cmd/cluster/pitr-config"
	readreplica "github.com/yugabyte/ybm-cli/cmd/cluster/read-replica"
	"github.com/yugabyte/ybm-cli/cmd/util"
)

// getCmd represents the list command
var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster operations",
	Long:  "Manage cluster operations",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	ClusterCmd.AddCommand(cert.CertCmd)

	ClusterCmd.AddCommand(log_exporter.DbQueryLoggingCmd)
	log_exporter.DbQueryLoggingCmd.PersistentFlags().StringVarP(&log_exporter.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	log_exporter.DbQueryLoggingCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(audit_log_exporter.DbAuditLoggingCmd)
	audit_log_exporter.DbAuditLoggingCmd.PersistentFlags().StringVarP(&audit_log_exporter.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	audit_log_exporter.DbAuditLoggingCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(network.NetworkCmd)
	network.NetworkCmd.PersistentFlags().StringVarP(&network.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	network.NetworkCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(readreplica.ReadReplicaCmd)
	readreplica.ReadReplicaCmd.PersistentFlags().StringVarP(&readreplica.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	readreplica.ReadReplicaCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(node.NodeCmd)
	node.NodeCmd.PersistentFlags().StringP("cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	node.NodeCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(encryption.EncryptionCmd)
	encryption.EncryptionCmd.PersistentFlags().StringP("cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	encryption.EncryptionCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(namespace.NamespaceCmd)
	namespace.NamespaceCmd.PersistentFlags().StringP("cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	namespace.NamespaceCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(pitrconfig.PitrConfigCmd)
	pitrconfig.PitrConfigCmd.PersistentFlags().StringVarP(&pitrconfig.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	pitrconfig.PitrConfigCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(connectionpooling.ConnectionPoolingCmd)
	connectionpooling.ConnectionPoolingCmd.PersistentFlags().StringVarP(&connectionpooling.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	connectionpooling.ConnectionPoolingCmd.MarkPersistentFlagRequired("cluster-name")

	util.AddCommandIfFeatureFlag(ClusterCmd, backupreplication.BackupReplicationCmd, util.BACKUP_REPLICATION_GCP_TARGET)
}
