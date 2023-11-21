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

const defaultMetricsExporterListing = "table {{.Name}}\t{{.Type}}"
const defaultMetricsExporterDataDog = "table {{.Name}}\t{{.Type}}\t{{.Site}}\t{{.ApiKey}}"
const defaultMetricsExporterGrafana = "table {{.Name}}\t{{.Type}}\t{{.Zone}}\t{{.AccessTokenPolicy}}\t{{.InstanceId}}\t{{.OrgSlug}}"
const defaultMetricsExporterSumologic = "table {{.Name}}\t{{.Type}}\t{{.AccessKey}}\t{{.AccessID}}\t{{.InstallationToken}}"

type MetricsExporterContext struct {
	HeaderContext
	Context
	me ybmclient.MetricsExporterConfigurationData
}

func NewMetricsExporterFormat(source string, metricsType string) Format {
	format := defaultMetricsExporterListing

	//Display will differ by exporter type for describe
	switch metricsType {
	case string(ybmclient.METRICSEXPORTERCONFIGTYPEENUM_DATADOG):
		format = defaultMetricsExporterDataDog
	case string(ybmclient.METRICSEXPORTERCONFIGTYPEENUM_GRAFANA):
		format = defaultMetricsExporterGrafana
	case string(ybmclient.METRICSEXPORTERCONFIGTYPEENUM_SUMOLOGIC):
		format = defaultMetricsExporterSumologic
	}

	switch source {
	case "table", "":
		return Format(format)
	default:
		return Format(source)
	}
}

func NewMetricsExporterContext() *MetricsExporterContext {
	metricsExporterCtx := MetricsExporterContext{}
	metricsExporterCtx.Header = SubHeaderContext{
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
	}
	return &metricsExporterCtx
}

func MetricsExporterWrite(ctx Context, metricsExporters []ybmclient.MetricsExporterConfigurationData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, metricsExporter := range metricsExporters {
			err := format(&MetricsExporterContext{me: metricsExporter})
			if err != nil {
				logrus.Debugf("Error rendering metrics exporter: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewMetricsExporterContext(), render)
}

func (me *MetricsExporterContext) ID() string {
	return me.me.Info.Id
}

func (me *MetricsExporterContext) Name() string {
	return me.me.Spec.Name
}

func (me *MetricsExporterContext) Type() string {
	return string(me.me.Spec.Type)
}

func (me *MetricsExporterContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(me.me)
}

// DATADOG
func (me *MetricsExporterContext) ApiKey() string {
	return me.me.Spec.GetDatadogSpec().ApiKey
}

func (me *MetricsExporterContext) Site() string {
	return me.me.Spec.GetDatadogSpec().Site
}

// GRAFANA
func (me *MetricsExporterContext) Zone() string {
	return me.me.Spec.GetGrafanaSpec().Zone
}

func (me *MetricsExporterContext) AccessTokenPolicy() string {
	return ShortenKey(me.me.Spec.GetGrafanaSpec().AccessPolicyToken, 32)
}

func (me *MetricsExporterContext) InstanceId() string {
	return me.me.Spec.GetGrafanaSpec().InstanceId
}

func (me *MetricsExporterContext) OrgSlug() string {
	return me.me.Spec.GetGrafanaSpec().OrgSlug
}

// Sumologic
func (me *MetricsExporterContext) AccessID() string {
	return me.me.Spec.GetSumologicSpec().AccessId
}

func (me *MetricsExporterContext) AccessKey() string {
	return ShortenKey(me.me.Spec.GetSumologicSpec().AccessKey, 32)
}

func (me *MetricsExporterContext) InstallationToken() string {
	return ShortenKey(me.me.Spec.GetSumologicSpec().InstallationToken, 32)
}

func ShortenKey(key string, stringLen int) string {
	if len(key) > stringLen {
		return fmt.Sprintf("%s%s%s", key[:12], "...", key[len(key)-17:])
	}
	return key
}
