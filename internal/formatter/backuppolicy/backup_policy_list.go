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

package backuppolicy

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	"github.com/yugabyte/ybm-cli/internal/formatter/util"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultBackupPolicyListing = "table {{.TimeInterval}}\t{{.DaysOfTheWeek}}\t{{.BackupStartTime}}\t{{.RetentionPeriod}}\t{{.State}}"
	timeIntervalHeader         = "Time Interval(days)"
	daysOfTheWeekHeader        = "Days of the Week"
	backupStartTimeHeader      = "Backup Start Time"
	retentionPeriodInDays      = "Retention Period(days)"
	state                      = "State"
)

type BackupPolicyContext struct {
	formatter.HeaderContext
	formatter.Context
	c ybmclient.BackupScheduleData
}

func NewBackupPolicyFormat(source string) formatter.Format {
	switch source {
	case "table", "":
		format := defaultBackupPolicyListing
		return formatter.Format(format)
	default: // custom format or json or pretty
		return formatter.Format(source)
	}
}

// BackupPolicyListWrite renders the context for a list of backup policies
func BackupPolicyListWrite(ctx formatter.Context, backupPolicies []ybmclient.BackupScheduleData) error {
	render := func(format func(subContext formatter.SubContext) error) error {
		for _, backupPolicy := range backupPolicies {
			err := format(&BackupPolicyContext{c: backupPolicy})
			if err != nil {
				logrus.Debugf("Error rendering backup policy: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewBackupPolicyContext(), render)
}

// NewBackupPolicyContext creates a new context for rendering backup policy
func NewBackupPolicyContext() *BackupPolicyContext {
	backupPolicyCtx := BackupPolicyContext{}
	backupPolicyCtx.Header = formatter.SubHeaderContext{
		"TimeInterval":    timeIntervalHeader,
		"DaysOfTheWeek":   daysOfTheWeekHeader,
		"BackupStartTime": backupStartTimeHeader,
		"RetentionPeriod": retentionPeriodInDays,
		"State":           state,
	}
	return &backupPolicyCtx
}

func (c *BackupPolicyContext) TimeInterval() string {
	timeInterval := c.c.Spec.GetTimeIntervalInDays()

	if timeInterval != 0 {
		return fmt.Sprintf("%d", timeInterval)
	}
	return "NA"
}

func (c *BackupPolicyContext) State() string {
	state := c.c.Spec.GetState()
	return fmt.Sprintf("%v", state)
}

func (c *BackupPolicyContext) RetentionPeriod() string {
	retentionPeriodInDays := int32(c.c.Info.GetTaskParams()["retention_period_in_days"].(float64))
	return fmt.Sprintf("%d", retentionPeriodInDays)
}

func (c *BackupPolicyContext) DaysOfTheWeek() string {
	if cronExpression := c.c.Spec.GetCronExpression(); cronExpression != "" {
		return util.GetDaysOfTheWeek(cronExpression)
	}
	return "NA"
}

func (c *BackupPolicyContext) BackupStartTime() string {
	if cronExpression := c.c.Spec.GetCronExpression(); cronExpression != "" {
		return util.GetLocalTime(cronExpression)
	}
	return "NA"
}

func (c *BackupPolicyContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
