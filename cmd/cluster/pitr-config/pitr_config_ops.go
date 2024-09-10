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
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
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

var describePitrConfigCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe PITR Configs of a namespace in a cluster",
	Long:  "Describe PITR Configs of a namespace in a cluster in YugabyteDB Aeon",
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
		validateNamespaceNameType(namespaceName, namespaceType)
		pitrConfigId := requirePitrConfig(authApi, clusterID, namespaceName, namespaceType)

		resp, r, err := authApi.GetPitrConfig(clusterID, pitrConfigId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		pitrConfigCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewPitrConfigFormat(viper.GetString("output")),
		}

		formatter.SinglePitrConfigWrite(pitrConfigCtx, resp.GetData())

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
		validateNamespaceNameType(namespaceName, namespaceType)
		retentionPeriod, _ := cmd.Flags().GetInt32("retention-period-in-days")

		pitrConfigSpec, err := authApi.CreatePitrConfigSpec(namespaceName, namespaceType, retentionPeriod)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.CreatePitrConfig(clusterID).DatabasePitrConfigSpec(*pitrConfigSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		pitrConfigId := resp.Data.Info.Id

		msg := fmt.Sprintf("The PITR Configuration for %s namespace %s in cluster %s is being created\n\n", namespaceType, formatter.Colorize(namespaceName, formatter.GREEN_COLOR), formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			handleTaskCompletion(authApi, clusterID, msg, ybmclient.TASKTYPEENUM_ENABLE_DB_PITR)
			fmt.Printf("Successfully created PITR configuration.\n\n")

			getConfigResp, r, err := authApi.GetPitrConfig(clusterID, *pitrConfigId).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			pitrConfigData := getConfigResp.GetData()
			pitrConfigCtx := formatter.Context{
				Output: os.Stdout,
				Format: formatter.NewPitrConfigFormat(viper.GetString("output")),
			}

			formatter.SinglePitrConfigWrite(pitrConfigCtx, pitrConfigData)
		} else {
			fmt.Println(msg)
		}
	},
}

var restorePitrConfigCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore namespace via PITR Config for a cluster",
	Long:  "Restore namespace via PITR Config for a cluster in YugabyteDB Aeon",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		namespaceName, _ := cmd.Flags().GetString("namespace-name")
		namespaceType, _ := cmd.Flags().GetString("namespace-type")
		validateNamespaceNameType(namespaceName, namespaceType)
		restoreAtMilis, _ := cmd.Flags().GetInt64("restore-at-millis")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to restore the %s namespace: %s in cluster %s to the snapshot at %d", namespaceType, namespaceName, ClusterName, restoreAtMilis), viper.GetBool("force"))
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
		namespaceType, _ := cmd.Flags().GetString("namespace-type")
		validateNamespaceNameType(namespaceName, namespaceType)
		restoreAtMilis, _ := cmd.Flags().GetInt64("restore-at-millis")
		pitrConfigId := requirePitrConfig(authApi, clusterID, namespaceName, namespaceType)

		restoreViaPitrConfigSpec, err := authApi.CreateRestoreViaPitrConfigSpec(restoreAtMilis)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		_, r, err := authApi.RestoreViaPitrConfig(clusterID, pitrConfigId).DatabaseRestoreViaPitrSpec(*restoreViaPitrConfigSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The %s namespace %s in cluster %s is being restored via PITR Configuration.\n\n", namespaceType, formatter.Colorize(namespaceName, formatter.GREEN_COLOR), formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			handleTaskCompletion(authApi, clusterID, msg, ybmclient.TASKTYPEENUM_RESTORE_DB_PITR)
			fmt.Printf("\nSuccessfully restored %s namespace %s in cluster %s to the snapshot at %d ms.\n\n", namespaceType, namespaceName, ClusterName, restoreAtMilis)
		} else {
			fmt.Println(msg)
		}
	},
}

var deletePitrConfigCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete PITR Config for a cluster",
	Long:  "Delete PITR Config for a cluster in YugabyteDB Aeon",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		namespaceName, _ := cmd.Flags().GetString("namespace-name")
		namespaceType, _ := cmd.Flags().GetString("namespace-type")
		validateNamespaceNameType(namespaceName, namespaceType)
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to delete PITR Configuration for the %s namespace: %s in cluster: %s", namespaceType, namespaceName, ClusterName), viper.GetBool("force"))
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
		namespaceType, _ := cmd.Flags().GetString("namespace-type")
		validateNamespaceNameType(namespaceName, namespaceType)
		pitrConfigId := requirePitrConfig(authApi, clusterID, namespaceName, namespaceType)

		r, err := authApi.DeletePitrConfig(clusterID, pitrConfigId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The PITR Configuration for %s namespace %s in cluster %s is being removed.\n\n", namespaceType, formatter.Colorize(namespaceName, formatter.GREEN_COLOR), formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			handleTaskCompletion(authApi, clusterID, msg, ybmclient.TASKTYPEENUM_DISABLE_DB_PITR)
			fmt.Printf("\nSuccessfully removed PITR Configuration for %s namespace %s in cluster %s.\n\n", namespaceType, namespaceName, ClusterName)
		} else {
			fmt.Println(msg)
		}
	},
}

