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
	"github.com/spf13/cobra"
)

var ClusterName string
var showAll bool

var GcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Manage replication of cluster backups to GCP buckets",
	Long:  "Manage replication of cluster backups to GCP buckets",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	GcpCmd.PersistentFlags().StringVarP(&ClusterName, "cluster-name", "c", "", "[REQUIRED] The name of the cluster.")
	GcpCmd.MarkPersistentFlagRequired("cluster-name")

	GcpCmd.AddCommand(describeGcpCmd)
	describeGcpCmd.Flags().BoolVar(&showAll, "show-all", false, "Show all configurations including scheduled for expiry and removed regions")

	GcpCmd.AddCommand(enableGcpCmd)
	GcpCmd.AddCommand(updateGcpCmd)
	GcpCmd.AddCommand(disableGcpCmd)
}
