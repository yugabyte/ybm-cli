package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var getBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Get backups in YugabyteDB Managed",
	Long:  "Get backups in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		listBackupRequest := authApi.ListBackups()
		if cmd.Flags().Changed("cluster-name") {
			clusterName, _ := cmd.Flags().GetString("cluster-name")
			clusterID, err := authApi.GetClusterID(clusterName)
			if err != nil {
				logrus.Error(err)
				return
			}
			listBackupRequest = listBackupRequest.ClusterId(clusterID)
		}
		resp, r, err := listBackupRequest.Execute()
		if err != nil {
			logrus.Errorf("Error when calling `BackupApi.ListBackups`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}
		backupsCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewBackupFormat(viper.GetString("output")),
		}

		formatter.BackupWrite(backupsCtx, resp.GetData())
	},
}

var restoreBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Restore backups in YugabyteDB Managed",
	Long:  "Restore backups in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		backupID, _ := cmd.Flags().GetString("backup-id")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterID(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}

		restoreSpec := ybmclient.NewRestoreSpec()
		restoreSpec.SetBackupId(backupID)
		restoreSpec.SetClusterId(clusterID)

		_, r, err := authApi.RestoreBackup().RestoreSpec(*restoreSpec).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `BackupApi.RestoreBackup``: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
			return
		}
		logrus.Infof("The backup %v is being restored onto the cluster %v\n", backupID, clusterName)
	},
}

var createBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create backup in YugabyteDB Managed",
	Long:  "Create backup in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterID(clusterName)
		if err != nil {
			logrus.Error(err)
			return
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
			logrus.Errorf("Error when calling `BackupApi.CreateBackup``: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", response)
			return
		}
		logrus.Infof("The backup for cluster %v is being created\n", clusterName)

		backupsCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewBackupFormat(viper.GetString("output")),
		}

		formatter.BackupWrite(backupsCtx, []ybmclient.BackupData{backupResp.GetData()})
	},
}

var deleteBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Delete backup in YugabyteDB Managed",
	Long:  "Delete backup in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		backupID, _ := cmd.Flags().GetString("backup-id")

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		response, err := authApi.DeleteBackup(backupID).Execute()

		if err != nil {
			logrus.Errorf("Error when calling `BackupApi.DeleteBackup``: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", response)
			return
		}

		logrus.Infof("Backup %v was queued for deletion.\n", backupID)
	},
}

func init() {
	getCmd.AddCommand(getBackupCmd)
	getBackupCmd.Flags().String("cluster-name", "", "Name of the cluster to fetch backups")

	restoreCmd.AddCommand(restoreBackupCmd)
	restoreBackupCmd.Flags().String("cluster-name", "", "Name of the cluster to restore backups")
	restoreBackupCmd.MarkFlagRequired("cluster-name")
	restoreBackupCmd.Flags().String("backup-id", "", "ID of the backup to be restored")
	restoreBackupCmd.MarkFlagRequired("backup-id")

	createCmd.AddCommand(createBackupCmd)
	createBackupCmd.Flags().String("cluster-name", "", "Name for the cluster")
	createBackupCmd.MarkFlagRequired("name")
	createBackupCmd.Flags().Int32("retention-period", 0, "Retention period of the backup")
	createBackupCmd.Flags().String("description", "", "Description of the backup")

	deleteCmd.AddCommand(deleteBackupCmd)
	deleteBackupCmd.Flags().String("backup-id", "", "The backup ID")
	deleteBackupCmd.MarkFlagRequired("backup-id")
}