func validateNamespaceNameType(namespaceName string, namespaceType string) {
	if len(namespaceName) == 0 {
		logrus.Fatalln("Namespace name must be provided.")
	}
	if !(namespaceType == "YCQL" || namespaceType == "YSQL") {
		logrus.Fatalln("Only YCQL or YSQL namespace types are allowed.")
	}
}

func requirePitrConfig(authApi *ybmAuthClient.AuthApiClient, clusterID string, namespaceName string, namespaceType string) string {
	var pitrConfigId string
	listConfigsResp, listConfigsResponse, listConfigsError := authApi.ListClusterPitrConfigs(clusterID).Execute()
	if listConfigsError != nil {
		logrus.Debugf("Full HTTP response: %v", listConfigsResponse)
		logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(listConfigsError))
	}

	for _, pitrConfig := range listConfigsResp.GetData() {
		if pitrConfig.Spec.DatabaseName == namespaceName && pitrConfig.Spec.DatabaseType == ybmclient.YbApiEnum(namespaceType) {
			pitrConfigId = *pitrConfig.Info.Id
			break
		}
	}

	if len(pitrConfigId) == 0 {
		logrus.Fatalf("No PITR Configs found for %s namespace %s in cluster %s.\n", namespaceType, namespaceName, ClusterName)
	}
	return pitrConfigId
}

func handleTaskCompletion(authApi *ybmAuthClient.AuthApiClient, clusterID string, msg string, taskType ybmclient.TaskTypeEnum) {
	returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, taskType, []string{"FAILED", "SUCCEEDED"}, msg)
	if err != nil {
		logrus.Fatalf("error when getting task status: %s", err)
	}
	if returnStatus != "SUCCEEDED" {
		logrus.Fatalf("Operation failed with error: %s", returnStatus)
	}
}

func init() {
	util.AddCommandIfFeatureFlag(PitrConfigCmd, listPitrConfigCmd, util.PITR_CONFIG)

	util.AddCommandIfFeatureFlag(PitrConfigCmd, describePitrConfigCmd, util.PITR_CONFIG)
	describePitrConfigCmd.Flags().SortFlags = false
	describePitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace to be restored via PITR Config.")
	describePitrConfigCmd.MarkFlagRequired("namespace-name")
	describePitrConfigCmd.Flags().String("namespace-type", "", "[REQUIRED] The type of the namespace. Available options are YCQL and YSQL")
	describePitrConfigCmd.MarkFlagRequired("namespace-type")

	util.AddCommandIfFeatureFlag(PitrConfigCmd, createPitrConfigCmd, util.PITR_CONFIG)
	createPitrConfigCmd.Flags().SortFlags = false
	createPitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace for which the PITR Config is to be created.")
	createPitrConfigCmd.MarkFlagRequired("namespace-name")
	createPitrConfigCmd.Flags().String("namespace-type", "", "[REQUIRED] The type of the namespace. Available options are YCQL and YSQL")
	createPitrConfigCmd.MarkFlagRequired("namespace-type")
	createPitrConfigCmd.Flags().Int32("retention-period-in-days", 2, "[REQUIRED] The time duration in days to retain a snapshot for.")
	createPitrConfigCmd.MarkFlagRequired("retention-period-in-days")

	util.AddCommandIfFeatureFlag(PitrConfigCmd, restorePitrConfigCmd, util.PITR_CONFIG)
	restorePitrConfigCmd.Flags().SortFlags = false
	restorePitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace to be restored via PITR Config.")
	restorePitrConfigCmd.MarkFlagRequired("namespace-name")
	restorePitrConfigCmd.Flags().String("namespace-type", "", "[REQUIRED] The type of the namespace. Available options are YCQL and YSQL")
	restorePitrConfigCmd.MarkFlagRequired("namespace-type")
	restorePitrConfigCmd.Flags().Int64("restore-at-millis", 1, "[REQUIRED] The time in milliseconds to which the namespace is to be restored")
	restorePitrConfigCmd.MarkFlagRequired("restore-at-millis")
	restorePitrConfigCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	util.AddCommandIfFeatureFlag(PitrConfigCmd, deletePitrConfigCmd, util.PITR_CONFIG)
	deletePitrConfigCmd.Flags().SortFlags = false
	deletePitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace to be restored via PITR Config.")
	deletePitrConfigCmd.MarkFlagRequired("namespace-name")
	deletePitrConfigCmd.Flags().String("namespace-type", "", "[REQUIRED] The type of the namespace. Available options are YCQL and YSQL")
	deletePitrConfigCmd.MarkFlagRequired("namespace-type")
	deletePitrConfigCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

}
