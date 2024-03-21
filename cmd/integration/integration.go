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

		IntegrationId := resp.GetData().Info.Id

		msg := fmt.Sprintf("The Integration %s is being created", formatter.Colorize(IntegrationId, formatter.GREEN_COLOR))

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

		config, err := authApi.GetConfigByName(configName)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		r1, err := authApi.DeleteIntegration(config.GetInfo().Id).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r1)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Deleting Integration %s\n", formatter.Colorize(configName, formatter.GREEN_COLOR))
	},
}

var updateIntegrationCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Integration",
	Long:  "Update Integration",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		IntegrationName, _ := cmd.Flags().GetString("config-name")
		sinkType, _ := cmd.Flags().GetString("type")

		sinkTypeEnum, err := ybmclient.NewTelemetryProviderTypeEnumFromValue(strings.ToUpper(sinkType))
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		oldname := IntegrationName
		if cmd.Flags().Changed("new-config-name") {
			IntegrationName, _ = cmd.Flags().GetString("new-config-name")
		}
		//We initialise this one here, even if we error out later
		IntegrationSpec, err := setIntegrationConfiguration(cmd, IntegrationName, *sinkTypeEnum)
		if err != nil {
			logrus.Fatalf(err.Error())
		}

		config, err := authApi.GetConfigByName(oldname)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.UpdateIntegration(config.GetInfo().Id).TelemetryProviderSpec(*IntegrationSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The Integration %s is being updated", formatter.Colorize(config.GetInfo().Id, formatter.GREEN_COLOR))

		fmt.Println(msg)

		IntegrationCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewIntegrationFormat(viper.GetString("output"), string(resp.GetData().Spec.Type)),
		}

		respArr := []ybmclient.TelemetryProviderData{resp.GetData()}

		formatter.IntegrationWrite(IntegrationCtx, respArr)
	},
}

var validateIntegrationCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate Integration",
	Long:  "Validate Integration",
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

		resp, r, err := authApi.ValidateIntegration().TelemetryProviderSpec(*IntegrationSpec).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		isValid := resp.GetData().IsValid

		if isValid {
			fmt.Println("Integration is valid")
		}else{
			fmt.Println("Integration is invalid. Please check configuration parameters")
		}
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

	IntegrationCmd.AddCommand(listIntegrationCmd)

	IntegrationCmd.AddCommand(deleteIntegrationCmd)
	deleteIntegrationCmd.Flags().String("config-name", "", "[REQUIRED] The name of the Integration")
	deleteIntegrationCmd.MarkFlagRequired("config-name")
	deleteIntegrationCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	IntegrationCmd.AddCommand(updateIntegrationCmd)
	updateIntegrationCmd.Flags().SortFlags = false
	updateIntegrationCmd.Flags().String("config-name", "", "[REQUIRED] The name of the Integration")
	updateIntegrationCmd.MarkFlagRequired("config-name")
	updateIntegrationCmd.Flags().String("type", "", "[REQUIRED] The type of third party Integration sink")
	updateIntegrationCmd.MarkFlagRequired("type")
	updateIntegrationCmd.Flags().String("new-config-name", "", "[OPTIONAL] The new name of the Integration")
	updateIntegrationCmd.MarkFlagRequired("new-config-name")
	updateIntegrationCmd.Flags().StringToString("datadog-spec", nil, `Configuration for Datadog. 
	Please provide key value pairs as follows: 
	api-key=<your-datadog-api-key>,site=<your-datadog-site-parameters>`)
	updateIntegrationCmd.Flags().StringToString("grafana-spec", nil, `Configuration for Grafana. 
	Please provide key value pairs as follows: 
	access-policy-token=<your-grafana-token>,zone=<your-grafana-zone-parameter>,instance-id=<your-grafana-instance-id>,org-slug=<your-grafana-org-slug>`)
	updateIntegrationCmd.Flags().StringToString("sumologic-spec", nil, `Configuration for sumologic. 
	Please provide key value pairs as follows: 
	access-key=<your-sumologic-access-key>,access-id=<your-sumologic-access-id>,installation-token=<your-sumologic-installation-token>`)

	IntegrationCmd.AddCommand(validateIntegrationCmd)
	validateIntegrationCmd.Flags().SortFlags = false
	validateIntegrationCmd.Flags().String("config-name", "", "[REQUIRED] The name of the Integration")
	validateIntegrationCmd.MarkFlagRequired("config-name")
	validateIntegrationCmd.Flags().String("type", "", "[REQUIRED] The type of third party Integration sink")
	validateIntegrationCmd.MarkFlagRequired("type")
	validateIntegrationCmd.Flags().StringToString("datadog-spec", nil, `Configuration for Datadog. 
	Please provide key value pairs as follows: 
	api-key=<your-datadog-api-key>,site=<your-datadog-site-parameters>`)
	validateIntegrationCmd.Flags().StringToString("grafana-spec", nil, `Configuration for Grafana. 
	Please provide key value pairs as follows: 
	access-policy-token=<your-grafana-token>,zone=<your-grafana-zone-parameter>,instance-id=<your-grafana-instance-id>,org-slug=<your-grafana-org-slug>`)
	validateIntegrationCmd.Flags().StringToString("sumologic-spec", nil, `Configuration for sumologic. 
	Please provide key value pairs as follows: 
	access-key=<your-sumologic-access-key>,access-id=<your-sumologic-access-id>,installation-token=<your-sumologic-installation-token>`)
}

func setIntegrationConfiguration(cmd *cobra.Command, IntegrationName string, sinkTypeEnum ybmclient.TelemetryProviderTypeEnum) (*ybmclient.TelemetryProviderSpec, error) {
	// We initialise this one here, even if we error out later
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
	default:
		return nil, fmt.Errorf("only datadog, grafana or sumologic are accepted as third party sink for now")
	}
	return IntegrationSpec, nil
}
