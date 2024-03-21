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

package db_audit_logs_exporter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	"github.com/yugabyte/ybm-cli/cmd/util"
)

var DbAuditLogsExporterCmd = &cobra.Command{
	Use:   "db-audit-logs-exporter",
	Short: "Manage DB Audit Logs",
	Long:  "Manage DB Audit Logs",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var assignDbAuditLogsExporterCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign DB Audit",
	Long:  "Assign DB Audit Logs to a Cluster",
	Run: func(cmd *cobra.Command, args []string) {

		clusterId, _ := cmd.Flags().GetString("cluster-id")
		integrationId, _ := cmd.Flags().GetString("integration-id")
		ysqlConfig, _ := cmd.Flags().GetStringToString("ysql-config")
		statement_classes, _ := cmd.Flags().GetString("statement_classes")

		dbAuditLogsExporterSpec, err := setDbAuditLogsExporterSpec(ysqlConfig, statement_classes, integrationId)

		if err != nil {
			logrus.Fatalf(err.Error())
		}

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.AssignDbAuditLogsExporterConfig(clusterId).DbAuditExporterConfigSpec(*dbAuditLogsExporterSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		dbAuditIntegrationId := resp.GetData().Info.Id

		msg := fmt.Sprintf("The db audit exporter config %s is being created", formatter.Colorize(dbAuditIntegrationId, formatter.GREEN_COLOR))

		fmt.Println(msg)

		dbAuditLogsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbAuditLogsExporterFormat(viper.GetString("output")),
		}

		formatter.DbAuditLogsExporterWrite(dbAuditLogsExporterCtx, []openapi.DbAuditExporterConfigurationData{resp.GetData()})
	},
}

var updateDbAuditLogsExporterCmd = &cobra.Command{
	Use:   "update",
	Short: "Update DB Audit",
	Long:  "Update DB Audit Log Configuration for a Cluster",
	Run: func(cmd *cobra.Command, args []string) {

		clusterId, _ := cmd.Flags().GetString("cluster-id")
		exportConfigId, _ := cmd.Flags().GetString("export-config-id")
		integrationId, _ := cmd.Flags().GetString("integration-id")
		ysqlConfig, _ := cmd.Flags().GetStringToString("ysql-config")
		statement_classes, _ := cmd.Flags().GetString("statement_classes")

		dbAuditLogsExporterSpec, err := setDbAuditLogsExporterSpec(ysqlConfig, statement_classes, integrationId)

		if err != nil {
			logrus.Fatalf(err.Error())
		}
		
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.UpdateDbAuditExporterConfig(clusterId, exportConfigId).DbAuditExporterConfigSpec(*dbAuditLogsExporterSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		dbAuditIntegrationId := resp.GetData().Info.Id

		msg := fmt.Sprintf("The db audit exporter config %s is being updated", formatter.Colorize(dbAuditIntegrationId, formatter.GREEN_COLOR))

		fmt.Println(msg)

		dbAuditLogsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbAuditLogsExporterFormat(viper.GetString("output")),
		}

		formatter.DbAuditLogsExporterWrite(dbAuditLogsExporterCtx, []openapi.DbAuditExporterConfigurationData{resp.GetData()})
	},
}

