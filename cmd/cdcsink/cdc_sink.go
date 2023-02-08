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

package cdcsink

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
	Use:   "cdc-sink",
	Short: "cdc-sink",
	Long:  "Cdc Sink commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getCdcSinkCmd = &cobra.Command{
	Use:   "get",
	Short: "Get CDC Sink in YugabyteDB Managed",
	Long:  `Get CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
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
			logrus.Fatalf("Error when calling `CdcApi.GetCdcSink`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		printCdcSinkOutput(resp.GetData())
	},
}

var createCdcSinkCmd = &cobra.Command{
	Use:   "create",
	Short: "Create CDC Sink in YugabyteDB Managed",
	Long:  `Create CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		cdcSinkName, _ := cmd.Flags().GetString("name")
		sinkType, _ := cmd.Flags().GetString("cdc-sink-type")
		hostname, _ := cmd.Flags().GetString("hostname")
		authType, _ := cmd.Flags().GetString("auth-type")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		sinkTypeEnum, _ := ybmclient.NewCdcSinkTypeEnumFromValue(sinkType)
		kafkaSpec := ybmclient.NewCdcSinkKafka(hostname)

		cdcSinkSpec := ybmclient.CdcSinkSpec{
			Name:     cdcSinkName,
			SinkType: *sinkTypeEnum,
			Kafka:    kafkaSpec,
		}

		authTypeEnum, err := ybmclient.NewCdcSinkAuthTypeEnumFromValue(authType)
		if err != nil {
			logrus.Fatalf("Error when getting auth type enum from value: %s", err)
		}
		cdcSinkAuthSpec := ybmclient.NewCdcSinkAuthSpec(*authTypeEnum)
		cdcSinkAuthSpec.SetUsername(username)
		cdcSinkAuthSpec.SetPassword(password)

		createSinkRequest := ybmclient.NewCreateCdcSinkRequest(cdcSinkSpec, *cdcSinkAuthSpec)

		resp, r, err := authApi.CreateCdcSink().CreateCdcSinkRequest(*createSinkRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `CdcApi.CreateCdcSink`: %s", ybmAuthClient.GetApiErrorDetails(err))
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
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
			updatedAuthTypeEnum, _ := ybmclient.NewCdcSinkAuthTypeEnumFromValue(updatedAuthType)
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
			logrus.Fatalf("Error when calling `CdcApi.EditCdcSink`: %s", ybmAuthClient.GetApiErrorDetails(err))
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
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
			logrus.Fatalf("Error when calling `CdcApi.DeleteCdcSink`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Fprintf(os.Stdout, "CDC sink deleted successfully")
	},
}

func init() {
	CDCSinkCmd.AddCommand(getCdcSinkCmd)
	getCdcSinkCmd.Flags().String("name", "", "Name of the CDC Sink")

	CDCSinkCmd.AddCommand(createCdcSinkCmd)
	createCdcSinkCmd.Flags().String("name", "", "Name of the CDC sink")
	createCdcSinkCmd.Flags().String("cdc-sink-type", "", "Name of the CDC sink type")
	createCdcSinkCmd.Flags().String("auth-type", "", "Name of the CDC sink authentication type")
	createCdcSinkCmd.Flags().String("hostname", "", "Hostname of the CDC sink")
	createCdcSinkCmd.Flags().String("username", "", "Username of the CDC sink")
	createCdcSinkCmd.Flags().String("password", "", "Password of the CDC sink")

	CDCSinkCmd.AddCommand(editCdcSinkCmd)
	editCdcSinkCmd.Flags().String("name", "", "Name of the CDC Sink")
	editCdcSinkCmd.Flags().String("new-name", "", "Name of the new CDC Sink")
	editCdcSinkCmd.Flags().String("username", "", "Username of the CDC Sink")
	editCdcSinkCmd.Flags().String("password", "", "Password of the CDC Sink")

	CDCSinkCmd.AddCommand(deleteCdcSinkCmd)
	deleteCdcSinkCmd.Flags().String("name", "", "Name of the CDC Sink")

}
