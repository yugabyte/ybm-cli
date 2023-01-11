/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var getCdcStreamCmd = &cobra.Command{
	Use:   "cdc_stream",
	Short: "Get CDC Stream in YugabyteDB Managed",
	Long:  `Get CDC Stream in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, cdcStreamIDOk, _ := getCdcStreamID(context.Background(), apiClient, accountID, cdcStreamName)

		if !cdcStreamIDOk {
			fmt.Fprintf(os.Stderr, "No Cdc Stream named `%s` found\n", cdcStreamName)
			return
		}

		resp, r, err := apiClient.CdcApi.GetCdcStream(context.Background(), accountID, projectID, clusterID, cdcStreamID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.GetCdcStream``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var createCdcStreamCmd = &cobra.Command{
	Use:   "cdc_stream",
	Short: "Create CDC Stream in YugabyteDB Managed",
	Long:  `Create CDC Stream in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)
		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		// TODO: handle failures in the above
		fmt.Fprintf(os.Stderr, "accountID: %v, projectID: %v, clusterID: %v", accountID, projectID, clusterID)

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcSinkName, _ := cmd.Flags().GetString("sink")
		sinkId, _, _ := getCdcSinkID(context.Background(), apiClient, accountID, cdcSinkName)
		dbName, _ := cmd.Flags().GetString("db-name")
		tables, _ := cmd.Flags().GetStringArray("tables")
		snapshotExistingData, _ := cmd.Flags().GetBool("snapshot-existing-data")
		kafkaPrefix, _ := cmd.Flags().GetString("kafka-prefix")

		cdcStreamSpec := ybmclient.CdcStreamSpec{
			Name:                 cdcStreamName,
			CdcSinkId:            sinkId,
			DbName:               dbName,
			Tables:               tables,
			SnapshotExistingData: &snapshotExistingData,
			KafkaPrefix:          &kafkaPrefix,
		}

		resp, r, err := apiClient.CdcApi.CreateCdcStream(context.Background(), accountID, projectID, clusterID).CdcStreamSpec(cdcStreamSpec).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.CreateCdcStream``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var editCdcStreamCmd = &cobra.Command{
	Use:   "cdc_stream",
	Short: "Edit CDC Stream in YugabyteDB Managed",
	Long:  `Edit CDC Stream in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, cdcStreamIDOk, _ := getCdcStreamID(context.Background(), apiClient, accountID, cdcStreamName)
		if !cdcStreamIDOk {
			fmt.Fprintf(os.Stderr, "No Cdc Stream named `%s` found\n", cdcStreamName)
			return
		}

		editCdcStreamRequest := ybmclient.NewEditCdcStreamRequest()
		if cmd.Flags().Changed("new-name") {
			updatedName, _ := cmd.Flags().GetString("new-name")
			editCdcStreamRequest.SetName(updatedName)
		}

		if cmd.Flags().Changed("tables") {
			tables, _ := cmd.Flags().GetStringArray("tables")
			editCdcStreamRequest.SetTables(tables)
		}

		resp, r, err := apiClient.CdcApi.EditCdcStream(context.Background(), accountID, projectID, clusterID, cdcStreamID).EditCdcStreamRequest(*editCdcStreamRequest).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.EditCdcStream`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var deleteCdcStreamCmd = &cobra.Command{
	Use:   "cdc_stream",
	Short: "Delete CDC Stream in YugabyteDB Managed",
	Long:  `Delete CDC Stream in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, _ := getApiClient(context.Background(), cmd)
		accountID, _, _ := getAccountID(context.Background(), apiClient)
		projectID, _, _ := getProjectID(context.Background(), apiClient, accountID)

		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, _, _ := getClusterID(context.Background(), apiClient, accountID, projectID, clusterName)

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, cdcStreamIDOk, _ := getCdcStreamID(context.Background(), apiClient, accountID, cdcStreamName)
		if !cdcStreamIDOk {
			fmt.Fprintf(os.Stderr, "No Cdc Stream named `%s` found\n", cdcStreamName)
			return
		}

		resp, err := apiClient.CdcApi.DeleteCdcStream(context.Background(), accountID, projectID, clusterID, cdcStreamID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.DeleteCdcStream`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", resp)
			return
		}

		prettyPrintJson(resp)
	},
}

func init() {
	getCmd.AddCommand(getCdcStreamCmd)
	getCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	getCdcStreamCmd.Flags().String("cluster", "", "Name of the Cluster")

	createCmd.AddCommand(createCdcStreamCmd)
	createCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	createCdcStreamCmd.Flags().String("cluster", "", "Name of the Cluster")
	createCdcStreamCmd.Flags().StringArray("tables", []string{}, "Database tables the Cdc Stream will listen to")
	createCdcStreamCmd.Flags().String("sink", "", "Destination sink for the CDC Stream")
	createCdcStreamCmd.Flags().String("db-name", "", "Database that the Cdc Stream will listen to")
	createCdcStreamCmd.Flags().String("snapshot-existing-data", "", "Whether to snapshot the existing data in the database")
	createCdcStreamCmd.Flags().String("kafka-prefix", "", "A prefix for the Kafka topics")

	updateCmd.AddCommand(editCdcStreamCmd)
	editCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	editCdcStreamCmd.Flags().String("cluster", "", "Name of the Cluster")
	editCdcStreamCmd.Flags().String("new-name", "", "Updated name of the CDC Stream")
	editCdcStreamCmd.Flags().StringArray("tables", []string{}, "Tables the Cdc Stream will listen to")

	deleteCmd.AddCommand(deleteCdcStreamCmd)
	deleteCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	deleteCdcStreamCmd.Flags().String("cluster", "", "Name of the Cluster")

}
