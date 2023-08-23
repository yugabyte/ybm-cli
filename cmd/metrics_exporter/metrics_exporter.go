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

package metrics_exporter

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

var MetricsExporterCmd = &cobra.Command{
	Use:   "metrics-exporter",
	Short: "Manage Metrics Exporter",
	Long:  "Manage Metrics Exporter",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var createMetricsExporterCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Metrics Exporter Config",
	Long:  "Create Metrics Exporter Config",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		metricsExporterName, _ := cmd.Flags().GetString("name")
		metricsSinkType, _ := cmd.Flags().GetString("type")
		datadogSpecString, _ := cmd.Flags().GetStringToString("datadog-spec")

		metricsSinkTypeEnum, err := ybmclient.NewMetricsExporterConfigTypeEnumFromValue(metricsSinkType)

		apiKey := datadogSpecString["api-key"]
		site := datadogSpecString["site"]

		datadogSpec := ybmclient.NewDatadogMetricsExporterConfigurationSpec(apiKey, site)
		metricsExporterConfigSpec := ybmclient.NewMetricsExporterConfigurationSpec(metricsExporterName, *metricsSinkTypeEnum)

		metricsExporterConfigSpec.SetDatadogSpec(*datadogSpec)

		resp, r, err := authApi.CreateMetricsExporterConfig().MetricsExporterConfigurationSpec(*metricsExporterConfigSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		metricsExporterId := resp.GetData().Info.Id

		msg := fmt.Sprintf("The metrics exporter config %s is being created", formatter.Colorize(metricsExporterId, formatter.GREEN_COLOR))

		fmt.Println(msg)

		metricsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewMetricsExporterFormat(viper.GetString("output")),
		}

		respArr := [1]ybmclient.MetricsExporterConfigurationData{resp.GetData()}

		formatter.MetricsExporterWrite(metricsExporterCtx, respArr[:])
	},
}

var listMetricsExporterCmd = &cobra.Command{
	Use:   "list",
	Short: "List Metrics Exporter Config",
	Long:  "List Metrics Exporter Config",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.ListMetricsExporterConfigs().Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		metricsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewMetricsExporterFormat(viper.GetString("output")),
		}

		if len(resp.GetData()) < 1 {
			fmt.Println("No metrics exporters found")
			return
		}

		formatter.MetricsExporterWrite(metricsExporterCtx, resp.GetData())
	},
}

var deleteMetricsExporterCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Metrics Exporter Config",
	Long:  "Delete Metrics Exporter Config",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		configName, _ := cmd.Flags().GetString("config-name")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to delete %s: %s", "config", configName), viper.GetBool("force"))
		if err != nil {
			logrus.Fatal(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		configName, _ := cmd.Flags().GetString("config-name")

		resp, r, err := authApi.ListMetricsExporterConfigs().Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		configId := ""

		for _, metricsExporter := range resp.Data {
			if metricsExporter.GetSpec().Name == configName {
				configId = metricsExporter.GetInfo().Id
				break
			}
		}

		if configId == "" {
			logrus.Fatalf("Could not find config with name %s", configName)
		}

		r1, err := authApi.DeleteMetricsExporterConfig(configId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r1)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Deleting Metrics Exporter Config %s", configName)
		fmt.Println()
	},
}

var removeMetricsExporterFromClusterCmd = &cobra.Command{
	Use:   "remove-from-cluster",
	Short: "Remove Metrics Exporter Config from Cluster",
	Long:  "Remove Metrics Exporter Config from Cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		r, err := authApi.RemoveMetricsExporterConfigFromCluster(clusterId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Removing associated Metrics Exporter Config from cluster %s", clusterName)
		fmt.Println()
	},
}

var associateMetricsExporterWithClusterCmd = &cobra.Command{
	Use:   "attach",
	Short: "Associate Metrics Exporter Config with Cluster",
	Long:  "Associate Metrics Exporter Config with Cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		configName, _ := cmd.Flags().GetString("config-name")

		resp, r, err := authApi.ListMetricsExporterConfigs().Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		configId := ""

		for _, metricsExporter := range resp.Data {
			if metricsExporter.GetSpec().Name == configName {
				configId = metricsExporter.GetInfo().Id
				break
			}
		}

		if configId == "" {
			logrus.Fatalf("Could not find config with name %s", configName)
		}

		metricsExporterClusterConfigSpec := ybmclient.NewMetricsExporterClusterConfigurationSpec(configId)

		resp1, r, err := authApi.AssociateMetricsExporterWithCluster(clusterId).MetricsExporterClusterConfigurationSpec(*metricsExporterClusterConfigSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Attaching Metrics Exporter Config %s with cluster %s", resp1.Data.Spec.Name, clusterName)
		fmt.Println()
	},
}

