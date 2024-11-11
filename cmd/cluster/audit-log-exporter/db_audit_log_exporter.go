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

package audit_log_exporter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var ClusterName string

var DbAuditLoggingCmd = &cobra.Command{
	Use:   "db-audit-logging",
	Short: "Configure Database Audit Logging for your Cluster.",
	Long:  "Configure Database Audit Logging for your Cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var enableDbAuditLoggingCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable Database Audit Logging",
	Long:  "Enable Database Audit Logging",
	Run: func(cmd *cobra.Command, args []string) {

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		integrationName, _ := cmd.Flags().GetString("integration-name")

		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		integrationId, err := authApi.GetIntegrationIdFromName(integrationName)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		ysqlConfig, _ := cmd.Flags().GetStringToString("ysql-config")
		statement_classes, _ := cmd.Flags().GetString("statement_classes")

		dbAuditLogsExporterSpec, err := setDbAuditLogsExporterSpec(ysqlConfig, statement_classes, integrationId)

		if err != nil {
			logrus.Fatalf(err.Error())
		}

		resp, r, err := authApi.AssignDbAuditLogsExporterConfig(clusterId).DbAuditExporterConfigSpec(*dbAuditLogsExporterSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		respData := resp.GetData()

		msg := fmt.Sprintf("Db audit logging is being enabled for cluster %s", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_ENABLE_DATABASE_AUDIT_LOGGING, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("DB audit logging has been enabled on the cluster %v\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.ListDbAuditExporterConfig(clusterId).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			respData = respC.GetData()[0]
		} else {
			fmt.Println(msg)
		}

		dbAuditLogsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbAuditLogsExporterFormat(viper.GetString("output")),
		}

		formatter.DbAuditLogsExporterWrite(dbAuditLogsExporterCtx, []openapi.DbAuditExporterConfigurationData{respData})
	},
}

var updateDbAuditLoggingCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Database Audit Logging Configuration",
	Long:  "Update Database Audit Logging Configuration",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		integrationName, _ := cmd.Flags().GetString("integration-name")
		ysqlConfig, _ := cmd.Flags().GetStringToString("ysql-config")
		statement_classes, _ := cmd.Flags().GetString("statement_classes")

		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		integrationId, err := getIntegrationIdFromName(integrationName, authApi)

		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		dbAuditLogsExporterSpec, err := setDbAuditLogsExporterSpec(ysqlConfig, statement_classes, integrationId)
		if err != nil {
			logrus.Fatalf(err.Error())
		}

		exportConfigId := getDbAuditExportConfigIdForCluster(authApi, clusterId)

		resp, r, err := authApi.UpdateDbAuditExporterConfig(clusterId, exportConfigId).DbAuditExporterConfigSpec(*dbAuditLogsExporterSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		respData := resp.GetData()

		msg := fmt.Sprintf("DB audit logging configuration is being updated for cluster %s", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_EDIT_DATABASE_AUDIT_LOGGING, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("DB audit logging configuration has been updated for the cluster %v\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.ListDbAuditExporterConfig(clusterId).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			respData = respC.GetData()[0]
		} else {
			fmt.Println(msg)
		}

		dbAuditLogsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbAuditLogsExporterFormat(viper.GetString("output")),
		}

		formatter.DbAuditLogsExporterWrite(dbAuditLogsExporterCtx, []openapi.DbAuditExporterConfigurationData{respData})
	},
}

var describeDbAuditLoggingCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe Database Audit Logging configuration",
	Long:  "Describe Database Audit Logging configuration",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")

		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		resp, r, err := authApi.ListDbAuditExporterConfig(clusterId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		dbAuditLogsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbAuditLogsExporterFormat(viper.GetString("output")),
		}

		if len(resp.GetData()) < 1 {
			fmt.Println("No DB Audit Logs Exporter found")
			return
		}

		formatter.DbAuditLogsExporterWrite(dbAuditLogsExporterCtx, resp.GetData())
	},
}

var disableDbAuditLoggingCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable Database Audit Logging",
	Long:  "Disable Database Audit Logging, if enabled",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to disable DB audit logging for cluster: %s", clusterName), viper.GetBool("force"))
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

		clusterName, _ := cmd.Flags().GetString("cluster-name")

		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		exportConfigId := getDbAuditExportConfigIdForCluster(authApi, clusterId)

		resp, _, err := authApi.UnassignDbAuditLogsExportConfig(clusterId, exportConfigId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("DB Audit Logging is being disabled for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterId, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_DISABLE_DATABASE_AUDIT_LOGGING, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("DB audit logging has been disabled for the cluster %v\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
			return
		} else {
			fmt.Println(msg)
		}
	},
}

func init() {
	DbAuditLoggingCmd.AddCommand(enableDbAuditLoggingCmd)
	enableDbAuditLoggingCmd.Flags().SortFlags = false
	enableDbAuditLoggingCmd.Flags().String("integration-name", "", "[REQUIRED] Name of the Integration")
	enableDbAuditLoggingCmd.MarkFlagRequired("integration-name")
	enableDbAuditLoggingCmd.Flags().StringToString("ysql-config", nil, `[REQUIRED] The ysql config to setup DB auditting
	Please provide key value pairs as follows:
	log_catalog=<boolean>,log_level=<LOG_LEVEL>,log_client=<boolean>,log_parameter=<boolean>,
	log_relation=<boolean>,log_statement_once=<boolean>`)
	enableDbAuditLoggingCmd.MarkFlagRequired("ysql-config")
	enableDbAuditLoggingCmd.Flags().String("statement_classes", "", `[REQUIRED] The ysql config statement classes
	Please provide key value pairs as follows:
	statement_classes=READ,WRITE,MISC`)
	enableDbAuditLoggingCmd.MarkFlagRequired("statement_classes")
	enableDbAuditLoggingCmd.Flags().String("cluster-name", "", "[REQUIRED] The cluster name to assign DB auditting")
	enableDbAuditLoggingCmd.MarkFlagRequired("cluster-name")

	DbAuditLoggingCmd.AddCommand(describeDbAuditLoggingCmd)
	describeDbAuditLoggingCmd.Flags().SortFlags = false
	describeDbAuditLoggingCmd.Flags().String("cluster-name", "", "[REQUIRED] The cluster name to list DB audit export config")
	describeDbAuditLoggingCmd.MarkFlagRequired("cluster-name")

	DbAuditLoggingCmd.AddCommand(updateDbAuditLoggingCmd)
	updateDbAuditLoggingCmd.Flags().SortFlags = false
	updateDbAuditLoggingCmd.Flags().String("integration-name", "", "[REQUIRED] Name of the Integration")
	updateDbAuditLoggingCmd.MarkFlagRequired("integration-name")
	updateDbAuditLoggingCmd.Flags().StringToString("ysql-config", nil, `The ysql config to setup DB auditting
	Please provide key value pairs as follows:
	log_catalog=<boolean>,log_level=<LOG_LEVEL>,log_client=<boolean>,log_parameter=<boolean>,
	log_relation=<boolean>,log_statement_once=<boolean>`)
	updateDbAuditLoggingCmd.Flags().String("statement_classes", "", `The ysql config statement classes
	Please provide key value pairs as follows:
	statement_classes=READ,WRITE,MISC`)
	updateDbAuditLoggingCmd.Flags().String("cluster-name", "", "[REQUIRED] The cluster name to assign DB auditting")
	updateDbAuditLoggingCmd.MarkFlagRequired("cluster-name")

	DbAuditLoggingCmd.AddCommand(disableDbAuditLoggingCmd)
	disableDbAuditLoggingCmd.Flags().SortFlags = false
	disableDbAuditLoggingCmd.Flags().String("cluster-name", "", "[REQUIRED] The cluster name to assign DB auditting")
	disableDbAuditLoggingCmd.MarkFlagRequired("cluster-name")
	disableDbAuditLoggingCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
}

