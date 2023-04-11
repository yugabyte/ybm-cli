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

package backup

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage backup operations of a cluster",
	Long:  "Manage backup operations of a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getBackupCmd = &cobra.Command{
	Use:   "get",
	Short: "Get list of existing backups available for a cluster in YugabyteDB Managed",
	Long:  "Get list of existing backups available for a cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		listBackupCmd.Run(cmd, args)
		logrus.Warnln("\nThe command `ybm backup get` is deprecated. Please use `ybm backup list` instead.")
	},
}

var listBackupCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing backups available for a cluster in YugabyteDB Managed",
	Long:  "List existing backups available for a cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		listBackupRequest := authApi.ListBackups()
		if cmd.Flags().Changed("cluster-name") {
			clusterName, _ := cmd.Flags().GetString("cluster-name")
			clusterID, err := authApi.GetClusterIdByName(clusterName)
			if err != nil {
				logrus.Fatal(err)
			}
			listBackupRequest = listBackupRequest.ClusterId(clusterID)
		}
		resp, r, err := listBackupRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		backupsCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewBackupFormat(viper.GetString("output")),
		}

		formatter.BackupWrite(backupsCtx, resp.GetData())
	},
}

// TODO: implement a backup describe command that shows the details of a backup

var restoreBackupCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore backups into a cluster in YugabyteDB Managed",
	Long:  "Restore backups into a cluster in  YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		backupID, _ := cmd.Flags().GetString("backup-id")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		restoreSpec := ybmclient.NewRestoreSpec()
		restoreSpec.SetBackupId(backupID)
		restoreSpec.SetClusterId(clusterID)

		_, r, err := authApi.RestoreBackup().RestoreSpec(*restoreSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		msg := fmt.Sprintf("Backup %v is being restored onto the cluster %v", formatter.Colorize(backupID, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_RESTORE_BACKUP, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("Backup %v has been restored onto the cluster %v\n", formatter.Colorize(backupID, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))
			return
		} else {
			fmt.Println(msg)
		}
	},
}

var createBackupCmd = &cobra.Command{
	Use:   "create",
	Short: "Create backup for a cluster in YugabyteDB Managed",
	Long:  "Create backup for a cluster in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		createBackupSpec := *ybmclient.NewBackupSpecWithDefaults()
		createBackupSpec.SetClusterId(clusterID)
		// Set default retention period to 1 day
		retentionPeriod := int32(1)
		if cmd.Flags().Changed("retention-period") {
			retentionPeriod, _ = cmd.Flags().GetInt32("retention-period")
			createBackupSpec.SetRetentionPeriodInDays(retentionPeriod)
		} else {

			createBackupSpec.SetRetentionPeriodInDays(retentionPeriod)
		}
		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			createBackupSpec.SetDescription(description)
		}

		backupResp, response, err := authApi.CreateBackup().BackupSpec(createBackupSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		backupID := backupResp.GetData().Info.Id

		msg := fmt.Sprintf("The backup for cluster %s is being created", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(*backupID, ybmclient.ENTITYTYPEENUM_BACKUP, ybmclient.TASKTYPEENUM_CREATE_BACKUP, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The backup for cluster %s has been created\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.GetBackup(*backupID).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			backupResp = respC
		} else {
			fmt.Println(msg)
		}

		backupsCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewBackupFormat(viper.GetString("output")),
		}

		formatter.BackupWrite(backupsCtx, []ybmclient.BackupData{backupResp.GetData()})
	},
}

var deleteBackupCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete backup for a cluster in YugabyteDB Managed",
	Long:  "Delete backup for a cluster in YugabyteDB Managed",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		backupID, _ := cmd.Flags().GetString("backup-id")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to delete %s: %s", "backup", backupID), viper.GetBool("force"))
		if err != nil {
			logrus.Fatal(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		backupID, _ := cmd.Flags().GetString("backup-id")

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		response, err := authApi.DeleteBackup(backupID).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("The backup %s is being queued for deletion.\n", formatter.Colorize(backupID, formatter.GREEN_COLOR))
	},
}

func init() {
	BackupCmd.AddCommand(getBackupCmd)
	getBackupCmd.Flags().String("cluster-name", "", "[OPTIONAL] Name of the cluster to fetch backups.")

	BackupCmd.AddCommand(listBackupCmd)
	listBackupCmd.Flags().String("cluster-name", "", "[OPTIONAL] Name of the cluster to fetch backups.")

	BackupCmd.AddCommand(restoreBackupCmd)
	restoreBackupCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster to restore backups.")
	restoreBackupCmd.MarkFlagRequired("cluster-name")
	restoreBackupCmd.Flags().String("backup-id", "", "[REQUIRED] ID of the backup to be restored.")
	restoreBackupCmd.MarkFlagRequired("backup-id")

	BackupCmd.AddCommand(createBackupCmd)
	createBackupCmd.Flags().String("cluster-name", "", "[REQUIRED] Name for the cluster.")
	createBackupCmd.MarkFlagRequired("cluster-name")
	createBackupCmd.Flags().Int32("retention-period", 0, "[OPTIONAL] Retention period of the backup in days. (Default: 1)")
	createBackupCmd.Flags().String("description", "", "[OPTIONAL] Description of the backup.")

	BackupCmd.AddCommand(deleteBackupCmd)
	deleteBackupCmd.Flags().String("backup-id", "", "[REQUIRED] The backup ID.")
	deleteBackupCmd.MarkFlagRequired("backup-id")
	deleteBackupCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

}
