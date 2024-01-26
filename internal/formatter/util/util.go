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

package util

import (
	"strings"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

func GetLocalTime(cronExpression string) string {
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

func GetDaysOfTheWeek(cronExpression string) string {

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