var listDbAuditLogsExporterCmd = &cobra.Command{
	Use:   "list",
	Short: "List DB Audit Logs Export Config",
	Long:  "List DB Audit Logs Export Config",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterId, _ := cmd.Flags().GetString("cluster-id")

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

var removeDbAuditLogsExporterCmd = &cobra.Command{
	Use:   "unassign",
	Short: "Unassign DB Audit Logs Export Config",
	Long:  "Unassign DB Audit Logs Export Config",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		exportConfigId, _ := cmd.Flags().GetString("export-config-id")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to unassign DB audit %s: %s", "config", exportConfigId), viper.GetBool("force"))
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

		clusterId, _ := cmd.Flags().GetString("cluster-id")
		exportConfigId, _ := cmd.Flags().GetString("export-config-id")

		resp, err := authApi.UnassignDbAuditLogsExportConfig(clusterId, exportConfigId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Deleting Db Audit Logs Exporter Config %s\n", formatter.Colorize(exportConfigId, formatter.GREEN_COLOR))
	},
}

func init() {
	DbAuditLogsExporterCmd.AddCommand(assignDbAuditLogsExporterCmd)
	assignDbAuditLogsExporterCmd.Flags().SortFlags = false
	assignDbAuditLogsExporterCmd.Flags().String("integration-id", "", "[REQUIRED] The ID of the Integration")
	assignDbAuditLogsExporterCmd.MarkFlagRequired("integration-id")
	assignDbAuditLogsExporterCmd.Flags().StringToString("ysql-config", nil, `[REQUIRED] The ysql config to setup DB auditting
	Please provide key value pairs as follows:
	log_catalog=<boolean>,log_level=<LOG_LEVEL>,log_client=<boolean>,log_parameter=<boolean>,
	log_relation=<boolean>,log_statement_once=<boolean>`)
	assignDbAuditLogsExporterCmd.MarkFlagRequired("ysql-config")
	assignDbAuditLogsExporterCmd.Flags().String("statement_classes", "", `[REQUIRED] The ysql config statement classes
	Please provide key value pairs as follows:
	statement_classes=READ,WRITE,MISC`)
	assignDbAuditLogsExporterCmd.MarkFlagRequired("statement_classes")
	assignDbAuditLogsExporterCmd.Flags().String("cluster-id", "", "[REQUIRED] The cluster ID to assign DB auditting")
	assignDbAuditLogsExporterCmd.MarkFlagRequired("cluster-id")

	DbAuditLogsExporterCmd.AddCommand(listDbAuditLogsExporterCmd)
	listDbAuditLogsExporterCmd.Flags().SortFlags = false
	listDbAuditLogsExporterCmd.Flags().String("cluster-id", "", "[REQUIRED] The cluster ID to list DB audit export config")
	listDbAuditLogsExporterCmd.MarkFlagRequired("cluster-id")

	DbAuditLogsExporterCmd.AddCommand(updateDbAuditLogsExporterCmd)
	updateDbAuditLogsExporterCmd.Flags().SortFlags = false
	updateDbAuditLogsExporterCmd.Flags().String("export-config-id", "", "[REQUIRED] The ID of the DB audit export config")
	updateDbAuditLogsExporterCmd.MarkFlagRequired("export-config-id")
	updateDbAuditLogsExporterCmd.Flags().String("integration-id", "", "[REQUIRED] The ID of the Integration")
	updateDbAuditLogsExporterCmd.MarkFlagRequired("integration-id")
	updateDbAuditLogsExporterCmd.Flags().StringToString("ysql-config", nil, `The ysql config to setup DB auditting
	Please provide key value pairs as follows:
	log_catalog=<boolean>,log_level=<LOG_LEVEL>,log_client=<boolean>,log_parameter=<boolean>,
	log_relation=<boolean>,log_statement_once=<boolean>`)
	updateDbAuditLogsExporterCmd.Flags().String("statement_classes", "", `The ysql config statement classes
	Please provide key value pairs as follows:
	statement_classes=READ,WRITE,MISC`)
	updateDbAuditLogsExporterCmd.Flags().String("cluster-id", "", "[REQUIRED] The cluster ID to assign DB auditting")
	updateDbAuditLogsExporterCmd.MarkFlagRequired("cluster-id")

	DbAuditLogsExporterCmd.AddCommand(removeDbAuditLogsExporterCmd)
	removeDbAuditLogsExporterCmd.Flags().SortFlags = false
	removeDbAuditLogsExporterCmd.Flags().String("export-config-id", "", "[REQUIRED] The ID of the DB audit export config")
	removeDbAuditLogsExporterCmd.MarkFlagRequired("export-config-id")
	removeDbAuditLogsExporterCmd.Flags().String("cluster-id", "", "[REQUIRED] The cluster ID to assign DB auditting")
	removeDbAuditLogsExporterCmd.MarkFlagRequired("cluster-id")
	removeDbAuditLogsExporterCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
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
	}else{
		return nil, fmt.Errorf("log_catalog required for log settings")
	}

	if log_client != "" {
		client, err := strconv.ParseBool(log_client)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogClient(client)
	}else{
		return nil, fmt.Errorf("log_client required for log settings")
	}

	if log_level != "" {
		level, err := ybmclient.NewDbAuditLogLevelEnumFromValue(log_level)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogLevel(*level)
	}else{
		return nil, fmt.Errorf("log_level required for log settings")
	}

	if log_parameter != "" {
		parameter, err := strconv.ParseBool(log_parameter)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogParameter(parameter)
	}else{
		return nil, fmt.Errorf("log_parameter required for log settings")
	}

	if log_relation != "" {
		relation, err := strconv.ParseBool(log_relation)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogRelation(relation)
	}else{
		return nil, fmt.Errorf("log_relation required for log settings")
	}

	if log_statement_once != "" {
		statement_once, err := strconv.ParseBool(log_statement_once)
		if err != nil {
			return nil, err
		}
		log_settings.SetLogStatementOnce(statement_once)
	}else{
		return nil, fmt.Errorf("log_statement_once required for log settings")
	}

	ysqlConfig := ybmclient.NewDbAuditYsqlExportConfigWithDefaults()
	if len(statement_classes_enum) > 0 {
		ysqlConfig.SetStatementClasses(statement_classes_enum)
	}

	ysqlConfig.SetLogSettings(*log_settings)

	return ybmclient.NewDbAuditExporterConfigSpec(*ysqlConfig, integrationId), nil
}
