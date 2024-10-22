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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const defaultDbQueryLoggingConfigListing = "table {{.State}}\t{{.IntegrationName}}"
const loggingConfigListing = "table {{.LogConfigKey}}\t{{.LogConfigValue}}"

type DbQueryLoggingContext struct {
	HeaderContext
	Context
	data            ybmclient.PgLogExporterConfigData
	integrationName string
}

type LogConfigContext struct {
	HeaderContext
	Context
	configKey string
	configVal string
}

func NewDbQueryLoggingFormat() Format {
	source := viper.GetString("output")
	switch source {
	case "table", "":
		format := defaultDbQueryLoggingConfigListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewLogConfigFormat() Format {
	source := viper.GetString("output")
	switch source {
	case "table", "":
		format := loggingConfigListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewDbQueryLoggingContext() *DbQueryLoggingContext {
	DbQueryLoggingContext := DbQueryLoggingContext{}
	DbQueryLoggingContext.Header = SubHeaderContext{
		"State":           "State",
		"IntegrationName": "Integration Name",
	}
	return &DbQueryLoggingContext
}

func NewLogConfigContext() *LogConfigContext {
	LogConfigContext := LogConfigContext{}
	LogConfigContext.Header = SubHeaderContext{
		"LogConfigKey":   "Log Config Key",
		"LogConfigValue": "Log Config Value",
	}
	return &LogConfigContext
}

func dbLogConfigWrite(ctx Context, PgLogExportConfig ybmclient.PgLogExportConfig) error {
	render := func(format func(subContext SubContext) error) error {
		addRow(format, "debug-print-plan", fmt.Sprintf("%t", PgLogExportConfig.DebugPrintPlan))
		addRow(format, "log-min-duration-statement", fmt.Sprintf("%d", PgLogExportConfig.LogMinDurationStatement))
		addRow(format, "log-connections", fmt.Sprintf("%t", PgLogExportConfig.LogConnections))
		addRow(format, "log-disconnections", fmt.Sprintf("%t", PgLogExportConfig.LogDisconnections))
		addRow(format, "log-duration", fmt.Sprintf("%t", PgLogExportConfig.LogDuration))
		addRow(format, "log-error-verbosity", string(PgLogExportConfig.LogErrorVerbosity))
		addRow(format, "log-statement", string(PgLogExportConfig.LogStatement))
		addRow(format, "log-min-error-statement", string(PgLogExportConfig.LogMinErrorStatement))
		addRow(format, "log-line-prefix", PgLogExportConfig.LogLinePrefix)
		return nil
	}
	return ctx.Write(NewLogConfigContext(), render)
}

func addRow(format func(subContext SubContext) error, key string, value string) {
	err := format(&LogConfigContext{configKey: key, configVal: value})
	if err != nil {
		logrus.Fatal(err)
	}
}

func dbQueryLoggingWrite(ctx Context, PgLogExporterConfigData ybmclient.PgLogExporterConfigData, integrationName string) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&DbQueryLoggingContext{data: PgLogExporterConfigData, integrationName: integrationName})
		if err != nil {
			logrus.Debugf("Error rendering Pg Log Exporter config data: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewDbQueryLoggingContext(), render)
}

func DbQueryLoggingWriteFull(PgLogExporterConfigData ybmclient.PgLogExporterConfigData, integrationName string) {
	ctx := Context{
		Output: os.Stdout,
		Format: NewDbQueryLoggingFormat(),
	}

	err := dbQueryLoggingWrite(ctx, PgLogExporterConfigData, integrationName)
	if err != nil {
		logrus.Fatal(err.Error())
	}
	ctx.Output.Write([]byte("\n"))

	// Only render Log config for table output format
	if viper.GetString("output") == "table" {
		ctx = Context{
			Output: os.Stdout,
			Format: NewLogConfigFormat(),
		}

		err = dbLogConfigWrite(ctx, PgLogExporterConfigData.Spec.ExportConfig)
		if err != nil {
			logrus.Fatal(err.Error())
		}
	}

}

func (context *DbQueryLoggingContext) State() string {
	return string(context.data.Info.State)
}

func (context *DbQueryLoggingContext) IntegrationName() string {
	return context.integrationName
}

func (context *DbQueryLoggingContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(context.data)
}

func (context *LogConfigContext) LogConfigKey() string {
	return string(context.configKey)
}

func (context *LogConfigContext) LogConfigValue() string {
	return string(context.configVal)
}
