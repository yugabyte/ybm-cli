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
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var ClusterName string
var allPitrConfigSpecs []string

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

		pitrConfigSpecs, err := ParsePitrConfigSpecs(authApi, clusterID, allPitrConfigSpecs)
		if err != nil {
			logrus.Fatalf("Error while parsing PITR Config specs: %s", ybmAuthClient.GetApiErrorDetails(err))
			return
		}

		bulkPitrConfigSpec, err := authApi.CreateBulkPitrConfigSpec(pitrConfigSpecs)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.CreatePitrConfig(clusterID).BulkCreateDatabasePitrConfigSpec(*bulkPitrConfigSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		pitrConfigsData := resp.GetData()

		msg := fmt.Sprintf("The requested PITR Configurations are being created\n\n")

		if viper.GetBool("wait") {
			handleTaskCompletion(authApi, clusterID, msg, ybmclient.TASKTYPEENUM_BULK_ENABLE_DB_PITR)
			fmt.Printf("Successfully created PITR configurations.\n\n")
			createdConfigsData := []ybmclient.DatabasePitrConfigData{}
			for _, configData := range pitrConfigsData {
				configId := configData.Info.Id
				getConfigResp, r, err := authApi.GetPitrConfig(clusterID, *configId).Execute()
				if err != nil {
					logrus.Debugf("Full HTTP response: %v", r)
					logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
				}
				createdConfigsData = append(createdConfigsData, getConfigResp.GetData())
			}

			pitrConfigCtx := formatter.Context{
				Output: os.Stdout,
				Format: formatter.NewPitrConfigFormat(viper.GetString("output")),
			}

			formatter.PitrConfigWrite(pitrConfigCtx, createdConfigsData)
		} else {
			fmt.Println(msg)
		}
	},
}

// Parse array of PITR config spec string to params
func ParsePitrConfigSpecs(authApi *ybmAuthClient.AuthApiClient, clusterID string, configSpecs []string) ([]ybmclient.DatabasePitrConfigSpec, error) {
	var err error
	pitrConfigSpecs := []ybmclient.DatabasePitrConfigSpec{}
	namespacesResp, r, err := authApi.GetClusterNamespaces(clusterID).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
	}

	for _, configSpec := range configSpecs {
		var namespaceNameProvided bool
		var namespaceTypeProvided bool
		var retentionPeriodProvided bool
		var namespaceName string
		var namespaceType string
		spec := *ybmclient.NewDatabasePitrConfigSpecWithDefaults()
		configSpec := strings.ReplaceAll(configSpec, " ", "")

		for _, subSpec := range strings.Split(configSpec, ",") {
			if !strings.Contains(subSpec, "=") {
				return nil, fmt.Errorf("namespace-name, namespace-type and retention-period-in-days must be provided as key value pairs for each PITR Config to be created")
			}
			kvp := strings.Split(subSpec, "=")
			key := kvp[0]
			val := kvp[1]
			n := 0
			err = nil
			switch key {
			case "namespace-name":
				if len(val) == 0 {
					return nil, fmt.Errorf("Namespace name must be provided.")
				}
				namespaceName = val
				namespaceNameProvided = true
			case "namespace-type":
				if !(val == "YCQL" || val == "YSQL") {
					return nil, fmt.Errorf("Only YCQL or YSQL namespace types are allowed.")
				}
				namespaceType = val
				namespaceTypeProvided = true
			case "retention-period-in-days":
				n, err = strconv.Atoi(val)
				if err != nil {
					return nil, err
				}
				if n > 1 && n <= math.MaxInt32 {
					retentionPeriod := int32(n)
					spec.SetRetentionPeriod(retentionPeriod)
					retentionPeriodProvided = true
				} else {
					return nil, fmt.Errorf("Minimum retention period is 2 days")
				}
			}
		}
		if !(namespaceNameProvided && namespaceTypeProvided && retentionPeriodProvided) {
			return nil, fmt.Errorf("namespace-name, namespace-type and retention-period-in-days must be provided for each PITR Config to be created")
		}
		namespaceId := requireNamespace(namespacesResp, namespaceName, namespaceType)
		spec.SetDatabaseId(namespaceId)
		pitrConfigSpecs = append(pitrConfigSpecs, spec)
	}

	if len(pitrConfigSpecs) == 0 {
		return nil, fmt.Errorf("namespace-name, namespace-type and retention-period-in-days must be provided for each PITR Config to be created")
	}

	return pitrConfigSpecs, nil
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

