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

package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const defaultIntegrationListing = "table {{.ID}}\t{{.Name}}\t{{.Type}}"
const defaultIntegrationDataDog = "table {{.ID}}\t{{.Name}}\t{{.Type}}\t{{.Site}}\t{{.ApiKey}}"
const defaultIntegrationPrometheus = "table {{.ID}}\t{{.Name}}\t{{.Type}}\t{{.Endpoint}}"
const defaultIntegrationGrafana = "table {{.ID}}\t{{.Name}}\t{{.Type}}\t{{.Zone}}\t{{.AccessTokenPolicy}}\t{{.InstanceId}}\t{{.OrgSlug}}"
const defaultIntegrationSumologic = "table {{.ID}}\t{{.Name}}\t{{.Type}}\t{{.AccessKey}}\t{{.AccessID}}\t{{.InstallationToken}}"

type IntegrationContext struct {
	HeaderContext
	Context
	tp ybmclient.TelemetryProviderData
}

func NewIntegrationFormat(source string, providerType string) Format {
	format := defaultIntegrationListing

	//Display will differ by exporter type for describe
	switch providerType {
	case string(ybmclient.TELEMETRYPROVIDERTYPEENUM_DATADOG):
		format = defaultIntegrationDataDog
	case string(ybmclient.TELEMETRYPROVIDERTYPEENUM_PROMETHEUS):
		format = defaultIntegrationPrometheus
	case string(ybmclient.TELEMETRYPROVIDERTYPEENUM_GRAFANA):
		format = defaultIntegrationGrafana
	case string(ybmclient.TELEMETRYPROVIDERTYPEENUM_SUMOLOGIC):
		format = defaultIntegrationSumologic
	}

	switch source {
	case "table", "":
		return Format(format)
	default:
		return Format(source)
	}
}

func NewIntegrationContext() *IntegrationContext {
	IntegrationCtx := IntegrationContext{}
	IntegrationCtx.Header = SubHeaderContext{
		"Name":              nameHeader,
		"ID":                "ID",
		"Type":              "Type",
		"Site":              "Site",
		"ApiKey":            "ApiKey",
		"InstanceId":        "InstanceId",
		"OrgSlug":           "OrgSlug",
		"AccessTokenPolicy": "Access Token Policy",
		"Zone":              "Zone",
		"InstallationToken": "InstallationToken",
		"AccessID":          "Access ID",
		"AccessKey":         "Access Key",
		"Endpoint":          "Endpoint",
	}
	return &IntegrationCtx
}

func IntegrationWrite(ctx Context, Integrations []ybmclient.TelemetryProviderData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, Integration := range Integrations {
			err := format(&IntegrationContext{tp: Integration})
			if err != nil {
				logrus.Debugf("Error rendering Integration: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewIntegrationContext(), render)
}

func (tp *IntegrationContext) ID() string {
	return tp.tp.Info.Id
}

func (tp *IntegrationContext) Name() string {
	return tp.tp.Spec.Name
}

func (tp *IntegrationContext) Type() string {
	return string(tp.tp.Spec.Type)
}

func (tp *IntegrationContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(tp.tp)
}

// DATADOG
func (tp *IntegrationContext) ApiKey() string {
	return tp.tp.Spec.GetDatadogSpec().ApiKey
}

func (tp *IntegrationContext) Site() string {
	return tp.tp.Spec.GetDatadogSpec().Site
}

// GRAFANA
func (tp *IntegrationContext) Zone() string {
	return tp.tp.Spec.GetGrafanaSpec().Zone
}

func (tp *IntegrationContext) AccessTokenPolicy() string {
	return IntegrationShortenKey(tp.tp.Spec.GetGrafanaSpec().AccessPolicyToken, 32)
}

func (tp *IntegrationContext) InstanceId() string {
	return tp.tp.Spec.GetGrafanaSpec().InstanceId
}

func (tp *IntegrationContext) OrgSlug() string {
	return tp.tp.Spec.GetGrafanaSpec().OrgSlug
}

// Sumologic
func (tp *IntegrationContext) AccessID() string {
	return tp.tp.Spec.GetSumologicSpec().AccessId
}

func (tp *IntegrationContext) AccessKey() string {
	return IntegrationShortenKey(tp.tp.Spec.GetSumologicSpec().AccessKey, 32)
}

func (tp *IntegrationContext) InstallationToken() string {
	return IntegrationShortenKey(tp.tp.Spec.GetSumologicSpec().InstallationToken, 32)
}

func IntegrationShortenKey(key string, stringLen int) string {
	if len(key) > stringLen {
		return fmt.Sprintf("%s%s%s", key[:12], "...", key[len(key)-17:])
	}
	return key
}

// Prometheus
func (tp *IntegrationContext) Endpoint() string {
	return tp.tp.Spec.GetPrometheusSpec().Endpoint
}
