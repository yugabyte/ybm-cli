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

package integration

import (
	"fmt"
	"os"
	"strings"

	"encoding/json"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var IntegrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "Manage Integration",
	Long:  "Manage Integration",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var createIntegrationCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Integration",
	Long:  "Create Integration",
	Run: func(cmd *cobra.Command, args []string) {

		IntegrationName, _ := cmd.Flags().GetString("config-name")
		sinkType, _ := cmd.Flags().GetString("type")

		//We initialise here, even if we error out later
		sinkTypeEnum, err := ybmclient.NewTelemetryProviderTypeEnumFromValue(strings.ToUpper(sinkType))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		IntegrationSpec, err := setIntegrationConfiguration(cmd, IntegrationName, *sinkTypeEnum)
		if err != nil {
			logrus.Fatalf(err.Error())
		}
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.CreateIntegration().TelemetryProviderSpec(*IntegrationSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The Integration %s has been created", formatter.Colorize(IntegrationName, formatter.GREEN_COLOR))

		fmt.Println(msg)

		IntegrationCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewIntegrationFormat(viper.GetString("output"), string(resp.GetData().Spec.Type)),
		}

		respArr := []ybmclient.TelemetryProviderData{resp.GetData()}

		formatter.IntegrationWrite(IntegrationCtx, respArr)
	},
}

var listIntegrationCmd = &cobra.Command{
	Use:   "list",
	Short: "List Integration",
	Long:  "List Integration",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.ListIntegrations().Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		IntegrationCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewIntegrationFormat(viper.GetString("output"), ""),
		}

		if len(resp.GetData()) < 1 {
			fmt.Println("No Integrations found")
			return
		}

		formatter.IntegrationWrite(IntegrationCtx, resp.GetData())
	},
}

var deleteIntegrationCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Integration",
	Long:  "Delete Integration",
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

		config, err := authApi.GetIntegrationByName(configName)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		r1, err := authApi.DeleteIntegration(config.GetInfo().Id).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r1)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("The Integration %s has been deleted\n", formatter.Colorize(configName, formatter.GREEN_COLOR))
	},
}

func init() {
	IntegrationCmd.AddCommand(createIntegrationCmd)
	createIntegrationCmd.Flags().SortFlags = false
	createIntegrationCmd.Flags().String("config-name", "", "[REQUIRED] The name of the Integration")
	createIntegrationCmd.MarkFlagRequired("config-name")
	createIntegrationCmd.Flags().String("type", "", "[REQUIRED] The type of third party Integration sink")
	createIntegrationCmd.MarkFlagRequired("type")
	createIntegrationCmd.Flags().StringToString("datadog-spec", nil, `Configuration for Datadog. 
	Please provide key value pairs as follows: 
	api-key=<your-datadog-api-key>,site=<your-datadog-site-parameters>`)
	createIntegrationCmd.Flags().StringToString("grafana-spec", nil, `Configuration for Grafana. 
	Please provide key value pairs as follows: 
	access-policy-token=<your-grafana-token>,zone=<your-grafana-zone-parameter>,instance-id=<your-grafana-instance-id>,org-slug=<your-grafana-org-slug>`)
	createIntegrationCmd.Flags().StringToString("sumologic-spec", nil, `Configuration for sumologic. 
	Please provide key value pairs as follows: 
	access-key=<your-sumologic-access-key>,access-id=<your-sumologic-access-id>,installation-token=<your-sumologic-installation-token>`)
	createIntegrationCmd.Flags().StringToString("prometheus-spec", nil, `Configuration for prometheus. 
	Please provide key value pairs as follows: 
	endpoint=<prometheus-otlp-endpoint-url>`)
	createIntegrationCmd.Flags().StringToString("victoriametrics-spec", nil, `Configuration for victoriametrics. 
	Please provide key value pairs as follows: 
	endpoint=<victoriametrics-otlp-endpoint-url>`)

	if util.IsFeatureFlagEnabled(util.GOOGLECLOUD_INTEGRATION) {
		createIntegrationCmd.Flags().String("googlecloud-cred-filepath", "", `Filepath for Google Cloud service account credentials. 
	Please provide absolute file path`)
	}

	IntegrationCmd.AddCommand(listIntegrationCmd)

	IntegrationCmd.AddCommand(deleteIntegrationCmd)
	deleteIntegrationCmd.Flags().String("config-name", "", "[REQUIRED] The name of the Integration")
	deleteIntegrationCmd.MarkFlagRequired("config-name")
	deleteIntegrationCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
}

