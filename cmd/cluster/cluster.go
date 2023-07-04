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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yugabyte/ybm-cli/cmd/cluster/cert"
	encryption "github.com/yugabyte/ybm-cli/cmd/cluster/encryption"
	"github.com/yugabyte/ybm-cli/cmd/cluster/network"
	"github.com/yugabyte/ybm-cli/cmd/cluster/node"
	readreplica "github.com/yugabyte/ybm-cli/cmd/cluster/read-replica"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
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

	ClusterCmd.AddCommand(network.NetworkCmd)
	network.NetworkCmd.PersistentFlags().StringVarP(&network.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	network.NetworkCmd.MarkPersistentFlagRequired("cluster-name")
	network.NetworkCmd.RegisterFlagCompletionFunc("cluster-name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterListRequest := authApi.ListClusters()
		resp, r, err := clusterListRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// make a string array of cluster names
		var clusterNames []string
		for _, cluster := range resp.GetData() {
			clusterNames = append(clusterNames, cluster.Spec.Name)
		}
		return clusterNames, cobra.ShellCompDirectiveNoFileComp
	})

	ClusterCmd.AddCommand(readreplica.ReadReplicaCmd)
	readreplica.ReadReplicaCmd.PersistentFlags().StringVarP(&readreplica.ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	readreplica.ReadReplicaCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(node.NodeCmd)
	node.NodeCmd.PersistentFlags().StringP("cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	node.NodeCmd.MarkPersistentFlagRequired("cluster-name")

	ClusterCmd.AddCommand(encryption.EncryptionCmd)
	encryption.EncryptionCmd.PersistentFlags().StringP("cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	encryption.EncryptionCmd.MarkPersistentFlagRequired("cluster-name")
}
