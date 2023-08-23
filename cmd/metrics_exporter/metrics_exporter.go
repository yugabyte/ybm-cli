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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	},
}

func init() {
	MetricsExporterCmd.AddCommand(createMetricsExporterCmd)

	createMetricsExporterCmd.Flags().String("name", "", "[REQUIRED] The name of the cluster.")
	createMetricsExporterCmd.MarkFlagRequired("name")
	createMetricsExporterCmd.Flags().String("type", "", "[REQUIRED] The type of third party metrics sink")
	createMetricsExporterCmd.MarkFlagRequired("type")
	createMetricsExporterCmd.Flags().StringToString("datadog-spec", nil, "Spec for datadog")
}