func setIntegrationConfiguration(cmd *cobra.Command, IntegrationName string, sinkTypeEnum ybmclient.TelemetryProviderTypeEnum) (*ybmclient.TelemetryProviderSpec, error) {
	// We initialize this one here, even if we error out later
	IntegrationSpec := ybmclient.NewTelemetryProviderSpec(IntegrationName, sinkTypeEnum)

	switch sinkTypeEnum {
	case ybmclient.TELEMETRYPROVIDERTYPEENUM_DATADOG:
		if !cmd.Flags().Changed("datadog-spec") {
			return nil, fmt.Errorf("datadog-spec is required for datadog sink")
		}
		datadogSpecString, _ := cmd.Flags().GetStringToString("datadog-spec")
		apiKey := datadogSpecString["api-key"]
		site := datadogSpecString["site"]
		if len(apiKey) < 1 {
			return nil, fmt.Errorf("api-key is a required field for datadog-spec")
		}
		if len(site) < 1 {
			return nil, fmt.Errorf("site is a required field for datadog-spec")
		}
		datadogSpec := ybmclient.NewDatadogTelemetryProviderSpec(apiKey, site)
		IntegrationSpec.SetDatadogSpec(*datadogSpec)
	case ybmclient.TELEMETRYPROVIDERTYPEENUM_PROMETHEUS:
		if !cmd.Flags().Changed("prometheus-spec") {
			return nil, fmt.Errorf("prometheus-spec is required for prometheus sink")
		}
		prometheusSpecs, _ := cmd.Flags().GetStringToString("prometheus-spec")
		endpoint := prometheusSpecs["endpoint"]
		if len(endpoint) < 1 {
			return nil, fmt.Errorf("endpoint is a required field for prometheus-spec")
		}
		prometheusSpec := ybmclient.NewPrometheusTelemetryProviderSpec(endpoint)
		IntegrationSpec.SetPrometheusSpec(*prometheusSpec)
	case ybmclient.TELEMETRYPROVIDERTYPEENUM_VICTORIAMETRICS:
		if !cmd.Flags().Changed("victoriametrics-spec") {
			return nil, fmt.Errorf("victoriametrics-spec is required for victoriametrics sink")
		}
		victoriametricsSpecs, _ := cmd.Flags().GetStringToString("victoriametrics-spec")
		endpoint := victoriametricsSpecs["endpoint"]
		if len(endpoint) < 1 {
			return nil, fmt.Errorf("endpoint is a required field for victoriametrics-spec")
		}
		victoriametricsSpec := ybmclient.NewVictoriaMetricsTelemetryProviderSpec(endpoint)
		IntegrationSpec.SetVictoriametricsSpec(*victoriametricsSpec)
	case ybmclient.TELEMETRYPROVIDERTYPEENUM_GRAFANA:
		if !cmd.Flags().Changed("grafana-spec") {
			return nil, fmt.Errorf("grafana-spec is required for grafana sink")
		}
		grafanaSpecString, _ := cmd.Flags().GetStringToString("grafana-spec")
		apiKey := grafanaSpecString["access-policy-token"]
		zone := grafanaSpecString["zone"]
		instanceId := grafanaSpecString["instance-id"]
		orgSlug := grafanaSpecString["org-slug"]
		if len(apiKey) < 1 {
			return nil, fmt.Errorf("access-policy-token is a required field for grafana-spec")
		}
		if len(zone) < 1 {
			return nil, fmt.Errorf("zone is a required field for grafana-spec")
		}
		if len(instanceId) < 1 {
			return nil, fmt.Errorf("instance-id is a required field for grafana-spec")
		}
		if len(orgSlug) < 1 {
			return nil, fmt.Errorf("org-slug is a required field for grafana-spec")
		}

		grafanaSpec := ybmclient.NewGrafanaTelemetryProviderSpec(apiKey, zone, instanceId, orgSlug)
		IntegrationSpec.SetGrafanaSpec(*grafanaSpec)
	case ybmclient.TELEMETRYPROVIDERTYPEENUM_SUMOLOGIC:
		if !cmd.Flags().Changed("sumologic-spec") {
			return nil, fmt.Errorf("sumologic-spec is required for sumologic sink")
		}
		sumoLogicSpecString, _ := cmd.Flags().GetStringToString("sumologic-spec")
		accessKey := sumoLogicSpecString["access-key"]
		accessId := sumoLogicSpecString["access-id"]
		installationToken := sumoLogicSpecString["installation-token"]
		if len(accessKey) < 1 {
			return nil, fmt.Errorf("access-key is a required field for sumologic-spec")
		}
		if len(accessId) < 1 {
			return nil, fmt.Errorf("access-id is a required field for sumologic-spec")
		}
		if len(installationToken) < 1 {
			return nil, fmt.Errorf("installation-token is a required field for sumologic-spec")
		}
		sumoLogicSpec := ybmclient.NewSumologicTelemetryProviderSpec(installationToken, accessId, accessKey)
		IntegrationSpec.SetSumologicSpec(*sumoLogicSpec)
	case ybmclient.TELEMETRYPROVIDERTYPEENUM_GOOGLECLOUD:
		if !util.IsFeatureFlagEnabled(util.GOOGLECLOUD_INTEGRATION) {
			return nil, fmt.Errorf("Integration of type GOOGLECLOUD is currently not supported")
		}

		if !cmd.Flags().Changed("googlecloud-cred-filepath") {
			return nil, fmt.Errorf("googlecloud-cred-filepath is required for googlecloud sink")
		}
		filepath, _ := cmd.Flags().GetString("googlecloud-cred-filepath")
		jsonFile, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %s", err)
		}
		defer jsonFile.Close()

		// Read the file into a byte array
		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %s", err)
		}

		// Unmarshal the byte array into a map
		var googlecloudMap map[string]string
		if err := json.Unmarshal(byteValue, &googlecloudMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %s", err)
		}

		credType, valid := googlecloudMap["type"]
		if !valid {
			return nil, fmt.Errorf("type is a required field for googlecloud credentials")
		}
		projectId, valid := googlecloudMap["project_id"]
		if !valid {
			return nil, fmt.Errorf("project_id is a required field for googlecloud credentials")
		}
		privateKey, valid := googlecloudMap["private_key"]
		if !valid {
			return nil, fmt.Errorf("private_key is a required field for googlecloud credentials")
		}
		privateKeyId, valid := googlecloudMap["private_key_id"]
		if !valid {
			return nil, fmt.Errorf("private_key_id is a required field for googlecloud credentials")
		}
		clientEmail, valid := googlecloudMap["client_email"]
		if !valid {
			return nil, fmt.Errorf("client_email is a required field for googlecloud credentials")
		}
		clientId, valid := googlecloudMap["client_id"]
		if !valid {
			return nil, fmt.Errorf("client_id is a required field for googlecloud credentials")
		}
		authUri, valid := googlecloudMap["auth_uri"]
		if !valid {
			return nil, fmt.Errorf("auth_uri is a required field for googlecloud credentials")
		}
		tokenUri, valid := googlecloudMap["token_uri"]
		if !valid {
			return nil, fmt.Errorf("token_uri is a required field for googlecloud credentials")
		}
		authProviderX509CertUrl, valid := googlecloudMap["auth_provider_x509_cert_url"]
		if !valid {
			return nil, fmt.Errorf("auth_provider_x509_cert_url is a required field for googlecloud credentials")
		}
		clientX509CertUrl, valid := googlecloudMap["client_x509_cert_url"]
		if !valid {
			return nil, fmt.Errorf("client_x509_cert_url is a required field for googlecloud credentials")
		}

		googlecloudSpec := ybmclient.NewGCPServiceAccount(credType, projectId, privateKey, privateKeyId, clientEmail, clientId, authUri, tokenUri, authProviderX509CertUrl, clientX509CertUrl)
		IntegrationSpec.SetGooglecloudSpec(*googlecloudSpec)
	default:
		return nil, fmt.Errorf("only datadog, grafana, googlecloud, prometheus, victoriametrics or sumologic are accepted as third party sink for now")
	}
	return IntegrationSpec, nil
}