var stopMetricsExporterCmd = &cobra.Command{
	Use:   "pause",
	Short: "Stop Metrics Exporter",
	Long:  "Stop Metrics Exporter",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		r, err := authApi.StopMetricsExporter(clusterId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Stopping Metrics Exporter for cluster %s", clusterName)
		fmt.Println()
	},
}

var updateMetricsExporterCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Metrics Exporter Config",
	Long:  "Update Metrics Exporter Config",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		metricsExporterName, _ := cmd.Flags().GetString("config-name")
		metricsSinkType, _ := cmd.Flags().GetString("type")
		datadogSpecString, _ := cmd.Flags().GetStringToString("datadog-spec")

		metricsSinkTypeEnum, err := ybmclient.NewMetricsExporterConfigTypeEnumFromValue(metricsSinkType)

		apiKey := datadogSpecString["api-key"]
		site := datadogSpecString["site"]

		datadogSpec := ybmclient.NewDatadogMetricsExporterConfigurationSpec(apiKey, site)
		metricsExporterConfigSpec := ybmclient.NewMetricsExporterConfigurationSpec(metricsExporterName, *metricsSinkTypeEnum)

		metricsExporterConfigSpec.SetDatadogSpec(*datadogSpec)

		resp, r, err := authApi.ListMetricsExporterConfigs().Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		configId := ""

		for _, metricsExporter := range resp.Data {
			if metricsExporter.GetSpec().Name == metricsExporterName {
				configId = metricsExporter.GetInfo().Id
				break
			}
		}

		if configId == "" {
			logrus.Fatalf("Could not find config with name %s", metricsExporterName)
		}

		resp1, r, err := authApi.UpdateMetricsExporterConfig(configId).MetricsExporterConfigurationSpec(*metricsExporterConfigSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The metrics exporter config %s is being updated", formatter.Colorize(configId, formatter.GREEN_COLOR))

		fmt.Println(msg)

		metricsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewMetricsExporterFormat(viper.GetString("output")),
		}

		respArr := [1]ybmclient.MetricsExporterConfigurationData{resp1.GetData()}

		formatter.MetricsExporterWrite(metricsExporterCtx, respArr[:])
	},
}

func init() {
	MetricsExporterCmd.AddCommand(createMetricsExporterCmd)
	createMetricsExporterCmd.Flags().String("name", "", "[REQUIRED] The name of the cluster.")
	createMetricsExporterCmd.MarkFlagRequired("name")
	createMetricsExporterCmd.Flags().String("type", "", "[REQUIRED] The type of third party metrics sink")
	createMetricsExporterCmd.MarkFlagRequired("type")
	createMetricsExporterCmd.Flags().StringToString("datadog-spec", nil, "Spec for datadog")

	MetricsExporterCmd.AddCommand(listMetricsExporterCmd)

	MetricsExporterCmd.AddCommand(deleteMetricsExporterCmd)
	deleteMetricsExporterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the metrics exporter config")
	deleteMetricsExporterCmd.MarkFlagRequired("config-name")

	MetricsExporterCmd.AddCommand(removeMetricsExporterFromClusterCmd)
	removeMetricsExporterFromClusterCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	removeMetricsExporterFromClusterCmd.MarkFlagRequired("cluster-name")

	MetricsExporterCmd.AddCommand(associateMetricsExporterWithClusterCmd)
	associateMetricsExporterWithClusterCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	associateMetricsExporterWithClusterCmd.MarkFlagRequired("cluster-name")
	associateMetricsExporterWithClusterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the metrics exporter config")
	associateMetricsExporterWithClusterCmd.MarkFlagRequired("config-name")

	MetricsExporterCmd.AddCommand(stopMetricsExporterCmd)
	stopMetricsExporterCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	stopMetricsExporterCmd.MarkFlagRequired("cluster-name")

	MetricsExporterCmd.AddCommand(updateMetricsExporterCmd)
	updateMetricsExporterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the cluster.")
	updateMetricsExporterCmd.MarkFlagRequired("config-name")
	updateMetricsExporterCmd.Flags().String("type", "", "[REQUIRED] The type of third party metrics sink")
	updateMetricsExporterCmd.MarkFlagRequired("type")
	updateMetricsExporterCmd.Flags().StringToString("datadog-spec", nil, "Spec for datadog")
}
