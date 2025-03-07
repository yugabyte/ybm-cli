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
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultPitrConfigListing         = "table {{.Namespace}}\t{{.TableType}}\t{{.RetentionPeriodInDays}}\t{{.State}}\t{{.EarliestRecoveryTimeMillis}}\t{{.LatestRecoveryTimeMillis}}"
	retentionPeriodInDaysHeader      = "Retention Period in Days"
	earliestRecoveryTimeMillisHeader = "Earliest Recovery Time in Millis"
	latestRecoveryTimeMillisHeader   = "Latest Recovery Time in Millis"
)

type PitrConfigContext struct {
	HeaderContext
	Context
	d ybmclient.DatabasePitrConfigData
}

func NewPitrConfigFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultPitrConfigListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func SinglePitrConfigWrite(ctx Context, pitrConfig ybmclient.DatabasePitrConfigData) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&PitrConfigContext{d: pitrConfig})
		if err != nil {
			logrus.Debugf("Error rendering PITR Config: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewPitrConfigContext(), render)
}

func PitrConfigWrite(ctx Context, pitrConfig []ybmclient.DatabasePitrConfigData) error {
	//Sort by database name
	sort.Slice(pitrConfig, func(i, j int) bool {
		return string(pitrConfig[i].Info.GetDatabaseName()) < string(pitrConfig[j].Info.GetDatabaseName())
	})

	render := func(format func(subContext SubContext) error) error {
		for _, pitrConfig := range pitrConfig {
			err := format(&PitrConfigContext{d: pitrConfig})
			if err != nil {
				logrus.Debugf("Error rendering PITR Config: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewPitrConfigContext(), render)
}

func NewPitrConfigContext() *PitrConfigContext {
	pitrConfigCtx := PitrConfigContext{}
	pitrConfigCtx.Header = SubHeaderContext{
		"Namespace":                  namespaceHeader,
		"TableType":                  tableTypeHeader,
		"RetentionPeriodInDays":      retentionPeriodInDaysHeader,
		"State":                      stateHeader,
		"EarliestRecoveryTimeMillis": earliestRecoveryTimeMillisHeader,
		"LatestRecoveryTimeMillis":   latestRecoveryTimeMillisHeader,
	}
	return &pitrConfigCtx
}

func (d *PitrConfigContext) Namespace() string {
	return d.d.Info.GetDatabaseName()
}

func (d *PitrConfigContext) TableType() string {
	return string(d.d.Info.GetDatabaseType())
}

func (d *PitrConfigContext) RetentionPeriodInDays() int32 {
	return d.d.Spec.RetentionPeriod
}

func (d *PitrConfigContext) State() string {
	return string(d.d.Info.GetState())
}

func (d *PitrConfigContext) EarliestRecoveryTimeMillis() string {
	return strconv.Itoa(int(d.d.Info.GetEarliestRecoveryTimeMillis()))
}

func (d *PitrConfigContext) LatestRecoveryTimeMillis() string {
	return strconv.Itoa(int(d.d.Info.GetLatestRecoveryTimeMillis()))
}

func (d *PitrConfigContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.d)
}
