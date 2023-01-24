/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var getCdcStreamCmd = &cobra.Command{
	Use:   "cdc_stream",
	Short: "Get CDC Stream in YugabyteDB Managed",
	Long:  `Get CDC Stream in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, err := authApi.GetClusterID(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, err := authApi.GetCdcStreamIDByStreamName(cdcStreamName)
		if err != nil {
			logrus.Errorf("Error when getting StreamId with the name %s: %v", cdcStreamName, err)
			return
		}

		resp, r, err := authApi.GetCdcStream(clusterID, cdcStreamID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `CdcApi.GetCdcStream`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
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
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, err := authApi.GetClusterID(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcSinkName, _ := cmd.Flags().GetString("sink")
		sinkId, _ := authApi.GetCdcSinkIDBySinkName(cdcSinkName)

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

		resp, r, err := authApi.CreateCdcStream(clusterID).CdcStreamSpec(cdcStreamSpec).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `CdcApi.CreateCdcStream`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
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
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, err := authApi.GetClusterID(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, err := authApi.GetCdcStreamIDByStreamName(cdcStreamName)
		if err != nil {
			logrus.Errorf("Error when getting StreamId with the name %s: %v", cdcStreamName, err)
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

		resp, r, err := authApi.EditCdcStream(clusterID, cdcStreamID).EditCdcStreamRequest(*editCdcStreamRequest).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `CdcApi.EditCdcStream`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", r)
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
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: ", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster")
		clusterID, err := authApi.GetClusterID(clusterName)
		if err != nil {
			logrus.Error(err)
			return
		}

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, err := authApi.GetCdcStreamIDByStreamName(cdcStreamName)
		if err != nil {
			logrus.Errorf("Error when getting StreamId with the name %s: %v", cdcStreamName, err)
			return
		}
		resp, err := authApi.DeleteCdcStream(clusterID, cdcStreamID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `CdcApi.DeleteCdcStream`: %v\n", err)
			logrus.Debugf("Full HTTP response: %v\n", resp)
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
