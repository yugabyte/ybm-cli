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
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
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
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
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
		formatter.CMKWrite(cmkCtx, *resp.GetData().Spec.Get())
	},
}

var updateCmkState = &cobra.Command{
	Use:   "update-state",
	Short: "Update Encryption at Rest (EaR) state for a cluster",
	Long:  "Update Encryption at Rest (EaR) state for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
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
		enableFlag, _ := cmd.Flags().GetBool("enable")
		disableFlag, _ := cmd.Flags().GetBool("disable")

		if enableFlag == disableFlag {
			logrus.Fatalf("Please enter valid input. Specify either enable or disable flag.")
		}

		if resp.Data == nil {
			logrus.Fatalf("No Encryption at rest configuration found for this cluster")
		}

		cmkId := resp.Data.Info.GetCmkId()

		cmkStatus := true
		if disableFlag {
			cmkStatus = false
		}

		updateCMKStateSpec := ybmclient.NewUpdateCMKStateSpec(cmkStatus)
		resp, r, err = authApi.UpdateClusterCmkState(clusterId, cmkId).UpdateCMKStateSpec(*updateCMKStateSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		cmkStatusDisplay := "DISABLED"
		if cmkStatus {
			cmkStatusDisplay = "ENABLED"
		}

		fmt.Printf("Successfully %s encryption spec status for cluster %s\n", formatter.Colorize(cmkStatusDisplay, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))
	},
}

var updateCmk = &cobra.Command{
	// This API creates new EAR configuration if not found, else updates the current one
	Use:   "update",
	Short: "Update Encryption at Rest (EaR) configurations for a cluster",
	Long:  "Update Encryption at Rest (EaR) configurations for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")

		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		cmkSpec, err := GetCmkSpecFromCommand(cmd)
		if err != nil {
			logrus.Fatalf("Unable to parse new CMK spec: %v", err)
		}

		_, res, err := authApi.EditClusterCMKs(clusterId).CMKSpec(*cmkSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", res)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		fmt.Printf("Successfully updated encryption spec for cluster %s\n", clusterName)
	},
}

func init() {
	EncryptionCmd.AddCommand(listCmk)
	EncryptionCmd.AddCommand(updateCmk)
	util.AddCommandIfFeatureFlag(EncryptionCmd, updateCmkState, util.CLUSTER_CMK_UPDATE)
	updateCmk.Flags().String("encryption-spec", "", `[REQUIRED] The customer managed key spec for the cluster.
	Please provide key value pairs as follows:
	For AWS: 
	cloud-provider=AWS,aws-secret-key=<secret-key>,aws-access-key=<access-key>,aws-arn=<arn1>,aws-arn=<arn2> .
	aws-access-key can be ommitted if the environment variable YBM_AWS_SECRET_KEY is set. If the environment variable is not set, the user will be prompted to enter the value.
	For GCP:
	cloud-provider=GCP,gcp-resource-id=<resource-id>,gcp-service-account-path=<service-account-path>.`)
	updateCmk.MarkFlagRequired("encryption-spec")
	updateCmkState.Flags().Bool("enable", false, "Enable EAR")
	updateCmkState.Flags().Bool("disable", false, "Disable EAR")
}
