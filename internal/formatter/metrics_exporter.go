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

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const defaultMetricsExporterListing = "table {{.Name}}\t{{.Type}}\t{{.Site}}\t{{.ApiKey}}\t{{.InstanceId}}\t{{.OrgSlug}}"

type MetricsExporterContext struct {
	HeaderContext
	Context
	me ybmclient.MetricsExporterConfigurationData
}

func NewMetricsExporterFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultMetricsExporterListing
		return Format(format)
	default:
		return Format(source)
	}
}

func NewMetricsExporterContext() *MetricsExporterContext {
	metricsExporterCtx := MetricsExporterContext{}
	metricsExporterCtx.Header = SubHeaderContext{
		"Name":       nameHeader,
		"ID":         "ID",
		"Type":       "Type",
		"Site":       "Site",
		"ApiKey":     "ApiKey",
		"InstanceId": "InstanceId",
		"OrgSlug":    "OrgSlug",
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

func (me *MetricsExporterContext) Site() string {
	if string(me.me.Spec.GetType()) == "DATADOG" {
		return me.me.Spec.GetDatadogSpec().Site
	} else if string(me.me.Spec.GetType()) == "GRAFANA" {
		return me.me.Spec.GetGrafanaSpec().Endpoint
	}
	return ""
}

func (me *MetricsExporterContext) ApiKey() string {
	if string(me.me.Spec.GetType()) == "DATADOG" {
		return me.me.Spec.GetDatadogSpec().ApiKey
	} else if string(me.me.Spec.GetType()) == "GRAFANA" {
		return me.me.Spec.GetGrafanaSpec().ApiKey
	}
	return ""
}

func (me *MetricsExporterContext) InstanceId() string {
	if string(me.me.Spec.GetType()) == "GRAFANA" {
		return me.me.Spec.GetGrafanaSpec().InstanceId
	}
	return ""
}

func (me *MetricsExporterContext) OrgSlug() string {
	if string(me.me.Spec.GetType()) == "GRAFANA" {
		return me.me.Spec.GetGrafanaSpec().OrgSlug
	}
	return ""
}

func (me *MetricsExporterContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(me.me)
}
