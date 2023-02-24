// Copyright (c) YugaByte, Inc.
//
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
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022-present Yugabyte, Inc.

package cdc

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
	Use:   "stream",
	Short: "Manage Change Data Capture stream operations",
	Long:  "Manage Change Data Capture stream operations",
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		cdcStreamRequest := authApi.ListCdcStreamsForAccount()
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		if clusterName != "" {
			clusterID, err := authApi.GetClusterIdByName(clusterName)
			if err != nil {
				logrus.Fatal(err)
			}
			cdcStreamRequest.ClusterId(clusterID)
		}
		cdcStreamName, _ := cmd.Flags().GetString("name")
		if cdcStreamName != "" {
			cdcStreamRequest.Name(cdcStreamName)
		}

		resp, r, err := cdcStreamRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `CdcApi.GetCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcSinkName, _ := cmd.Flags().GetString("sink")
		sinkId, err := authApi.GetCdcSinkIDBySinkName(cdcSinkName)
		if err != nil {
			logrus.Fatalf("Please provide a valid sink name: %s", err)
		}

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
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `CdcApi.CreateCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The CDC stream %s is being created", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "CREATE_CDC_SERVICE", []string{"FAILED", "SUCCEEDED"}, msg, 1200)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, err := authApi.GetCdcStreamIDByStreamName(cdcStreamName)
		if err != nil {
			logrus.Fatalf("Error when getting StreamId with the name %s: %v", cdcStreamName, err)
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
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `CdcApi.EditCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The CDC stream %s is being updated", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {

			if cmd.Flags().Changed("tables") {
				returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "RECONFIGURE_CDC_SERVICE", []string{"FAILED", "SUCCEEDED"}, msg, 1200)
				if err != nil {
					logrus.Fatalf("error when getting task status: %s", err)
				}
				if returnStatus != "SUCCEEDED" {
					logrus.Fatalf("Operation failed with error: %s", returnStatus)
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterID, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatal(err)
		}

		cdcStreamName, _ := cmd.Flags().GetString("name")
		cdcStreamID, err := authApi.GetCdcStreamIDByStreamName(cdcStreamName)
		if err != nil {
			logrus.Errorf("Error when getting StreamId with the name %s: %v", cdcStreamName, err)
			return
		}
		resp, err := authApi.DeleteCdcStream(cdcStreamID, clusterID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf("Error when calling `CdcApi.DeleteCdcStream`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		msg := fmt.Sprintf("The CDC stream %s is being deleted", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, "CLUSTER", "DELETE_CDC_SERVICE", []string{"FAILED", "SUCCEEDED"}, msg, 1200)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The CDC stream %s has been deleted\n", formatter.Colorize(cdcStreamName, formatter.GREEN_COLOR))
		} else {
			fmt.Println(msg)
		}

	},
}

func init() {
	CdcCmd.AddCommand(CDCStreamCmd)

	CDCStreamCmd.AddCommand(getCdcStreamCmd)
	getCdcStreamCmd.Flags().String("name", "", "[OPTIONAL] Name of the CDC Stream.")
	getCdcStreamCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the Cluster.")
	getCdcStreamCmd.MarkFlagRequired("cluster-name")

	CDCStreamCmd.AddCommand(createCdcStreamCmd)
	createCdcStreamCmd.Flags().String("name", "", "[REQUIRED] Name of the CDC Stream.")
	createCdcStreamCmd.MarkFlagRequired("name")
	createCdcStreamCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the Cluster.")
	createCdcStreamCmd.MarkFlagRequired("cluster-name")
	createCdcStreamCmd.Flags().StringArray("tables", []string{}, "[REQUIRED] Database tables the CDC Stream will listen to.")
	createCdcStreamCmd.MarkFlagRequired("tables")
	createCdcStreamCmd.Flags().String("sink", "", "[REQUIRED] Destination sink for the CDC Stream.")
	createCdcStreamCmd.MarkFlagRequired("sink")
	createCdcStreamCmd.Flags().String("db-name", "", "[REQUIRED] Database that the CDC Stream will listen to.")
	createCdcStreamCmd.MarkFlagRequired("db-name")
	createCdcStreamCmd.Flags().Bool("snapshot-existing-data", false, "[OPTIONAL] Whether to snapshot the existing data in the database.")
	createCdcStreamCmd.Flags().String("kafka-prefix", "", "[OPTIONAL] A prefix for the Kafka topics.")

	CDCStreamCmd.AddCommand(editCdcStreamCmd)
	editCdcStreamCmd.Flags().String("name", "", "[REQUIRED] Name of the CDC Stream.")
	editCdcStreamCmd.MarkFlagRequired("name")
	editCdcStreamCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the Cluster.")
	editCdcStreamCmd.MarkFlagRequired("cluster-name")
	editCdcStreamCmd.Flags().String("new-name", "", "[OPTIONAL] Updated name of the CDC Stream.")
	editCdcStreamCmd.Flags().StringArray("tables", []string{}, "[OPTIONAL] Tables the Cdc Stream will listen to.")

	CDCStreamCmd.AddCommand(deleteCdcStreamCmd)
	deleteCdcStreamCmd.Flags().String("name", "", "[REQUIRED] Name of the CDC Stream.")
	deleteCdcStreamCmd.MarkFlagRequired("name")
	deleteCdcStreamCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the Cluster.")
	deleteCdcStreamCmd.MarkFlagRequired("cluster-name")

}
