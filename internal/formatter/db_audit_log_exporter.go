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
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const defaultDbAuditLoggingConfigListing = "table {{.State}}\t{{.IntegrationName}}"
const ysqlConfigListing = "table {{.YsqlConfigKey}}\t{{.YsqlConfigValue}}"

type DbAuditLoggingContext struct {
	HeaderContext
	Context
	data            ybmclient.DbAuditExporterConfigurationData
	integrationName string
}

type YsqlConfigListing struct {
	HeaderContext
	Context
	configKey string
	configVal string
}

func NewDbAuditLoggingFormat() Format {
	source := viper.GetString("output")
	switch source {
	case "table", "":
		format := defaultDbAuditLoggingConfigListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewYsqlConfigFormat() Format {
	source := viper.GetString("output")
	switch source {
	case "table", "":
		format := ysqlConfigListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewDbAuditLoggingContext() *DbAuditLoggingContext {
	DbAuditLoggingContext := DbAuditLoggingContext{}
	DbAuditLoggingContext.Header = SubHeaderContext{
		"State":           "State",
		"IntegrationName": "Integration Name",
	}
	return &DbAuditLoggingContext
}

func NewYsqlConfigContext() *YsqlConfigListing {
	YsqlConfigListing := YsqlConfigListing{}
	YsqlConfigListing.Header = SubHeaderContext{
		"YsqlConfigKey":   "Ysql Config Key",
		"YsqlConfigValue": "Ysql Config Value",
	}
	return &YsqlConfigListing
}

func ysqlConfigWrite(ctx Context, ysqlConfig ybmclient.DbAuditYsqlExportConfig) error {
	logSettings := ysqlConfig.GetLogSettings()
	render := func(format func(subContext SubContext) error) error {
		addYsqlConfigRow(format, "statement-classes", convertEnumListToString(ysqlConfig.GetStatementClasses()))
		addYsqlConfigRow(format, "log-catalog", strconv.FormatBool(*logSettings.LogCatalog))
		addYsqlConfigRow(format, "log-client", strconv.FormatBool(*logSettings.LogClient))
		addYsqlConfigRow(format, "log-level", string(*logSettings.LogLevel))
		addYsqlConfigRow(format, "log-parameter", strconv.FormatBool(*logSettings.LogParameter))
		addYsqlConfigRow(format, "log-relation", strconv.FormatBool(*logSettings.LogRelation))
		addYsqlConfigRow(format, "log-statement-once", strconv.FormatBool(*logSettings.LogStatementOnce))
		return nil
	}
	return ctx.Write(NewYsqlConfigContext(), render)
}

func addYsqlConfigRow(format func(subContext SubContext) error, key string, value string) {
	err := format(&YsqlConfigListing{configKey: key, configVal: value})
	if err != nil {
		logrus.Fatal(err)
	}
}

func dbAuditLoggingWrite(ctx Context, dbAuditLoggingData ybmclient.DbAuditExporterConfigurationData, integrationName string) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&DbAuditLoggingContext{data: dbAuditLoggingData, integrationName: integrationName})
		if err != nil {
			logrus.Debugf("Error rendering DB Audit Logging configuration data: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewDbAuditLoggingContext(), render)
}

func DbAuditLoggingWriteFull(dbAuditLoggingData ybmclient.DbAuditExporterConfigurationData, integrationName string) {
	ctx := Context{
		Output: os.Stdout,
		Format: NewDbAuditLoggingFormat(),
	}

	err := dbAuditLoggingWrite(ctx, dbAuditLoggingData, integrationName)
	if err != nil {
		logrus.Fatal(err.Error())
	}
	ctx.Output.Write([]byte("\n"))

	// Only render Log config for table output format
	if viper.GetString("output") == "table" {
		ctx = Context{
			Output: os.Stdout,
			Format: NewYsqlConfigFormat(),
		}

		err = ysqlConfigWrite(ctx, dbAuditLoggingData.Spec.YsqlConfig)
		if err != nil {
			logrus.Fatal(err.Error())
		}
	}

}

func (context *DbAuditLoggingContext) State() string {
	return string(context.data.Info.State)
}

func (context *DbAuditLoggingContext) IntegrationName() string {
	return context.integrationName
}

func (context *DbAuditLoggingContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(context.data)
}

func (context *YsqlConfigListing) YsqlConfigKey() string {
	return string(context.configKey)
}

func (context *YsqlConfigListing) YsqlConfigValue() string {
	return string(context.configVal)
}

func convertEnumListToString(list []ybmclient.DbAuditYsqlStatmentClassesEnum) string {
	var strList []string
	for _, enumValue := range list {
		strList = append(strList, string(enumValue))
	}
	return "[" + strings.Join(strList, ", ") + "]"
}
