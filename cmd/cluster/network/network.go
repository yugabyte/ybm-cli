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

package network

import (
	"github.com/spf13/cobra"
	"github.com/yugabyte/ybm-cli/cmd/cluster/network/endpoint"
	"github.com/yugabyte/ybm-cli/cmd/cluster/network/nal"
)

var ClusterName string

var NetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage network operations",
	Long:  "Manage network operations for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	NetworkCmd.AddCommand()
	NetworkCmd.AddCommand(nal.AllowListCmd)
	NetworkCmd.AddCommand(endpoint.EndpointCmd)
}
