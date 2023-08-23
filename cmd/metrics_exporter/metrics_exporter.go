package metrics_exporter

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
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
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		metricsExporterName, _ := cmd.Flags().GetString("name")
		metricsSinkType, _ := cmd.Flags().GetString("type")
		datadogSpecString, _ := cmd.Flags().GetStringToString("datadog-spec")

		metricsSinkTypeEnum, err := ybmclient.NewMetricsExporterConfigTypeEnumFromValue(metricsSinkType)
		if err != nil {
			logrus.Fatalf(err.Error())
		}

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

		resp.GetData()

	},
}
