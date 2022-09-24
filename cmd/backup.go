package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var getBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Get backups in YugabyteDB Managed",
	Long:  "Get backups in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {

		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		listBackupRequest := apiClient.BackupApi.ListBackups(context.Background(), accountID, projectID)
		if cmd.Flags().Changed("cluster-name") {
			clusterID, clusterIDOK, errMsg := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)
			if !clusterIDOK {
				fmt.Fprintf(os.Stderr, "Error when fetching cluster ID: %v\n", errMsg)
				return
			}
			listBackupRequest = listBackupRequest.ClusterId(clusterID)
		}
		resp, r, err := listBackupRequest.Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `BackupApi.ListBackups``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var restoreBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Restore backups in YugabyteDB Managed",
	Long:  "Restore backups in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {

		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		backupID, _ := cmd.Flags().GetString("backup-id")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, clusterIDOK, errMsg := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)
		if !clusterIDOK {
			fmt.Fprintf(os.Stderr, "Error when fetching cluster ID: %v\n", errMsg)
			return
		}

		restoreSpec := ybmclient.NewRestoreSpec()
		restoreSpec.SetBackupId(backupID)
		restoreSpec.SetClusterId(clusterID)

		_, r, err := apiClient.BackupApi.RestoreBackup(context.Background(), accountID, projectID).RestoreSpec(*restoreSpec).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `BackupApi.RestoreBackup``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}
		fmt.Fprintf(os.Stdout, "The backup %v is being restored onto the cluster %v\n", backupID, clusterName)
	},
}

var createBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create backup in YugabyteDB Managed",
	Long:  "Create backup in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, clusterIDOK, errMsg := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)
		if !clusterIDOK {
			fmt.Fprintf(os.Stderr, "Error when fetching cluster ID: %v\n", errMsg)
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

		backupResp, response, err := apiClient.BackupApi.CreateBackup(context.Background(), accountID, projectID).BackupSpec(createBackupSpec).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `BackupApi.CreateBackup``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", response)
		}
		fmt.Fprintf(os.Stdout, "The backup for cluster %v is being created\n", clusterName)

		prettyPrintJson(backupResp)
	},
}

var deleteBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Delete backup in YugabyteDB Managed",
	Long:  "Delete backup in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		backupID, _ := cmd.Flags().GetString("backup-id")

		apiClient, _ := getApiClient(context.Background())
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		r, err := apiClient.BackupApi.DeleteBackup(context.Background(), accountID, projectID, backupID).Execute()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `BackupApi.DeleteBackup``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		fmt.Fprintf(os.Stdout, "Backup %v was queued for deletion.\n", backupID)
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
