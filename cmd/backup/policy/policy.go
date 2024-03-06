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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/backup/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	"github.com/yugabyte/ybm-cli/internal/formatter/backuppolicy"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

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
			Format: backuppolicy.NewBackupPolicyFormat(viper.GetString("output")),
		}
		if len(resp.GetData()) < 1 {
			logrus.Fatalln("No backup policies found for the given cluster")
		}
		backuppolicy.BackupPolicyListWrite(policyCtx, resp.GetData())
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
		scheduleSpec := resp.GetData()[0].GetSpec()
		if scheduleSpec.GetState() == ybmclient.SCHEDULESTATEENUM_ACTIVE {
			logrus.Fatalf("The backup policy is already enabled for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		}
		scheduleSpec.SetState(ybmclient.SCHEDULESTATEENUM_ACTIVE)
		info := resp.GetData()[0].GetInfo()
		scheduleId := info.GetId()
		retentionPeriodInDaysResp, retentionFound := info.GetTaskParams()["retention_period_in_days"]
		if !retentionFound || retentionPeriodInDaysResp == nil {
			logrus.Fatalln("Unable to fetch retention period in days for the backup schedule")
		}
		retentionPeriodInDays := int32(retentionPeriodInDaysResp.(float64))
		descriptionResp, descriptionFound := info.GetTaskParams()["description"]
		if !descriptionFound || descriptionResp == nil {
			logrus.Fatalln("Unable to fetch description for the backup schedule")
		}
		description := descriptionResp.(string)
		backupSpec := ybmclient.NewBackupSpec(clusterId)
		backupSpec.SetRetentionPeriodInDays(retentionPeriodInDays)
		backupSpec.SetDescription(description)

		backupScheduleSpec := ybmclient.NewBackupScheduleSpec(*backupSpec, scheduleSpec)
		_, r, err = authApi.UpdateBackupPolicy(scheduleId).BackupScheduleSpec(*backupScheduleSpec).Execute()
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
			logrus.Fatalln("No ACTIVE backup policies found to disable for the given cluster")
		}
		scheduleSpec := resp.GetData()[0].GetSpec()
		if scheduleSpec.GetState() == ybmclient.SCHEDULESTATEENUM_PAUSED {
			logrus.Fatalf("The backup policy is already disabled for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))
		}
		scheduleSpec.SetState(ybmclient.SCHEDULESTATEENUM_PAUSED)
		info := resp.GetData()[0].GetInfo()
		scheduleId := info.GetId()
		retentionPeriodInDaysResp, retentionFound := info.GetTaskParams()["retention_period_in_days"]
		if !retentionFound || retentionPeriodInDaysResp == nil {
			logrus.Fatalln("Unable to fetch retention period in days for the backup schedule")
		}
		retentionPeriodInDays := int32(retentionPeriodInDaysResp.(float64))
		descriptionResp, descriptionFound := info.GetTaskParams()["description"]
		if !descriptionFound || descriptionResp == nil {
			logrus.Fatalln("Unable to fetch description for the backup schedule")
			return
		}
		description := descriptionResp.(string)
		backupSpec := ybmclient.NewBackupSpec(clusterId)
		backupSpec.SetRetentionPeriodInDays(retentionPeriodInDays)
		backupSpec.SetDescription(description)

		backupScheduleSpec := ybmclient.NewBackupScheduleSpec(*backupSpec, scheduleSpec)
		_, r, err = authApi.UpdateBackupPolicy(scheduleId).BackupScheduleSpec(*backupScheduleSpec).Execute()
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
		scheduleSpec := resp.GetData()[0].GetSpec()
		if cmd.Flags().Changed("full-backup-frequency-in-days") {
			frequencyInDays, _ := cmd.Flags().GetInt32("full-backup-frequency-in-days")
			if frequencyInDays < 1 {
				logrus.Fatalln("Time interval for scheduling backup should be greater than or equal to 1 day")
			}
			scheduleSpec.SetTimeIntervalInDays(frequencyInDays)
			scheduleSpec.UnsetCronExpression()
		} else {
			daysOfWeek, _ := cmd.Flags().GetString("full-backup-schedule-days-of-week")
			if !util.IsDaysOfWeekValid(daysOfWeek) {
				logrus.Fatalln("The days of week specified is incorrect. Please ensure that it is a comma separated list of the first two letters to days of the week.")
			}
			backupTime, _ := cmd.Flags().GetString("full-backup-schedule-time")
			if !util.IsTimeFormatValid(backupTime) {
				logrus.Fatalln("The full backup schedule time is invalid. Please ensure that it in the 24 Hr HH:MM format.")
			}
			backupTimeUTC := util.ConvertLocalTimeToUTC(backupTime)
			cronExpression := util.GenerateCronExpression(daysOfWeek, backupTimeUTC)
			scheduleSpec.SetCronExpression(cronExpression)
			scheduleSpec.TimeIntervalInDays = nil
		}

		info := resp.GetData()[0].GetInfo()
		scheduleId := info.GetId()
		description := info.GetTaskParams()["description"].(string)
		backupSpec := ybmclient.NewBackupSpec(clusterId)
		backupSpec.SetRetentionPeriodInDays(retentionPeriodInDays)
		backupSpec.SetDescription(description)

		backupScheduleSpec := ybmclient.NewBackupScheduleSpec(*backupSpec, scheduleSpec)
		_, r, err = authApi.UpdateBackupPolicy(scheduleId).BackupScheduleSpec(*backupScheduleSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("Successfully updated backup policy for cluster %s\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

	},
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
	updatePolicyCmd.Flags().String("full-backup-schedule-days-of-week", "", "[OPTIONAL] Days of the week when the backup has to run. A comma separated list of the first two letters of the days of the week. Eg: 'Mo,Tu,Sa'")
	updatePolicyCmd.Flags().String("full-backup-schedule-time", "", "[OPTIONAL] Time of the day at which the backup has to run. Please specify local time in 24 hr HH:MM format. Eg: 15:04")
	updatePolicyCmd.MarkFlagsRequiredTogether("full-backup-schedule-days-of-week", "full-backup-schedule-time")
	updatePolicyCmd.MarkFlagsOneRequired("full-backup-frequency-in-days", "full-backup-schedule-days-of-week", "full-backup-schedule-time")
	updatePolicyCmd.MarkFlagsMutuallyExclusive("full-backup-frequency-in-days", "full-backup-schedule-days-of-week")
	updatePolicyCmd.MarkFlagsMutuallyExclusive("full-backup-frequency-in-days", "full-backup-schedule-time")

}
