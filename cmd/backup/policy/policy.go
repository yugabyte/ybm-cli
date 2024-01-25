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

package policy

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
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

var PolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage backup policy of a cluster",
	Long:  "Manage backup policy of a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listPolicyCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup policies",
	Long:  "List backup policies for cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		listBackupPoliciesRequest := authApi.ListBackupPolicies(clusterID, false /* fetchOnlyActive */)

		resp, r, err := listBackupPoliciesRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		policyCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewBackupPolicyFormat(viper.GetString("output")),
		}
		if len(resp.GetData()) < 1 {
			logrus.Fatalln("No backup policies found for the given cluster")
		}
		formatter.BackupPolicyListWrite(policyCtx, resp.GetData())
	},
}

var enablePolicyCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable backup policies",
	Long:  "Enable backup policies for cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		listBackupPoliciesRequest := authApi.ListBackupPolicies(clusterId, false /* fetchOnlyActive */)

		resp, r, err := listBackupPoliciesRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			logrus.Fatalln("No backup policies found for the given cluster")
		}
		backupScheduleSpec := resp.GetData()[0].GetSpec()
		if backupScheduleSpec.GetState() == ybmclient.SCHEDULESTATEENUM_ACTIVE {
			logrus.Fatalf("The backup policy is already enabled for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		}
		backupScheduleSpec.SetState(ybmclient.SCHEDULESTATEENUM_ACTIVE)
		info := resp.GetData()[0].GetInfo()
		scheduleId := info.GetId()

		_, r, err = authApi.UpdateBackupPolicy(clusterId, scheduleId).ScheduleSpecV2(backupScheduleSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Successfully enabled backup policy for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

	},
}

var disablePolicyCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable backup policies",
	Long:  "Disable backup policies for cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		listBackupPoliciesRequest := authApi.ListBackupPolicies(clusterId, true /* fetchOnlyActive */)

		resp, r, err := listBackupPoliciesRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			logrus.Fatalln("No backup policies found for the given cluster")
		}
		backupScheduleSpec := resp.GetData()[0].GetSpec()
		if backupScheduleSpec.GetState() == ybmclient.SCHEDULESTATEENUM_PAUSED {
			logrus.Fatalf("The backup policy is already disabled for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		}
		backupScheduleSpec.SetState(ybmclient.SCHEDULESTATEENUM_PAUSED)
		info := resp.GetData()[0].GetInfo()
		scheduleId := info.GetId()

		_, r, err = authApi.UpdateBackupPolicy(clusterId, scheduleId).ScheduleSpecV2(backupScheduleSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		fmt.Printf("Successfully disabled backup policy for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

	},
}

var updatePolicyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update backup policies",
	Long:  "Update backup policies for cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		retentionPeriodInDays, _ := cmd.Flags().GetInt32("retention-period-in-days")
		if retentionPeriodInDays < 1 {
			logrus.Fatalln("Retention period should be greater than or equal to 1 day")
		}

		listBackupPoliciesRequest := authApi.ListBackupPolicies(clusterId, false /* fetchOnlyActive */)
		resp, r, err := listBackupPoliciesRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) < 1 {
			logrus.Fatalln("No backup policies found for the given cluster")
		}

		info := resp.GetData()[0].GetInfo()
		scheduleId := info.GetId()
		backupScheduleSpec := resp.GetData()[0].GetSpec()

		if cmd.Flags().Changed("full-backup-frequency-in-days") {
			frequencyInDays, _ := cmd.Flags().GetInt32("full-backup-frequency-in-days")
			if frequencyInDays < 1 {
				logrus.Fatalln("Time interval for scheduling backup should be greater than or equal to 1 day")
			}
			backupScheduleSpec.SetTimeIntervalInDays(frequencyInDays)
			backupScheduleSpec.UnsetCronExpression()
		} else {
			daysOfWeek, _ := cmd.Flags().GetString("full-backup-schedule-days-of-week")
			if !isDaysOfWeekValid(daysOfWeek) {
				logrus.Fatalln("The days of week specified is incorrect. Please ensure that it is a comma separated list of the first two letters to days of the week.")
			}
			backupTime, _ := cmd.Flags().GetString("full-backup-schedule-time")
			if !isTimeFormatValid(backupTime) {
				logrus.Fatalln("The full backup schedule time is invalid. Please ensure that it in the 24 Hr HH:MM format.")
			}
			backupTimeUTC := convertLocalTimeToUTC(backupTime)
			cronExpression := generateCronExpression(daysOfWeek, backupTimeUTC)
			backupScheduleSpec.SetCronExpression(cronExpression)
			backupScheduleSpec.TimeIntervalInDays = nil
		}

		if util.IsFeatureFlagEnabled(util.INCREMENTAL_BACKUP) {
			if cmd.Flags().Changed("incremental-backup-frequency-in-minutes") {
				incrementalBackupFrequencyInMinutes, _ := cmd.Flags().GetInt32("incremental-backup-frequency-in-minutes")
				if incrementalBackupFrequencyInMinutes < 1 {
					logrus.Fatalln("Time interval for scheduling incremental backup cannot be negative or zero")
				}
				backupScheduleSpec.SetIncrementalIntervalInMinutes(incrementalBackupFrequencyInMinutes)
			} else {
				backupScheduleSpec.UnsetIncrementalIntervalInMinutes()
			}
		}

		_, r, err = authApi.UpdateBackupPolicy(clusterId, scheduleId).ScheduleSpecV2(backupScheduleSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Successfully updated backup policy for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

	},
}

