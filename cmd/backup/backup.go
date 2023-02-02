package backup

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup",
	Long:  "Backup commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getBackupCmd = &cobra.Command{
	Use:   "get",
	Short: "Get backups in YugabyteDB Managed",
	Long:  "Get backups in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		listBackupRequest := authApi.ListBackups()
		if cmd.Flags().Changed("cluster-name") {
			clusterName, _ := cmd.Flags().GetString("cluster-name")
			clusterID, err := authApi.GetClusterIdByName(clusterName)
			if err != nil {
				logrus.Error(err)
				return
			}
			listBackupRequest = listBackupRequest.ClusterId(clusterID)
		}
		resp, r, err := listBackupRequest.Execute()
		if err != nil {
			logrus.Errorf("Error when calling `BackupApi.ListBackups`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
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
	Use:   "restore",
	Short: "Restore backups in YugabyteDB Managed",
	Long:  "Restore backups in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		backupID, _ := cmd.Flags().GetString("backup-id")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}

		restoreSpec := ybmclient.NewRestoreSpec()
		restoreSpec.SetBackupId(backupID)
		restoreSpec.SetClusterId(clusterID)

		_, r, err := authApi.RestoreBackup().RestoreSpec(*restoreSpec).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `BackupApi.RestoreBackup`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}
		msg := fmt.Sprintf("Backup %v is being restored onto the cluster %v", formatter.Colorize(backupID, formatter.GREEN_COLOR), formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "RESTORE_BACKUP", []string{"FAILED", "SUCCEEDED"}, msg, 1500)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
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
	Short: "Create backup in YugabyteDB Managed",
	Long:  "Create backup in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
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
			logrus.Errorf("Error when calling `BackupApi.CreateBackup`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", response)
			return
		}
		backupID := backupResp.GetData().Info.Id

		msg := fmt.Sprintf("The backup for cluster %s is being created", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(*backupID, "BACKUP", "CREATE_BACKUP", []string{"FAILED", "SUCCEEDED"}, msg, 1500)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
			}
			fmt.Printf("The backup for cluster %s has been created\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.GetBackup(*backupID).Execute()
			if err != nil {
				logrus.Errorf("Error when calling `BackupApi.ListBackups`: %s", ybmAuthClient.GetApiErrorDetails(err))
				logrus.Debugf("Full HTTP response: %v", r)
				return
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
	Short: "Delete backup in YugabyteDB Managed",
	Long:  "Delete backup in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		backupID, _ := cmd.Flags().GetString("backup-id")

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		response, err := authApi.DeleteBackup(backupID).Execute()

		if err != nil {
			logrus.Errorf("Error when calling `BackupApi.DeleteBackup`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", response)
			return
		}

		msg := fmt.Sprintf("The backup %s is being deleted", formatter.Colorize(backupID, formatter.GREEN_COLOR))

		//Seems delete backup do not yet create any task, not sure it's by design or not.
		// if viper.GetBool("wait") {
		// 	returnStatus, err := authApi.WaitForTaskCompletion(backupID, "BACKUP", "DELETE_BACKUP", []string{"FAILED", "SUCCEEDED"}, msg, 30)
		// 	if err != nil {
		// 		logrus.Errorf("error when getting task status: %s", err)
		// 		return
		// 	}
		// 	if returnStatus != "SUCCEEDED" {
		// 		logrus.Errorf("Operation failed with error: %s", returnStatus)
		// 		return
		// 	}
		// 	fmt.Printf("The backup %s has been deleted\n", formatter.Colorize(backupID, formatter.GREEN_COLOR))
		// 	return

		// }
		fmt.Println(msg)
	},
}

func init() {
	BackupCmd.AddCommand(getBackupCmd)
	getBackupCmd.Flags().String("cluster-name", "", "Name of the cluster to fetch backups")

	BackupCmd.AddCommand(restoreBackupCmd)
	restoreBackupCmd.Flags().String("cluster-name", "", "Name of the cluster to restore backups")
	restoreBackupCmd.MarkFlagRequired("cluster-name")
	restoreBackupCmd.Flags().String("backup-id", "", "ID of the backup to be restored")
	restoreBackupCmd.MarkFlagRequired("backup-id")

	BackupCmd.AddCommand(createBackupCmd)
	createBackupCmd.Flags().String("cluster-name", "", "Name for the cluster")
	createBackupCmd.MarkFlagRequired("cluster-name")
	createBackupCmd.Flags().Int32("retention-period", 0, "Retention period of the backup")
	createBackupCmd.Flags().String("description", "", "Description of the backup")

	BackupCmd.AddCommand(deleteBackupCmd)
	deleteBackupCmd.Flags().String("backup-id", "", "The backup ID")
	deleteBackupCmd.MarkFlagRequired("backup-id")
}
