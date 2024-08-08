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

package pitrconfig

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var ClusterName string

var listPitrConfigCmd = &cobra.Command{
	Use:   "list",
	Short: "List PITR Configs for a cluster",
	Long:  "List PITR Configs for a cluster in YugabyteDB Aeon",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatal(err)
		}
		resp, r, err := authApi.ListClusterPitrConfigs(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			logrus.Info("No PITR Configs found for cluster.\n")
			return
		}

		pitrConfigCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewPitrConfigFormat(viper.GetString("output")),
		}

		formatter.PitrConfigWrite(pitrConfigCtx, resp.GetData())

	},
}

var createPitrConfigCmd = &cobra.Command{
	Use:   "create",
	Short: "Create PITR Config for a cluster",
	Long:  "Create PITR Config for a cluster in YugabyteDB Aeon",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterID, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		namespaceName, _ := cmd.Flags().GetString("namespace-name")
		namespaceType, _ := cmd.Flags().GetString("namespace-type")
		retentionPeriod, _ := cmd.Flags().GetInt32("retention-period-in-days")

		if !(namespaceType == "YCQL" || namespaceType == "YSQL") {
			logrus.Fatalln("Only YCQL or YSQL namespace types are allowed.")
		}

		pitrConfigSpec, err := authApi.CreatePitrConfigSpec(namespaceName, namespaceType, retentionPeriod)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.CreatePitrConfig(clusterID).DatabasePitrConfigSpec(*pitrConfigSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("\nSuccessfully created PITR configuration.\n\n")

		pitrConfigCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewPitrConfigFormat(viper.GetString("output")),
		}

		formatter.SinglePitrConfigWrite(pitrConfigCtx, resp.GetData())
	},
}

var restorePitrConfigCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore namespace via PITR Config for a cluster",
	Long:  "Restore namespace via PITR Config for a cluster in YugabyteDB Aeon",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		namespaceName, _ := cmd.Flags().GetString("namespace-name")
		restoreAtMilis, _ := cmd.Flags().GetInt64("restore-at-millis")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to restore the namespace: %s at %d", namespaceName, restoreAtMilis), viper.GetBool("force"))
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
		clusterID, err := authApi.GetClusterIdByName(ClusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		namespaceName, _ := cmd.Flags().GetString("namespace-name")
		restoreAtMilis, _ := cmd.Flags().GetInt64("restore-at-millis")

		var pitrConfigId string
		listConfigsResp, listConfigsResponse, listConfigsError := authApi.ListClusterPitrConfigs(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", listConfigsResponse)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(listConfigsError))
		}

		for _, pitrConfig := range listConfigsResp.GetData() {
			if pitrConfig.Spec.DatabaseName == namespaceName {
				pitrConfigId = *pitrConfig.Info.Id
				break
			}
		}

		if len(pitrConfigId) == 0 {
			logrus.Fatalf("No PITR Configs found for namespace %s\n", namespaceName)
		}

		restoreViaPitrConfigSpec, err := authApi.CreateRestoreViaPitrConfigSpec(restoreAtMilis)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		_, r, err := authApi.RestoreViaPitrConfig(clusterID, pitrConfigId).DatabaseRestoreViaPitrSpec(*restoreViaPitrConfigSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("\nSuccessfully restored namespace %s at %d ms.\n\n", namespaceName, restoreAtMilis)
	},
}

func init() {
	util.AddCommandIfFeatureFlag(PitrConfigCmd, listPitrConfigCmd, util.PITR_CONFIG)

	util.AddCommandIfFeatureFlag(PitrConfigCmd, createPitrConfigCmd, util.PITR_CONFIG)
	createPitrConfigCmd.Flags().SortFlags = false
	createPitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace for which the PITR Config is to be created.")
	createPitrConfigCmd.MarkFlagRequired("namespace-name")
	createPitrConfigCmd.Flags().String("namespace-type", "", "[REQUIRED] The type of the namespace. Available options are YCQL and YSQL")
	createPitrConfigCmd.MarkFlagRequired("namespace-type")
	createPitrConfigCmd.Flags().Int32("retention-period-in-days", 1, "[REQUIRED] The time duration in days to retain a snapshot for.")
	createPitrConfigCmd.MarkFlagRequired("retention-period-in-days")

	util.AddCommandIfFeatureFlag(PitrConfigCmd, restorePitrConfigCmd, util.PITR_CONFIG)
	restorePitrConfigCmd.Flags().SortFlags = false
	restorePitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace to be restored via PITR Config.")
	restorePitrConfigCmd.MarkFlagRequired("namespace-name")
	restorePitrConfigCmd.Flags().Int64("restore-at-millis", 1, "[REQUIRED] The time in milliseconds to which the namespace is to be restored")
	restorePitrConfigCmd.MarkFlagRequired("restore-at-millis")
	restorePitrConfigCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

}