var updatePitrConfigCmd = &cobra.Command{
	Use:   "update",
	Short: "Update PITR Config for a cluster",
	Long:  "Update PITR Config for a cluster in YugabyteDB Aeon",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		namespaceName, _ := cmd.Flags().GetString("namespace-name")
		namespaceType, _ := cmd.Flags().GetString("namespace-type")
		retentionPeriod, _ := cmd.Flags().GetInt32("retention-period-in-days")
		validateNamespaceNameType(namespaceName, namespaceType)
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to update PITR Configuration for the %s namespace: %s in cluster: %s to have a retention period of %d days", namespaceType, namespaceName, ClusterName, retentionPeriod), viper.GetBool("force"))
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
		retentionPeriod, _ := cmd.Flags().GetInt32("retention-period-in-days")
		validateNamespaceNameType(namespaceName, namespaceType)
		if retentionPeriod < 2 {
			logrus.Fatalln("Minimum retention period is 2 days")
		}
		pitrConfigId := requirePitrConfig(authApi, clusterID, namespaceName, namespaceType)

		updatePitrConfigSpec := ybmclient.NewUpdateDatabasePitrConfigSpec(retentionPeriod)

		_, r, err := authApi.UpdatePitrConfig(clusterID, pitrConfigId).UpdateDatabasePitrConfigSpec(*updatePitrConfigSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The PITR Configuration for %s namespace %s in cluster %s is being updated.\n\n", namespaceType, formatter.Colorize(namespaceName, formatter.GREEN_COLOR), formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			handleTaskCompletion(authApi, clusterID, msg, ybmclient.TASKTYPEENUM_UPDATE_DB_PITR)
			fmt.Printf("\nSuccessfully updated PITR Configuration for %s namespace %s in cluster %s.\n\n", namespaceType, namespaceName, ClusterName)
		} else {
			fmt.Println(msg)
		}
	},
}

var clonePitrConfigCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone namespace via PITR Config for a cluster",
	Long:  "Clone namespace via PITR Config for a cluster in YugabyteDB Aeon",
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
		cloneAs, _ := cmd.Flags().GetString("clone-as")
		cloneSpec := ybmclient.NewDatabaseCloneSpec()

		namespacesResp, r, err := authApi.GetClusterNamespaces(clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		namespaceId := requireNamespace(namespacesResp, namespaceName, namespaceType)
		pitrConfigId := checkPitrConfigExists(authApi, clusterID, namespaceId)
		if len(pitrConfigId) == 0 {
			// No PITR config exists, so we create one and clone to current time.
			// "clone-at-millis" argument should not be provided in this case.
			if cmd.Flags().Lookup("clone-at-millis").Changed {
				logrus.Fatalf("No PITR configuration found for %s namespace %s in cluster %s. The 'clone-at-millis' parameter cannot be used unless a valid PITR configuration is set up.\n", namespaceType, namespaceName, ClusterName)
			}
			cloneSpec.SetCloneNow(*ybmclient.NewDatabaseCloneNowSpec(namespaceId, cloneAs))
		} else {
			if cmd.Flags().Lookup("clone-at-millis").Changed {
				cloneAtMillis, _ := cmd.Flags().GetInt64("clone-at-millis")
				cloneSpec.SetClonePointInTime(*ybmclient.NewDatabaseClonePITSpec(cloneAtMillis, pitrConfigId, cloneAs))
			} else {
				cloneSpec.SetCloneNow(*ybmclient.NewDatabaseCloneNowSpec(namespaceId, cloneAs))
			}
		}

		_, cloneResp, cloneErr := authApi.CloneViaPitrConfig(clusterID).DatabaseCloneSpec(*cloneSpec).Execute()
		if cloneErr != nil {
			logrus.Debugf("Full HTTP response: %v", cloneResp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(cloneErr))
		}

		msg := fmt.Sprintf("The %s namespace %s in cluster %s is being cloned via PITR Configuration.\n\n", namespaceType, formatter.Colorize(namespaceName, formatter.GREEN_COLOR), formatter.Colorize(ClusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			handleTaskCompletion(authApi, clusterID, msg, ybmclient.TASKTYPEENUM_CLONE_DB_PITR)
			fmt.Printf("\nSuccessfully cloned %s namespace %s in cluster %s.\n\n", namespaceType, namespaceName, ClusterName)
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

func checkPitrConfigExists(authApi *ybmAuthClient.AuthApiClient, clusterID string, namespaceId string) string {
	var pitrConfigId string
	listConfigsResp, listConfigsResponse, listConfigsError := authApi.ListClusterPitrConfigs(clusterID).Execute()
	if listConfigsError != nil {
		logrus.Debugf("Full HTTP response: %v", listConfigsResponse)
		logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(listConfigsError))
	}

	for _, pitrConfig := range listConfigsResp.GetData() {
		if pitrConfig.Spec.DatabaseId == namespaceId {
			pitrConfigId = *pitrConfig.Info.Id
			break
		}
	}
	return pitrConfigId
}

func requirePitrConfig(authApi *ybmAuthClient.AuthApiClient, clusterID string, namespaceName string, namespaceType string) string {
	namespacesResp, r, err := authApi.GetClusterNamespaces(clusterID).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
	}
	namespaceId := requireNamespace(namespacesResp, namespaceName, namespaceType)
	pitrConfigId := checkPitrConfigExists(authApi, clusterID, namespaceId)
	if len(pitrConfigId) == 0 {
		logrus.Fatalf("No PITR Configs found for %s namespace %s in cluster %s.\n", namespaceType, namespaceName, ClusterName)
	}
	return pitrConfigId
}

func requireNamespace(namespacesResp ybmclient.ClusterNamespacesListResponse, namespaceName string, namespaceType string) string {
	var namespaceId string

	for _, namespace := range namespacesResp.Data {
		if namespace.GetName() == namespaceName && namespace.GetTableType() == GetNamespaceTypeMap()[namespaceType] {
			namespaceId = namespace.GetId()
		}
	}
	if len(namespaceId) == 0 {
		logrus.Fatalf("No %s namespace found with name %s in cluster %s.\n", namespaceType, namespaceName, ClusterName)
	}
	return namespaceId
}

func GetNamespaceTypeMap() map[string]string {
	return map[string]string{
		"YSQL": "PGSQL_TABLE_TYPE",
		"YCQL": "YQL_TABLE_TYPE",
	}
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
	createPitrConfigCmd.Flags().StringArrayVarP(&allPitrConfigSpecs, "pitr-config", "p", []string{}, `[REQUIRED] Information for the PITR Configs to be created. All values are mandatory. Available options for namespace type are YCQL and YSQL. Minimum retention period is 2 days. Please provide key value pairs namespace-name=<namespace-name>,namespace-type=<namespace-type>,retention-period-in-days=<retention-period-in-days>.`)

	util.AddCommandIfFeatureFlag(PitrConfigCmd, restorePitrConfigCmd, util.PITR_RESTORE)
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

	util.AddCommandIfFeatureFlag(PitrConfigCmd, updatePitrConfigCmd, util.PITR_CONFIG)
	updatePitrConfigCmd.Flags().SortFlags = false
	updatePitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace to be restored via PITR Config.")
	updatePitrConfigCmd.MarkFlagRequired("namespace-name")
	updatePitrConfigCmd.Flags().String("namespace-type", "", "[REQUIRED] The type of the namespace. Available options are YCQL and YSQL")
	updatePitrConfigCmd.MarkFlagRequired("namespace-type")
	updatePitrConfigCmd.Flags().Int32("retention-period-in-days", 2, "[REQUIRED] The retention period of the snapshots from the PITR Config")
	updatePitrConfigCmd.MarkFlagRequired("retention-period-in-days")
	updatePitrConfigCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	util.AddCommandIfFeatureFlag(PitrConfigCmd, clonePitrConfigCmd, util.PITR_CLONE)
	clonePitrConfigCmd.Flags().SortFlags = false
	clonePitrConfigCmd.Flags().String("namespace-name", "", "[REQUIRED] Namespace to be cloned via PITR Config.")
	clonePitrConfigCmd.MarkFlagRequired("namespace-name")
	clonePitrConfigCmd.Flags().String("namespace-type", "", "[REQUIRED] The type of the namespace. Available options are YCQL and YSQL")
	clonePitrConfigCmd.MarkFlagRequired("namespace-type")
	clonePitrConfigCmd.Flags().String("clone-as", "", "[REQUIRED] The name of the cloned namespace.")
	clonePitrConfigCmd.MarkFlagRequired("clone-as")
	clonePitrConfigCmd.Flags().Int64("clone-at-millis", 0, "[OPTIONAL] The time in milliseconds to which the namespace is to be cloned. If not provided, the current state is cloned.")
}
