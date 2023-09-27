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
	"strings"

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

		metricsExporterName, _ := cmd.Flags().GetString("config-name")
		metricsSinkType, _ := cmd.Flags().GetString("type")

		//We initialise here, even if we error out later
		metricsSinkTypeEnum, err := ybmclient.NewMetricsExporterConfigTypeEnumFromValue(strings.ToUpper(metricsSinkType))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		metricsExporterConfigSpec := ybmclient.NewMetricsExporterConfigurationSpec(metricsExporterName, *metricsSinkTypeEnum)

		switch *metricsSinkTypeEnum {
		case ybmclient.METRICSEXPORTERCONFIGTYPEENUM_DATADOG:
			if !cmd.Flags().Changed("datadog-spec") {
				logrus.Fatalf("datadog-spec is required for datadog sink")

			}
			datadogSpecString, _ := cmd.Flags().GetStringToString("datadog-spec")
			apiKey := datadogSpecString["api-key"]
			site := datadogSpecString["site"]
			if len(apiKey) < 1 {
				logrus.Fatal("api-key is a required field for datadog-spec")
			}
			if len(site) < 1 {
				logrus.Fatal("site is a required field for datadog-spec")
			}
			datadogSpec := ybmclient.NewDatadogMetricsExporterConfigurationSpec(apiKey, site)
			metricsExporterConfigSpec.SetDatadogSpec(*datadogSpec)
		case ybmclient.METRICSEXPORTERCONFIGTYPEENUM_GRAFANA:
			if !cmd.Flags().Changed("grafana-spec") {
				logrus.Fatalf("grafana-spec is required for grafana sink")

			}
			grafanaSpecString, _ := cmd.Flags().GetStringToString("grafana-spec")
			apiKey := grafanaSpecString["access-policy-token"]
			zone := grafanaSpecString["zone"]
			instanceId := grafanaSpecString["instance-id"]
			orgSlug := grafanaSpecString["org-slug"]
			if len(apiKey) < 1 {
				logrus.Fatal("access-policy-token is a required field for grafana-spec")
			}
			if len(zone) < 1 {
				logrus.Fatal("Zone is a required field for grafana-spec")
			}
			if len(instanceId) < 1 {
				logrus.Fatal("instance-id is a required field for grafana-spec")
			}
			if len(orgSlug) < 1 {
				logrus.Fatal("org-slug is a required field for grafana-spec")
			}

			grafanaSpec := ybmclient.NewGrafanaMetricsExporterConfigurationSpec(apiKey, zone, instanceId, orgSlug)
			metricsExporterConfigSpec.SetGrafanaSpec(*grafanaSpec)
		default:
			//We should never go there normally
			logrus.Fatalf("Only datadog is accepted as third party sink for now")
		}

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

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
			Format: formatter.NewMetricsExporterFormat(viper.GetString("output"), string(resp.GetData().Spec.Type)),
		}

		respArr := []ybmclient.MetricsExporterConfigurationData{resp.GetData()}

		formatter.MetricsExporterWrite(metricsExporterCtx, respArr)
	},
}

var listMetricsExporterCmd = &cobra.Command{
	Use:   "list",
	Short: "List Metrics Exporter Config",
	Long:  "List Metrics Exporter Config",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.ListMetricsExporterConfigs().Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		metricsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewMetricsExporterFormat(viper.GetString("output"), ""),
		}

		if len(resp.GetData()) < 1 {
			fmt.Println("No metrics exporters found")
			return
		}

		formatter.MetricsExporterWrite(metricsExporterCtx, resp.GetData())
	},
}

var describeMetricsExporterCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe Metrics Exporter Config",
	Long:  "Describe Metrics Exporter Config",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		metricsExporterName, _ := cmd.Flags().GetString("config-name")
		config, err := authApi.GetConfigByName(metricsExporterName)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		metricsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewMetricsExporterFormat(viper.GetString("output"), string(config.Spec.GetType())),
		}

		formatter.MetricsExporterWrite(metricsExporterCtx, []ybmclient.MetricsExporterConfigurationData{*config})
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
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		configName, _ := cmd.Flags().GetString("config-name")

		config, err := authApi.GetConfigByName(configName)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		r1, err := authApi.DeleteMetricsExporterConfig(config.GetInfo().Id).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r1)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Deleting Metrics Exporter Config %s\n", formatter.Colorize(configName, formatter.GREEN_COLOR))
	},
}

