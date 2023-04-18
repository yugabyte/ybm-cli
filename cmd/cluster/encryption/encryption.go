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
package ear

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var EncryptionCmd = &cobra.Command{
	Use:   "encryption",
	Short: "Manage Encryption at Rest (EaR) for a cluster",
	Long:  "Manage Encryption at Rest (EaR) for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listCmk = &cobra.Command{
	Use:   "list",
	Short: "List Encryption at Rest (EaR) configurations for a cluster",
	Long:  "List Encryption at Rest (EaR) configurations for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.ListClusterCMKs(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if resp.Data == nil {
			logrus.Fatalf("No Encryption at rest configuration found for this cluster")
		}

		cmkCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewCMKFormat(viper.GetString("output")),
		}
		formatter.CMKWrite(cmkCtx, resp.GetData())
	},
}

var updateCmk = &cobra.Command{
	Use:   "update",
	Short: "Update Encryption at Rest (EaR) configurations for a cluster",
	Long:  "Update Encryption at Rest (EaR) configurations for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.ListClusterCMKs(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if resp.Data == nil {
			logrus.Fatalf("No Encryption at rest configuration found for this cluster")
		}

		oldCmkSpec := resp.GetData()

		newCmkSpec, err := GetCmkSpecFromCommand(cmd)
		if err != nil {
			logrus.Fatalf("Unable to parse new CMK spec: %v", err)
		}

		// Need to copy over the AWS ARNs
		newCmkSpec.AwsCmkSpec.ArnList = oldCmkSpec.AwsCmkSpec.ArnList

		resp, r, err = authApi.EditClusterCMKs(clusterId).CMKSpec(*newCmkSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Successfully updated encryption spec for cluster %s\n", clusterName)
	},
}

func init() {
	EncryptionCmd.AddCommand(listCmk)
	EncryptionCmd.AddCommand(updateCmk)
	updateCmk.Flags().String("encryption-spec", "", "[REQUIRED] The customer managed key spec for the cluster. Please provide key value pairs provider=AWS,aws-secret-key=<secret-key>,aws-access-key=<access-key>. If specified, all parameters for that provider are mandatory.")
}
