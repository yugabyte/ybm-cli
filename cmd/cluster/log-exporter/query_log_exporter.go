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

package query_log_exporter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var ClusterName string

var DbQueryLoggingCmd = &cobra.Command{
	Use:   "db-query-log-exporter",
	Short: "Configure DB Query Log exporter for your Cluster.",
	Long:  "Configure DB Query Log exporter for your Cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var enableDbQueryLoggingCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable DB query log exporter",
	Long:  "Enable DB query log exporter",
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
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		integrationId, err := authApi.GetIntegrationIdFromName(integrationName)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		exportConfig := BuildNewPgExportConfig(cmd)

		resp, r, err := authApi.EnableDbQueryLogging(clusterId).PgLogExporterConfigSpec(
			ybmclient.PgLogExporterConfigSpec{ExportConfig: exportConfig, ExporterId: integrationId}).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v\n", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		DbQueryLoggingContext := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbQueryLoggingFormat(),
		}

		err = formatter.DbQueryLoggingWrite(DbQueryLoggingContext, []ybmclient.PgLogExporterConfigData{resp.GetData()})
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}

var describeLogExporterCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe DB query log exporter config",
	Long:  "Describe DB query log exporter config",
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

		resp, r, err := authApi.GetDbLoggingConfig(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response for Query Log Exporter Config: %v\n", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			fmt.Printf("DB query logs are not enabled for cluster: %s\n", clusterName)
			return
		}

		DbQueryLoggingContext := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbQueryLoggingFormat(),
		}

		err = formatter.DbQueryLoggingWrite(DbQueryLoggingContext, resp.GetData())
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}

var disableLogExporterCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable DB query log exporter",
	Long:  "Disable DB query log exporter, if enabled",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName == "" {
			logrus.Fatalf("cluster-name must not be empty")
		}

		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		// Fetch existing log exporter config
		resp, r, err := authApi.GetDbLoggingConfig(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response for Query Log Exporter Config: %v\n", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			fmt.Printf("DB query logs are not enabled for cluster: %s\n", clusterName)
			return
		}

		logExporterData := resp.GetData()[0]
		exporterConfigId := logExporterData.Info.Id

		r, err = authApi.RemoveDbQueryLoggingConfig(clusterId, exporterConfigId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response for disable query logging config: %v\n", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf(`Request submitted to disable DB query logging for the cluster, this may take a few minutes...
You can check the status via $ ybm cluster db-query-logging describe --cluster-name %s%s`, formatter.Colorize(clusterName, formatter.GREEN_COLOR), "\n")
	},
}

var updateLogExporterConfigCmd = &cobra.Command{
	Use:   "update",
	Short: "Update DB query log exporter config",
	Long:  "Update DB query log exporter config. Only the config values that are passed in args will be updated, the remaining one's will remain same as existing config.",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName == "" {
			logrus.Fatalf("cluster-name must not be empty")
		}

		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		// Fetch existing log exporter config
		resp, r, err := authApi.GetDbLoggingConfig(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response for Query Log Exporter Config: %v\n", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			fmt.Printf("DB query logs are not enabled for cluster: %s\n", clusterName)
			return
		}

		logExporterData := resp.GetData()[0]
		exporterConfigId := logExporterData.Info.Id

		var integrationId string = ""
		// use integration name if provided by user, else use existing one
		if cmd.Flags().Changed("integration-name") {
			integrationName, _ := cmd.Flags().GetString("integration-name")
			integrationId, err = authApi.GetIntegrationIdFromName(integrationName)
			if err != nil {
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
		} else {
			integrationId = logExporterData.Spec.ExporterId
		}

		existingExportConfig := logExporterData.Spec.ExportConfig
		newExportConfig := BuildNewPgExportConfigFromExistingConfig(cmd, existingExportConfig)

		var pgLogExporterConfigResponse ybmclient.PgLogExporterConfigResponse
		pgLogExporterConfigResponse, r, err = authApi.EditDbQueryLoggingConfig(clusterId, exporterConfigId).PgLogExporterConfigSpec(
			ybmclient.PgLogExporterConfigSpec{ExportConfig: newExportConfig, ExporterId: integrationId}).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response for update DB Query logging config: %v\n", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		DbQueryLoggingContext := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewDbQueryLoggingFormat(),
		}

		err = formatter.DbQueryLoggingWrite(DbQueryLoggingContext,
			[]ybmclient.PgLogExporterConfigData{pgLogExporterConfigResponse.GetData()})
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {

	DbQueryLoggingCmd.AddCommand(enableDbQueryLoggingCmd)
	enableDbQueryLoggingCmd.Flags().SortFlags = false
	enableDbQueryLoggingCmd.Flags().String("integration-name", "", "[REQUIRED] Name of the Integration")
	enableDbQueryLoggingCmd.MarkFlagRequired("integration-name")
	enableDbQueryLoggingCmd.Flags().String("debug-print-plan", "false", "[OPTIONAL] Enables various debugging output to be emitted.")
	enableDbQueryLoggingCmd.Flags().Int32("log-min-duration-statement", -1, "[OPTIONAL] Duration(in ms) of each completed statement to be logged if the statement ran for at least the specified amount of time. Default -1 (log all statements).")
	enableDbQueryLoggingCmd.Flags().String("log-connections", "false", "[OPTIONAL] Log connection attempts.")
	enableDbQueryLoggingCmd.Flags().String("log-disconnections", "false", "[OPTIONAL] Log session disconnections.")
	enableDbQueryLoggingCmd.Flags().String("log-duration", "false", "[OPTIONAL] Log the duration of each completed statement.")
	enableDbQueryLoggingCmd.Flags().String("log-error-verbosity", "DEFAULT", "[OPTIONAL] Controls the amount of detail written in the server log for each message that is logged. Options: DEFAULT, TERSE, VERBOSE.")
	enableDbQueryLoggingCmd.Flags().String("log-statement", "NONE", "[OPTIONAL] Log all statements or specific types of statements. Options: NONE, DDL, MOD, ALL.")
	enableDbQueryLoggingCmd.Flags().String("log-min-error-statement", "ERROR", "[OPTIONAL] Minimum error severity for logging the statement that caused it. Options: ERROR.")
	enableDbQueryLoggingCmd.Flags().String("log-line-prefix", "%m :%r :%u @ %d :[%p] :", "[OPTIONAL] A printf-style format string for log line prefixes.")

	DbQueryLoggingCmd.AddCommand(updateLogExporterConfigCmd)
	updateLogExporterConfigCmd.Flags().String("integration-name", "", "[OPTIONAL] Name of the Integration")
	updateLogExporterConfigCmd.Flags().String("debug-print-plan", "", "[OPTIONAL] Enables various debugging output to be emitted.")
	updateLogExporterConfigCmd.Flags().Int32("log-min-duration-statement", -1, "[OPTIONAL] Duration(in ms) of each completed statement to be logged if the statement ran for at least the specified amount of time.")
	updateLogExporterConfigCmd.Flags().String("log-connections", "", "[OPTIONAL] Log connection attempts.")
	updateLogExporterConfigCmd.Flags().String("log-disconnections", "", "[OPTIONAL] Log session disconnections.")
	updateLogExporterConfigCmd.Flags().String("log-duration", "", "[OPTIONAL] Log the duration of each completed statement.")
	updateLogExporterConfigCmd.Flags().String("log-error-verbosity", "", "[OPTIONAL] Controls the amount of detail written in the server log for each message that is logged. Options: DEFAULT, TERSE, VERBOSE.")
	updateLogExporterConfigCmd.Flags().String("log-statement", "", "[OPTIONAL] Log all statements or specific types of statements. Options: NONE, DDL, MOD, ALL.")
	updateLogExporterConfigCmd.Flags().String("log-min-error-statement", "", "[OPTIONAL] Minimum error severity for logging the statement that caused it. Options: ERROR.")
	updateLogExporterConfigCmd.Flags().String("log-line-prefix", "", "[OPTIONAL] A printf-style format string for log line prefixes.")

	DbQueryLoggingCmd.AddCommand(describeLogExporterCmd)
	DbQueryLoggingCmd.AddCommand(disableLogExporterCmd)
}

func BuildNewPgExportConfig(cmd *cobra.Command) ybmclient.PgLogExportConfig {
	// Build a new PgLogExportConfig from args
	config := ybmclient.PgLogExportConfig{}

	logMinDurationStatement, _ := cmd.Flags().GetInt32("log-min-duration-statement")
	config.LogMinDurationStatement = logMinDurationStatement

	debugPrintPlan, _ := cmd.Flags().GetString("debug-print-plan")
	config.DebugPrintPlan = ParseBoolString(debugPrintPlan)

	logConnections, _ := cmd.Flags().GetString("log-connections")
	config.LogConnections = ParseBoolString(logConnections)

	logDisconnections, _ := cmd.Flags().GetString("log-disconnections")
	config.LogDisconnections = ParseBoolString(logDisconnections)

	logDuration, _ := cmd.Flags().GetString("log-duration")
	config.LogDuration = ParseBoolString(logDuration)

	if logErrorVerbosity, _ := cmd.Flags().GetString("log-error-verbosity"); logErrorVerbosity != "" {
		logErrorVerbosityEnum, err := ybmclient.NewLogErrorVerbosityEnumFromValue(strings.ToUpper(logErrorVerbosity))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		config.LogErrorVerbosity = *logErrorVerbosityEnum
	}

	if logStatement, _ := cmd.Flags().GetString("log-statement"); logStatement != "" {
		logStatementEnum, err := ybmclient.NewLogStatementEnumFromValue(strings.ToUpper(logStatement))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		config.LogStatement = *logStatementEnum
	}

	if logMinErrorStatement, _ := cmd.Flags().GetString("log-min-error-statement"); logMinErrorStatement != "" {
		logMinErrorStatementEnum, err := ybmclient.NewLogMinErrorStatementEnumFromValue(strings.ToUpper(logMinErrorStatement))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		config.LogMinErrorStatement = *logMinErrorStatementEnum
	}

	if logLinePrefix, _ := cmd.Flags().GetString("log-line-prefix"); logLinePrefix != "" {
		config.LogLinePrefix = logLinePrefix
	}

	return config
}

func BuildNewPgExportConfigFromExistingConfig(cmd *cobra.Command, existingConfig ybmclient.PgLogExportConfig) ybmclient.PgLogExportConfig {
	// Copy existing config and only update the fields that are explicitly provided in args
	// This is to ensure that we do not set the flags(which are not provided in args) back to default values.
	var newConfig = existingConfig

	if cmd.Flags().Changed("log-min-duration-statement") {
		logMinDurationStatement, _ := cmd.Flags().GetInt32("log-min-duration-statement")
		newConfig.LogMinDurationStatement = logMinDurationStatement
	}

	if cmd.Flags().Changed("debug-print-plan") {
		debugPrintPlan, _ := cmd.Flags().GetString("debug-print-plan")
		newConfig.DebugPrintPlan = ParseBoolString(debugPrintPlan)
	}

	if cmd.Flags().Changed("log-connections") {
		logConnections, _ := cmd.Flags().GetString("log-connections")
		newConfig.LogConnections = ParseBoolString(logConnections)
	}

	if cmd.Flags().Changed("log-disconnections") {
		logDisconnections, _ := cmd.Flags().GetString("log-disconnections")
		newConfig.LogDisconnections = ParseBoolString(logDisconnections)
	}

	if cmd.Flags().Changed("log-duration") {
		logDuration, _ := cmd.Flags().GetString("log-duration")
		newConfig.LogDuration = ParseBoolString(logDuration)
	}

	if cmd.Flags().Changed("log-line-prefix") {
		logLinePrefix, _ := cmd.Flags().GetString("log-line-prefix")
		newConfig.LogLinePrefix = logLinePrefix
	}

	if cmd.Flags().Changed("log-error-verbosity") {
		logErrorVerbosity, _ := cmd.Flags().GetString("log-error-verbosity")
		logErrorVerbosityEnum, err := ybmclient.NewLogErrorVerbosityEnumFromValue(strings.ToUpper(logErrorVerbosity))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		newConfig.LogErrorVerbosity = *logErrorVerbosityEnum
	}

	if cmd.Flags().Changed("log-statement") {
		logStatement, _ := cmd.Flags().GetString("log-statement")
		logStatementEnum, err := ybmclient.NewLogStatementEnumFromValue(strings.ToUpper(logStatement))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		newConfig.LogStatement = *logStatementEnum
	}

	if cmd.Flags().Changed("log-min-error-statement") {
		logMinErrorStatement, _ := cmd.Flags().GetString("log-min-error-statement")
		logMinErrorStatementEnum, err := ybmclient.NewLogMinErrorStatementEnumFromValue(strings.ToUpper(logMinErrorStatement))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		newConfig.LogMinErrorStatement = *logMinErrorStatementEnum
	}

	return newConfig
}

func ParseBoolString(input string) bool {
	result, err := strconv.ParseBool(input)
	if err != nil {
		logrus.Fatalf("invalid boolean value \"%s\": expected true/false or 1/0", input)
	}
	return result
}
