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
	"strings"
	"time"

	cron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultBackupPolicyListing    = "table {{.TimeInterval}}\t{{.IncrementalTimeInterval}}\t{{.DaysOfTheWeek}}\t{{.BackupStartTime}}\t{{.RetentionPeriod}}\t{{.State}}"
	timeIntervalHeader            = "Time Interval(days)"
	incrementalTimeIntervalHeader = "Incremental Time Interval(minutes)"
	daysOfTheWeekHeader           = "Days of the Week"
	backupStartTimeHeader         = "Backup Start Time"
	retentionPeriodInDays         = "Retention Period(days)"
	state                         = "State"
)

type BackupPolicyContext struct {
	HeaderContext
	Context
	c ybmclient.BackupScheduleDataV2
}

func NewBackupPolicyFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultBackupPolicyListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// BackupPolicyListWrite renders the context for a list of backup policies
func BackupPolicyListWrite(ctx Context, backupPolicies []ybmclient.BackupScheduleDataV2) error {
	render := func(format func(subContext SubContext) error) error {
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
	backupPolicyCtx.Header = SubHeaderContext{
		"TimeInterval":            timeIntervalHeader,
		"IncrementalTimeInterval": incrementalTimeIntervalHeader,
		"DaysOfTheWeek":           daysOfTheWeekHeader,
		"BackupStartTime":         backupStartTimeHeader,
		"RetentionPeriod":         retentionPeriodInDays,
		"State":                   state,
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

func (c *BackupPolicyContext) IncrementalTimeInterval() string {
	incrementalTimeInterval := c.c.Spec.GetIncrementalIntervalInMinutes()
	if incrementalTimeInterval != 0 {
		return fmt.Sprintf("%d", incrementalTimeInterval)
	}
	return "NA"
}

func (c *BackupPolicyContext) State() string {
	state := c.c.Spec.GetState()
	return fmt.Sprintf("%v", state)
}

func (c *BackupPolicyContext) RetentionPeriod() string {
	retentionPeriodInDays := int32(c.c.Spec.GetRetentionPeriodInDays())
	return fmt.Sprintf("%d", retentionPeriodInDays)
}

func (c *BackupPolicyContext) DaysOfTheWeek() string {
	if cronExpression := c.c.Spec.GetCronExpression(); cronExpression != "" {
		return getDaysOfTheWeek(cronExpression)
	}
	return "NA"
}

func (c *BackupPolicyContext) BackupStartTime() string {
	if cronExpression := c.c.Spec.GetCronExpression(); cronExpression != "" {
		return getLocalTime(cronExpression)
	}
	return "NA"
}

func (c *BackupPolicyContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}

func getLocalTime(cronExpression string) string {
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := cronParser.Parse(cronExpression)
	if err != nil {
		logrus.Debugln("Error parsing cron expression:\n", err)
		return ""
	}

	// Get the next scheduled time in UTC
	utcTime := schedule.Next(time.Now().UTC())
	localTimeZone, err := time.LoadLocation("Local")
	if err != nil {
		logrus.Debugln("Error loading local time zone:\n", err)
		return utcTime.Format("15:04") + "UTC"
	}
	localTime := utcTime.In(localTimeZone)
	localTimeString := localTime.Format("15:04")
	return localTimeString
}

func getDaysOfTheWeek(cronExpression string) string {

	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := cronParser.Parse(cronExpression)
	specSchedule := schedule.(*cron.SpecSchedule)

	if err != nil {
		logrus.Debugln("Error parsing cron expression:\n", err)
		return ""
	}

	indexToDayMap := map[int]string{
		0: "Su",
		1: "Mo",
		2: "Tu",
		3: "We",
		4: "Th",
		5: "Fr",
		6: "Sa",
	}
	daysOfTheWeek := []string{}
	for i := 0; i < 7; i++ {
		dowMatch := 1<<uint(i)&specSchedule.Dow > 0
		day := indexToDayMap[i]
		if dowMatch {
			daysOfTheWeek = append(daysOfTheWeek, day)
		}
	}

	return strings.Join(daysOfTheWeek, ",")

}
