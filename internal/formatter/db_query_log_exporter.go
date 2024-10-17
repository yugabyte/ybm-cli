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

const defaultDbQueryLoggingConfigListing = "table {{.State}}\t{{.IntegrationID}}\t{{.LogConfig}}"

type DbQueryLoggingContext struct {
	HeaderContext
	Context
	data ybmclient.PgLogExporterConfigData
}

func NewDbQueryLoggingFormat() Format {
	return Format(defaultDbQueryLoggingConfigListing)
}

func NewDbQueryLoggingContext() *DbQueryLoggingContext {
	DbQueryLoggingContext := DbQueryLoggingContext{}
	DbQueryLoggingContext.Header = SubHeaderContext{
		"State":         "State",
		"IntegrationID": "Integration ID",
		"LogConfig":     "Log Config",
	}
	return &DbQueryLoggingContext
}

func DbQueryLoggingWrite(ctx Context, PgLogExporterConfigData []ybmclient.PgLogExporterConfigData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, data := range PgLogExporterConfigData {
			err := format(&DbQueryLoggingContext{data: data})
			if err != nil {
				logrus.Debugf("Error rendering Pg Log Exporter config data: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewDbQueryLoggingContext(), render)
}

func (context *DbQueryLoggingContext) State() string {
	return string(context.data.Info.State)
}

func (context *DbQueryLoggingContext) IntegrationID() string {
	return string(context.data.Spec.ExporterId)
}

func (context *DbQueryLoggingContext) LogConfig() string {
	x := context.data.GetSpec()
	y := x.GetExportConfig()

	return convertPgLogConfigToJson(&y)
}

func convertPgLogConfigToJson(cfg *ybmclient.PgLogExportConfig) string {
	jsonConfig, _ := json.Marshal(cfg)
	return string(jsonConfig)
}
