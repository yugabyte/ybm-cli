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
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Map weekdays to cron format
var dayMapping = map[string]string{
	"su": "0",
	"mo": "1",
	"tu": "2",
	"we": "3",
	"th": "4",
	"fr": "5",
	"sa": "6",
}

func IsTimeFormatValid(timeStr string) bool {

	// check if the time string is in "HH:MM" format
	timeRegex := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
	return timeRegex.MatchString(timeStr)
}

func IsDaysOfWeekValid(daysOfWeek string) bool {
	daysOfWeek = strings.TrimSpace(daysOfWeek)
	daysOfWeekList := strings.Split(daysOfWeek, ",")
	if len(daysOfWeekList) == 0 {
		return false
	}
	for _, day := range daysOfWeekList {
		day = strings.ToLower(day)
		_, found := dayMapping[day]
		if !found {
			return false
		}
	}
	return true
}

func ConvertLocalTimeToUTC(localTimeStr string) string {

	backupTimeList := strings.Split(localTimeStr, ":")
	localHour, _ := strconv.Atoi(backupTimeList[0])
	localMinute, _ := strconv.Atoi(backupTimeList[1])

	currentTime := time.Now()
	localTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), localHour, localMinute, 0, 0, time.Local)
	utcTime := localTime.UTC()

	utcTimeStr := utcTime.Format("15:04")

	return utcTimeStr
}

func GenerateCronExpression(daysOfWeek string, backupTime string) string {

	daysOfWeekList := strings.Split(daysOfWeek, ",")

	backupTimeList := strings.Split(backupTime, ":")
	hour, _ := strconv.Atoi(backupTimeList[0])
	minute, _ := strconv.Atoi(backupTimeList[1])

	var cronDays []string
	for _, day := range daysOfWeekList {
		day = strings.ToLower(day)
		cronDay, found := dayMapping[day]
		if found {
			cronDays = append(cronDays, cronDay)
		}

	}

	cronExpr := fmt.Sprintf("%d %d * * %s", minute, hour, strings.Join(cronDays, ","))

	return cronExpr
}