func getIntegrationIdFromName(integrationName string, authApi *ybmAuthClient.AuthApiClient) (string, error) {
	integration, _, err := authApi.ListIntegrations().Name(integrationName).Execute()
	if err != nil {
		return "", err
	}

	integrationData := integration.GetData()

	if len(integrationData) == 0 {
		return "", fmt.Errorf("no integrations found with name: %s%s", integrationName, "\n")
	}

	return integrationData[0].GetInfo().Id, nil
}

func setDbAuditLogsExporterSpec(ysqlConfigMap map[string]string, statementClasses string, integrationId string) (*ybmclient.DbAuditExporterConfigSpec, error) {
	log_catalog := ysqlConfigMap["log_catalog"]
	log_client := ysqlConfigMap["log_client"]
	log_level := ysqlConfigMap["log_level"]
	log_parameter := ysqlConfigMap["log_parameter"]
	log_relation := ysqlConfigMap["log_relation"]
	log_statement_once := ysqlConfigMap["log_statement_once"]

	var statement_classes_enum []ybmclient.DbAuditYsqlStatmentClassesEnum

	if statementClasses != "" {
		for _, statement := range strings.Split(statementClasses, ",") {
			enumVal, err := ybmclient.NewDbAuditYsqlStatmentClassesEnumFromValue(statement)
			if err != nil {
				return nil, err
			}
			statement_classes_enum = append(statement_classes_enum, *enumVal)
		}
	}
	if len(statement_classes_enum) == 0 {
		return nil, fmt.Errorf("statement_classes must have one or more of READ, WRITE, ROLE, FUNCTION, DDL, MISC")
	}

	log_settings := ybmclient.NewDbAuditYsqlLogSettingsWithDefaults()

	if log_catalog != "" {
		catalog, err := strconv.ParseBool(log_catalog)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogCatalog(catalog)
	} else {
		return nil, fmt.Errorf("log_catalog required for log settings")
	}

	if log_client != "" {
		client, err := strconv.ParseBool(log_client)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogClient(client)
	} else {
		return nil, fmt.Errorf("log_client required for log settings")
	}

	if log_level != "" {
		level, err := ybmclient.NewDbAuditLogLevelEnumFromValue(log_level)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogLevel(*level)
	} else {
		return nil, fmt.Errorf("log_level required for log settings")
	}

	if log_parameter != "" {
		parameter, err := strconv.ParseBool(log_parameter)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogParameter(parameter)
	} else {
		return nil, fmt.Errorf("log_parameter required for log settings")
	}

	if log_relation != "" {
		relation, err := strconv.ParseBool(log_relation)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogRelation(relation)
	} else {
		return nil, fmt.Errorf("log_relation required for log settings")
	}

	if log_statement_once != "" {
		statement_once, err := strconv.ParseBool(log_statement_once)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogStatementOnce(statement_once)
	} else {
		return nil, fmt.Errorf("log_statement_once required for log settings")
	}

	ysqlConfig := ybmclient.NewDbAuditYsqlExportConfigWithDefaults()
	if len(statement_classes_enum) > 0 {
		ysqlConfig.SetStatementClasses(statement_classes_enum)
	}

	ysqlConfig.SetLogSettings(*log_settings)

	return ybmclient.NewDbAuditExporterConfigSpec(*ysqlConfig, integrationId), nil
}

func getDbAuditExportConfigIdForCluster(authApi *ybmAuthClient.AuthApiClient, clusterId string) string {
	listResp, r, err := authApi.ListDbAuditExporterConfig(clusterId).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
	}

	if len(listResp.GetData()) < 1 {
		logrus.Fatalf("No DB Audit Log Configuration exists for cluster")
	}

	return listResp.GetData()[0].Info.Id
}
