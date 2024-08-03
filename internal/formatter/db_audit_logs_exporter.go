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

const (
	defaultDbAuditLogsExporterListing = "table {{.ID}}\t{{.CreatedAt}}\t{{.ClusterId}}\t{{.IntegrationId}}\t{{.State}}\t{{.YsqlConfig}}"
	integrationIdHeader               = "Integration ID"
	ysqlConfigHeader                  = "Ysql Config"
)

type DbAuditLogsExporterContext struct {
	HeaderContext
	Context
	a ybmclient.DbAuditExporterConfigurationData
}

func NewDbAuditLogsExporterFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultDbAuditLogsExporterListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// DbAuditWrite renders the context for a list of Db Audit Export config
func DbAuditLogsExporterWrite(ctx Context, dbAuditLogsExporterConfigs []ybmclient.DbAuditExporterConfigurationData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, dbAuditLogsExporterConfig := range dbAuditLogsExporterConfigs {
			err := format(&DbAuditLogsExporterContext{a: dbAuditLogsExporterConfig})
			if err != nil {
				logrus.Debugf("Error rendering DB Audit Logs Exporter Config: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewDbAuditLogsExporterContext(), render)
}

// NewDbAuditContext creates a new context for rendering Db Audit Export Config
func NewDbAuditLogsExporterContext() *DbAuditLogsExporterContext {
	dbAuditLogsExporterCtx := DbAuditLogsExporterContext{}
	dbAuditLogsExporterCtx.Header = SubHeaderContext{
		"YsqlConfig":    ysqlConfigHeader,
		"ID":            "ID",
		"State":         stateHeader,
		"IntegrationId": integrationIdHeader,
		"ClusterId":     clusterIdHeader,
		"CreatedAt":     dateCreatedAtHeader,
	}
	return &dbAuditLogsExporterCtx
}

func (a *DbAuditLogsExporterContext) ID() string {
	return a.a.Info.Id
}

func (a *DbAuditLogsExporterContext) State() string {
	return string(a.a.Info.State)
}

func (a *DbAuditLogsExporterContext) IntegrationId() string {
	return a.a.Spec.ExporterId
}

func (a *DbAuditLogsExporterContext) ClusterId() string {
	return a.a.Info.ClusterId
}

func (a *DbAuditLogsExporterContext) YsqlConfig() string {
	ysqlConfig, _ := json.Marshal(a.a.Spec.YsqlConfig)
	return string(ysqlConfig)
}

func (a *DbAuditLogsExporterContext) CreatedAt() string {
	return a.a.Info.Metadata.Get().GetCreatedOn()
}

func (a *DbAuditLogsExporterContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.a)
}