var removeMetricsExporterFromClusterCmd = &cobra.Command{
	Use:     "unassign",
	Aliases: []string{"remove-from-cluster"},
	Short:   "Unassign Metrics Exporter Config from Cluster",
	Long:    "Unassign Metrics Exporter Config from Cluster",
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

		r, err := authApi.RemoveMetricsExporterConfigFromCluster(clusterId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Unassigning associated Metrics Exporter Config from cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
	},
}

var associateMetricsExporterWithClusterCmd = &cobra.Command{
	Use:     "assign",
	Aliases: []string{"attach"},
	Short:   "Associate Metrics Exporter Config with Cluster",
	Long:    "Associate Metrics Exporter Config with Cluster",
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

		configName, _ := cmd.Flags().GetString("config-name")

		config, err := authApi.GetConfigByName(configName)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		metricsExporterClusterConfigSpec := ybmclient.NewMetricsExporterClusterConfigurationSpec(config.GetInfo().Id)

		_, r, err := authApi.AssociateMetricsExporterWithCluster(clusterId).MetricsExporterClusterConfigurationSpec(*metricsExporterClusterConfigSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Assigning Metrics Exporter Config %s with cluster %s\n", formatter.Colorize(configName, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))
	},
}

var stopMetricsExporterCmd = &cobra.Command{
	Use:   "pause",
	Short: "Stop Metrics Exporter",
	Long:  "Stop Metrics Exporter",
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

		r, err := authApi.StopMetricsExporter(clusterId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Stopping Metrics Exporter for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
	},
}

var updateMetricsExporterCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Metrics Exporter Config",
	Long:  "Update Metrics Exporter Config",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		metricsExporterName, _ := cmd.Flags().GetString("config-name")
		metricsSinkType, _ := cmd.Flags().GetString("type")

		metricsSinkTypeEnum, err := ybmclient.NewMetricsExporterConfigTypeEnumFromValue(strings.ToUpper(metricsSinkType))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		oldname := metricsExporterName
		if cmd.Flags().Changed("new-config-name") {
			metricsExporterName, _ = cmd.Flags().GetString("new-config-name")
		}
		//We initialise this one here, even if we error out later
		metricsExporterConfigSpec := ybmclient.NewMetricsExporterConfigurationSpec(metricsExporterName, *metricsSinkTypeEnum)

		switch *metricsSinkTypeEnum {
		case ybmclient.METRICSEXPORTERCONFIGTYPEENUM_DATADOG:
			if !cmd.Flags().Changed("datadog-spec") {
				logrus.Fatalf("datadog-spec is required for datadog sink")

			}
			datadogSpecString, _ := cmd.Flags().GetStringToString("datadog-spec")
			apiKey := datadogSpecString["api-key"]
			site := datadogSpecString["site"]
			if len(apiKey) < 1 {
				logrus.Fatal("api-key is a required field for datadog-spec")
			}
			if len(site) < 1 {
				logrus.Fatal("site is a required field for datadog-spec")
			}
			datadogSpec := ybmclient.NewDatadogMetricsExporterConfigurationSpec(apiKey, site)
			metricsExporterConfigSpec.SetDatadogSpec(*datadogSpec)
		case ybmclient.METRICSEXPORTERCONFIGTYPEENUM_GRAFANA:
			if !cmd.Flags().Changed("grafana-spec") {
				logrus.Fatalf("grafana-spec is required for grafana sink")

			}
			grafanaSpecString, _ := cmd.Flags().GetStringToString("grafana-spec")
			apiKey := grafanaSpecString["access-policy-token"]
			zone := grafanaSpecString["zone"]
			instanceId := grafanaSpecString["instance-id"]
			orgSlug := grafanaSpecString["org-slug"]
			if len(apiKey) < 1 {
				logrus.Fatal("access-policy-token is a required field for grafana-spec")
			}
			if len(zone) < 1 {
				logrus.Fatal("zone is a required field for grafana-spec")
			}
			if len(instanceId) < 1 {
				logrus.Fatal("instance-id is a required field for grafana-spec")
			}
			if len(orgSlug) < 1 {
				logrus.Fatal("org-slug is a required field for grafana-spec")
			}

			grafanaSpec := ybmclient.NewGrafanaMetricsExporterConfigurationSpec(apiKey, zone, instanceId, orgSlug)
			metricsExporterConfigSpec.SetGrafanaSpec(*grafanaSpec)
		default:
			logrus.Fatalf("Only datadog is accepted as third party sink for now")
		}

		config, err := authApi.GetConfigByName(oldname)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.UpdateMetricsExporterConfig(config.GetInfo().Id).MetricsExporterConfigurationSpec(*metricsExporterConfigSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The metrics exporter config %s is being updated", formatter.Colorize(config.GetInfo().Id, formatter.GREEN_COLOR))

		fmt.Println(msg)

		metricsExporterCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewMetricsExporterFormat(viper.GetString("output"), string(resp.GetData().Spec.Type)),
		}

		respArr := []ybmclient.MetricsExporterConfigurationData{resp.GetData()}

		formatter.MetricsExporterWrite(metricsExporterCtx, respArr)
	},
}

