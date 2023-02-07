// Copyright (c) YugaByte, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cdcstream

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

func printCdcStreamOutput(cdcStreamData []ybmclient.CdcStreamData) {
	cdcStreamCtx := formatter.Context{
		Output: os.Stdout,
		Format: formatter.NewCdcStreamFormat(viper.GetString("output")),
	}

	formatter.CdcStreamWrite(cdcStreamCtx, cdcStreamData)
}

var CDCStreamCmd = &cobra.Command{
	Use:   "cdc-stream",
	Short: "cdc-stream",
	Long:  "CDC stream commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getCdcStreamCmd = &cobra.Command{
	Use:   "get",
	Short: "Get CDC Stream in YugabyteDB Managed",
	Long:  "Get CDC Stream in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")
		cdcStreamRequest := authApi.ListCdcStreamsForAccount()
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName != "" {
			clusterID, err := authApi.GetClusterIdByName(clusterName)
			if err != nil {
				logrus.Error(err)
				return
			}
			cdcStreamRequest.ClusterId(clusterID)
		}
		cdcStreamName, _ := cmd.Flags().GetString("name")
		if cdcStreamName != "" {
			cdcStreamRequest.Name(cdcStreamName)
		}

		resp, r, err := cdcStreamRequest.Execute()
		if err != nil {
			logrus.Errorf("Error when calling `CdcApi.GetCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}
		printCdcStreamOutput(resp.GetData())
	},
}

var createCdcStreamCmd = &cobra.Command{
	Use:   "create",
	Short: "Create CDC Stream in YugabyteDB Managed",
	Long:  `Create CDC Stream in YugabyteDB Managed`,
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
			logrus.Errorf("Error when calling `CdcApi.CreateCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		msg := fmt.Sprintf("The CDC stream %s is being created", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "CREATE_CDC_SERVICE", []string{"FAILED", "SUCCEEDED"}, msg, 1200)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
			}
			fmt.Printf("The CDC stream %s has been created\n", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))
		} else {
			fmt.Println(msg)
		}

		printCdcStreamOutput([]ybmclient.CdcStreamData{resp.GetData()})
	},
}

var editCdcStreamCmd = &cobra.Command{
	Use:   "update",
	Short: "Update CDC Stream in YugabyteDB Managed",
	Long:  "Update CDC Stream in YugabyteDB Managed",
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

		resp, r, err := authApi.EditCdcStream(cdcStreamID, clusterID).EditCdcStreamRequest(*editCdcStreamRequest).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `CdcApi.EditCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", r)
			return
		}

		msg := fmt.Sprintf("The CDC stream %s is being updated", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {

			if cmd.Flags().Changed("tables") {
				returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "RECONFIGURE_CDC_SERVICE", []string{"FAILED", "SUCCEEDED"}, msg, 1200)
				if err != nil {
					logrus.Errorf("error when getting task status: %s", err)
					return
				}
				if returnStatus != "SUCCEEDED" {
					logrus.Errorf("Operation failed with error: %s", returnStatus)
					return
				}
			}
			fmt.Printf("The CDC stream %s has been updated\n", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))
		} else {
			if cmd.Flags().Changed("tables") {
				fmt.Println(msg)
			} else {
				fmt.Printf("The CDC stream %s has been updated\n", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))
			}
		}

		printCdcStreamOutput([]ybmclient.CdcStreamData{resp.GetData()})
	},
}

var deleteCdcStreamCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete CDC Stream in YugabyteDB Managed",
	Long:  `Delete CDC Stream in YugabyteDB Managed`,
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

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, err := authApi.GetCdcStreamIDByStreamName(cdcStreamName)
		if err != nil {
			logrus.Errorf("Error when getting StreamId with the name %s: %v", cdcStreamName, err)
			return
		}
		resp, err := authApi.DeleteCdcStream(cdcStreamID, clusterID).Execute()
		if err != nil {
			logrus.Errorf("Error when calling `CdcApi.DeleteCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
			logrus.Debugf("Full HTTP response: %v", resp)
			return
		}

		msg := fmt.Sprintf("The CDC stream %s is being deleted", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "DELETE_CDC_SERVICE", []string{"FAILED", "SUCCEEDED"}, msg, 1200)
			if err != nil {
				logrus.Errorf("error when getting task status: %s", err)
				return
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Errorf("Operation failed with error: %s", returnStatus)
				return
			}
			fmt.Printf("The CDC stream %s has been deleted\n", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))
		} else {
			fmt.Println(msg)
		}

	},
}

func init() {
	CDCStreamCmd.AddCommand(getCdcStreamCmd)
	getCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	getCdcStreamCmd.Flags().String("cluster-name", "", "Name of the Cluster")

	CDCStreamCmd.AddCommand(createCdcStreamCmd)
	createCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	createCdcStreamCmd.Flags().String("cluster-name", "", "Name of the Cluster")
	createCdcStreamCmd.Flags().StringArray("tables", []string{}, "Database tables the Cdc Stream will listen to")
	createCdcStreamCmd.Flags().String("sink", "", "Destination sink for the CDC Stream")
	createCdcStreamCmd.Flags().String("db-name", "", "Database that the Cdc Stream will listen to")
	createCdcStreamCmd.Flags().String("snapshot-existing-data", "", "Whether to snapshot the existing data in the database")
	createCdcStreamCmd.Flags().String("kafka-prefix", "", "A prefix for the Kafka topics")

	CDCStreamCmd.AddCommand(editCdcStreamCmd)
	editCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	editCdcStreamCmd.Flags().String("cluster-name", "", "Name of the Cluster")
	editCdcStreamCmd.Flags().String("new-name", "", "Updated name of the CDC Stream")
	editCdcStreamCmd.Flags().StringArray("tables", []string{}, "Tables the Cdc Stream will listen to")

	CDCStreamCmd.AddCommand(deleteCdcStreamCmd)
	deleteCdcStreamCmd.Flags().String("name", "", "Name of the CDC Stream")
	deleteCdcStreamCmd.Flags().String("cluster-name", "", "Name of the Cluster")

}