func isTimeFormatValid(timeStr string) bool {

	// check if the time string is in "HH:MM" format
	timeRegex := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
	return timeRegex.MatchString(timeStr)
}

func isDaysOfWeekValid(daysOfWeek string) bool {
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

func convertLocalTimeToUTC(localTimeStr string) string {

	backupTimeList := strings.Split(localTimeStr, ":")
	localHour, _ := strconv.Atoi(backupTimeList[0])
	localMinute, _ := strconv.Atoi(backupTimeList[1])

	currentTime := time.Now()
	localTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), localHour, localMinute, 0, 0, time.Local)
	utcTime := localTime.UTC()

	utcTimeStr := utcTime.Format("15:04")

	return utcTimeStr
}

func generateCronExpression(daysOfWeek string, backupTime string) string {

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

func init() {
	PolicyCmd.AddCommand(listPolicyCmd)
	PolicyCmd.AddCommand(enablePolicyCmd)
	PolicyCmd.AddCommand(disablePolicyCmd)
	PolicyCmd.AddCommand(updatePolicyCmd)

	listPolicyCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster to list backup policies.")
	listPolicyCmd.MarkFlagRequired("cluster-name")

	enablePolicyCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster to enable backup policies.")
	enablePolicyCmd.MarkFlagRequired("cluster-name")

	disablePolicyCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster to disable backup policies.")
	disablePolicyCmd.MarkFlagRequired("cluster-name")

	updatePolicyCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster to update backup policies.")
	updatePolicyCmd.MarkFlagRequired("cluster-name")
	updatePolicyCmd.Flags().Int32("retention-period-in-days", 1, "[REQUIRED] Retention period of the backup in days.")
	updatePolicyCmd.MarkFlagRequired("retention-period-in-days")
	updatePolicyCmd.Flags().Int32("full-backup-frequency-in-days", 1, "[OPTIONAL] Frequency of full backup in days.")
	if util.IsFeatureFlagEnabled(util.INCREMENTAL_BACKUP) {
		updatePolicyCmd.Flags().Int32("incremental-backup-frequency-in-minutes", 60, "[OPTIONAL] Frequency of incremental backup in minutes.")
	}
	updatePolicyCmd.Flags().String("full-backup-schedule-days-of-week", "", "[OPTIONAL] Days of the week when the backup has to run. A comma separated list of the first two letters of the days of the week. Eg: 'Mo,Tu,Sa'")
	updatePolicyCmd.Flags().String("full-backup-schedule-time", "", "[OPTIONAL] Time of the day at which the backup has to run. Please specify local time in 24 hr HH:MM format. Eg: 15:04")
	updatePolicyCmd.MarkFlagsRequiredTogether("full-backup-schedule-days-of-week", "full-backup-schedule-time")
	updatePolicyCmd.MarkFlagsOneRequired("full-backup-frequency-in-days", "full-backup-schedule-days-of-week", "full-backup-schedule-time")
	updatePolicyCmd.MarkFlagsMutuallyExclusive("full-backup-frequency-in-days", "full-backup-schedule-days-of-week")
	updatePolicyCmd.MarkFlagsMutuallyExclusive("full-backup-frequency-in-days", "full-backup-schedule-time")

}
