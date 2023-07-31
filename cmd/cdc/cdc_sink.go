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

func printCdcSinkOutput(cdcSinkData []ybmclient.CdcSinkData) {

	cdcSinkCtx := formatter.Context{
		Output: os.Stdout,
		Format: formatter.NewCdcSinkFormat(viper.GetString("output")),
	}

	formatter.CdcSinkWrite(cdcSinkCtx, cdcSinkData)

}

var CDCSinkCmd = &cobra.Command{
	Use:   "sink",
	Short: "Manage Change Data Capture Sink operations",
	Long:  "Manage Change Data Capture Sink operations",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listCdcSinkCmd = &cobra.Command{
	Use:   "list",
	Short: "List CDC Sinks in YugabyteDB Managed",
	Long:  `List CDC Sinks in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		cdcSinkRequest := authApi.ListCdcSinks()
		cdcSinkName, _ := cmd.Flags().GetString("name")
		if cdcSinkName != "" {
			cdcSinkRequest = cdcSinkRequest.Name(cdcSinkName)
		}

		resp, r, err := cdcSinkRequest.Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		printCdcSinkOutput(resp.GetData())
	},
}

// TODO: implement describe that shows the details of a sink

var createCdcSinkCmd = &cobra.Command{
	Use:   "create",
	Short: "Create CDC Sink in YugabyteDB Managed",
	Long:  `Create CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		cdcSinkName, _ := cmd.Flags().GetString("name")
		sinkType, _ := cmd.Flags().GetString("cdc-sink-type")
		hostname, _ := cmd.Flags().GetString("hostname")
		authType, _ := cmd.Flags().GetString("auth-type")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		sinkTypeEnum, err := ybmclient.NewCdcSinkTypeEnumFromValue(sinkType)
		if err != nil {
			logrus.Fatalf("Please provide a valid sink type: %s", err)
		}
		kafkaSpec := ybmclient.NewCdcSinkKafka(hostname)

		cdcSinkSpec := ybmclient.CdcSinkSpec{
			Name:     cdcSinkName,
			SinkType: *sinkTypeEnum,
			Kafka:    kafkaSpec,
		}

		authTypeEnum, err := ybmclient.NewCdcSinkAuthTypeEnumFromValue(authType)
		if err != nil {
			logrus.Fatalf("Please provide a valid auth type: %s", err)
		}
		cdcSinkAuthSpec := ybmclient.NewCdcSinkAuthSpec(*authTypeEnum)
		cdcSinkAuthSpec.SetUsername(username)
		cdcSinkAuthSpec.SetPassword(password)

		createSinkRequest := ybmclient.NewCreateCdcSinkRequest(cdcSinkSpec, *cdcSinkAuthSpec)

		resp, r, err := authApi.CreateCdcSink().CreateCdcSinkRequest(*createSinkRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		printCdcSinkOutput([]ybmclient.CdcSinkData{resp.GetData()})
	},
}

var editCdcSinkCmd = &cobra.Command{
	Use:   "update",
	Short: "Update CDC Sink in YugabyteDB Managed",
	Long:  "Update CDC Sink in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		cdcSinkName, _ := cmd.Flags().GetString("name")

		cdcSinkID, err := authApi.GetCdcSinkIDBySinkName(cdcSinkName)
		if err != nil {
			logrus.Fatalf("No Cdc Sink named `%s` found: %v", cdcSinkName, err)
		}

		editCdcSinkRequest := ybmclient.NewEditCdcSinkRequest()
		if cmd.Flags().Changed("new-name") {
			updatedName, _ := cmd.Flags().GetString("new-name")
			editCdcSinkRequest.SetName(updatedName)
		}

		if cmd.Flags().Changed("auth-type") {
			updatedAuthType, _ := cmd.Flags().GetString("auth-type")
			updatedAuthTypeEnum, err := ybmclient.NewCdcSinkAuthTypeEnumFromValue(updatedAuthType)
			if err != nil {
				logrus.Fatalf("Please provide a valid auth type: %s", err)
			}
			editCdcSinkRequest.Auth.SetAuthType(*updatedAuthTypeEnum)
		}

		if cmd.Flags().Changed("username") {
			updatedUsername, _ := cmd.Flags().GetString("username")
			editCdcSinkRequest.Auth.SetUsername(updatedUsername)
		}

		if cmd.Flags().Changed("password") {
			updatedPassword, _ := cmd.Flags().GetString("password")
			editCdcSinkRequest.Auth.SetPassword(updatedPassword)
		}

		resp, r, err := authApi.EditCdcSink(cdcSinkID).EditCdcSinkRequest(*editCdcSinkRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		printCdcSinkOutput([]ybmclient.CdcSinkData{resp.GetData()})
	},
}

var deleteCdcSinkCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete CDC Sink in YugabyteDB Managed",
	Long:  `Delete CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		cdcSinkName, _ := cmd.Flags().GetString("name")

		cdcSinkID, err := authApi.GetCdcSinkIDBySinkName(cdcSinkName)
		if err != nil {
			logrus.Fatalf("No Cdc Sink named `%s` found: %v", cdcSinkName, err)
		}

		resp, err := authApi.DeleteCdcSink(cdcSinkID).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Fprintf(os.Stdout, "CDC sink deleted successfully")
	},
}

func init() {
	CdcCmd.AddCommand(CDCSinkCmd)

	CDCSinkCmd.AddCommand(listCdcSinkCmd)
	listCdcSinkCmd.Flags().String("name", "", "[OPTIONAL] Name of the CDC sink.")

	CDCSinkCmd.AddCommand(createCdcSinkCmd)
	createCdcSinkCmd.Flags().String("name", "", "[REQUIRED] Name of the CDC sink.")
	createCdcSinkCmd.MarkFlagRequired("name")
	createCdcSinkCmd.Flags().String("cdc-sink-type", "", "[REQUIRED] Name of the CDC sink type.")
	createCdcSinkCmd.MarkFlagRequired("cdc-sink-type")
	createCdcSinkCmd.Flags().String("auth-type", "", "[REQUIRED] Name of the CDC sink authentication type.")
	createCdcSinkCmd.MarkFlagRequired("auth-type")
	createCdcSinkCmd.Flags().String("hostname", "", "[REQUIRED] Hostname of the CDC sink.")
	createCdcSinkCmd.MarkFlagRequired("hostname")
	createCdcSinkCmd.Flags().String("username", "", "[REQUIRED] Username of the CDC sink.")
	createCdcSinkCmd.MarkFlagRequired("username")
	createCdcSinkCmd.Flags().String("password", "", "[REQUIRED] Password of the CDC sink.")
	createCdcSinkCmd.MarkFlagRequired("password")

	CDCSinkCmd.AddCommand(editCdcSinkCmd)
	editCdcSinkCmd.Flags().String("name", "", "[REQUIRED] Name of the CDC Sink.")
	editCdcSinkCmd.MarkFlagRequired("name")
	editCdcSinkCmd.Flags().String("new-name", "", "[OPTIONAL] Name of the new CDC Sink.")
	editCdcSinkCmd.Flags().String("auth-type", "", "[OPTIONAL] Name of the new CDC Sink.")
	editCdcSinkCmd.Flags().String("username", "", "[OPTIONAL] Username of the CDC Sink.")
	editCdcSinkCmd.Flags().String("password", "", "[OPTIONAL] Password of the CDC Sink.")

	CDCSinkCmd.AddCommand(deleteCdcSinkCmd)
	deleteCdcSinkCmd.Flags().String("name", "", "[REQUIRED] Name of the CDC Sink.")
	deleteCdcSinkCmd.MarkFlagRequired("name")

}