func init() {
	MetricsExporterCmd.AddCommand(createMetricsExporterCmd)
	createMetricsExporterCmd.Flags().SortFlags = false
	createMetricsExporterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the metrics exporter configuration")
	createMetricsExporterCmd.MarkFlagRequired("config-name")
	createMetricsExporterCmd.Flags().String("type", "", "[REQUIRED] The type of third party metrics sink")
	createMetricsExporterCmd.MarkFlagRequired("type")
	createMetricsExporterCmd.Flags().StringToString("datadog-spec", nil, `Configuration for Datadog. 
	Please provide key value pairs as follows: 
	api-key=<your-datadog-api-key>,site=<your-datadog-site-parameters>`)
	createMetricsExporterCmd.Flags().StringToString("grafana-spec", nil, `Configuration for Grafana. 
	Please provide key value pairs as follows: 
	access-policy-token=<your-grafana-token>,zone=<your-grafana-zone-parameter>,instance-id=<your-grafana-instance-id>,org-slug=<your-grafana-org-slug>`)

	MetricsExporterCmd.AddCommand(listMetricsExporterCmd)

	MetricsExporterCmd.AddCommand(describeMetricsExporterCmd)
	describeMetricsExporterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the metrics exporter configuration")
	describeMetricsExporterCmd.MarkFlagRequired("config-name")

	MetricsExporterCmd.AddCommand(deleteMetricsExporterCmd)
	deleteMetricsExporterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the metrics exporter configuration")
	deleteMetricsExporterCmd.MarkFlagRequired("config-name")

	MetricsExporterCmd.AddCommand(removeMetricsExporterFromClusterCmd)
	removeMetricsExporterFromClusterCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster")
	removeMetricsExporterFromClusterCmd.MarkFlagRequired("cluster-name")

	MetricsExporterCmd.AddCommand(associateMetricsExporterWithClusterCmd)
	associateMetricsExporterWithClusterCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	associateMetricsExporterWithClusterCmd.MarkFlagRequired("cluster-name")
	associateMetricsExporterWithClusterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the metrics exporter configuration")
	associateMetricsExporterWithClusterCmd.MarkFlagRequired("config-name")

	MetricsExporterCmd.AddCommand(stopMetricsExporterCmd)
	stopMetricsExporterCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster.")
	stopMetricsExporterCmd.MarkFlagRequired("cluster-name")

	MetricsExporterCmd.AddCommand(updateMetricsExporterCmd)
	updateMetricsExporterCmd.Flags().SortFlags = false
	updateMetricsExporterCmd.Flags().String("config-name", "", "[REQUIRED] The name of the metrics exporter configuration")
	updateMetricsExporterCmd.MarkFlagRequired("config-name")
	updateMetricsExporterCmd.Flags().String("type", "", "[REQUIRED] The type of third party metrics sink")
	updateMetricsExporterCmd.MarkFlagRequired("type")
	updateMetricsExporterCmd.Flags().StringToString("datadog-spec", nil, `Configuration for Datadog. 
	Please provide key value pairs as follows: 
	api-key=<your-datadog-api-key>,site=<your-datadog-site-parameters>`)
	updateMetricsExporterCmd.Flags().StringToString("grafana-spec", nil, `Configuration for Grafana. 
	Please provide key value pairs as follows: 
	access-policy-token=<your-grafana-token>,zone=<your-grafana-zone-parameter>,instance-id=<your-grafana-instance-id>,org-slug=<your-grafana-org-slug>`)
	updateMetricsExporterCmd.Flags().String("new-config-name", "", "[OPTIONAL] The new name of the metrics exporter configuration")
}
