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
)

var getClusterCmd = &cobra.Command{
	Use:   "get",
	Short: "Get clusters",
	Long:  "Get clusters",
	Run: func(cmd *cobra.Command, args []string) {
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName != "" {
			describeClusterCmd.Run(cmd, args)
			logrus.Warnln("The command `ybm cluster get --cluster-name` is deprecated. Please use `ybm cluster describe --cluster-name` instead.")
		} else {
			listClusterCmd.Run(cmd, args)
			logrus.Warnln("The command `ybm cluster get` is deprecated. Please use `ybm cluster list` instead.")
		}
	},
}

func init() {
	ClusterCmd.AddCommand(getClusterCmd)
	getClusterCmd.Flags().String("cluster-name", "", "[OPTIONAL] The name of the cluster to get details.")
}
