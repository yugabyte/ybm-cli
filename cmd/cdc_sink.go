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

var getCdcSinkCmd = &cobra.Command{
	Use:   "cdc_sink",
	Short: "Get CDC Sink in YugabyteDB Managed",
	Long:  `Get CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, _ := getApiRequestInfo("", "")

		cdcSinkName, _ := cmd.Flags().GetString("name")
		cdcSinkID, cdcSinkIDOk, _ := getCdcSinkID(context.Background(), apiClient, accountID, cdcSinkName)

		if !cdcSinkIDOk {
			fmt.Fprintf(os.Stderr, "No Cdc Sink named `%s` found\n", cdcSinkName)
			return
		}

		resp, r, err := apiClient.CdcApi.GetCdcSink(context.Background(), accountID, cdcSinkID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.GetCdcSink``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var createCdcSinkCmd = &cobra.Command{
	Use:   "cdc_sink",
	Short: "Create CDC Sink in YugabyteDB Managed",
	Long:  `Create CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, _ := getApiRequestInfo("", "")

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

		authTypeEnum, _ := ybmclient.NewCdcSinkAuthTypeEnumFromValue(authType)
		cdcSinkAuthSpec := ybmclient.NewCdcSinkAuthSpec(*authTypeEnum)
		cdcSinkAuthSpec.SetUsername(username)
		cdcSinkAuthSpec.SetPassword(password)

		createSinkRequest := ybmclient.NewCreateCdcSinkRequest(cdcSinkSpec, *cdcSinkAuthSpec)

		resp, r, err := apiClient.CdcApi.CreateCdcSink(context.Background(), accountID).CreateCdcSinkRequest(*createSinkRequest).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.CreateCdcSink``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var editCdcSinkCmd = &cobra.Command{
	Use:   "cdc_sink",
	Short: "Edit CDC Sink in YugabyteDB Managed",
	Long:  `Edit CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, _ := getApiRequestInfo("", "")

		cdcSinkName, _ := cmd.Flags().GetString("name")

		cdcSinkID, cdcSinkIDOk, _ := getCdcSinkID(context.Background(), apiClient, accountID, cdcSinkName)
		if !cdcSinkIDOk {
			fmt.Fprintf(os.Stderr, "No Cdc Sink named `%s` found\n", cdcSinkName)
			return
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

		resp, r, err := apiClient.CdcApi.EditCdcSink(context.Background(), accountID, cdcSinkID).EditCdcSinkRequest(*editCdcSinkRequest).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.EditCdcSink`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			return
		}

		prettyPrintJson(resp)
	},
}

var deleteCdcSinkCmd = &cobra.Command{
	Use:   "cdc_sink",
	Short: "Delete CDC Sink in YugabyteDB Managed",
	Long:  `Delete CDC Sink in YugabyteDB Managed`,
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, accountID, _ := getApiRequestInfo("", "")

		cdcSinkName, _ := cmd.Flags().GetString("name")
		cdcSinkID, cdcSinkIDOk, _ := getCdcSinkID(context.Background(), apiClient, accountID, cdcSinkName)
		if !cdcSinkIDOk {
			fmt.Fprintf(os.Stderr, "No Cdc Sink named `%s` found\n", cdcSinkName)
			return
		}

		resp, err := apiClient.CdcApi.DeleteCdcSink(context.Background(), accountID, cdcSinkID).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CdcApi.DeleteCdcSink`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", resp)
			return
		}
		fmt.Fprintf(os.Stdout, "CDC sink deleted successfully")
		prettyPrintJson(resp)
	},
}

func init() {
	getCmd.AddCommand(getCdcSinkCmd)
	getCdcSinkCmd.Flags().String("name", "", "Name of the CDC Sink")

	createCmd.AddCommand(createCdcSinkCmd)
	createCdcSinkCmd.Flags().String("name", "", "Name of the CDC sink")
	createCdcSinkCmd.Flags().String("cdc-sink-type", "", "Name of the CDC sink type")
	createCdcSinkCmd.Flags().String("auth-type", "", "Name of the CDC sink authentication type")
	createCdcSinkCmd.Flags().String("hostname", "", "Hostname of the CDC sink")
	createCdcSinkCmd.Flags().String("username", "", "Username of the CDC sink")
	createCdcSinkCmd.Flags().String("password", "", "Password of the CDC sink")

	updateCmd.AddCommand(editCdcSinkCmd)
	editCdcSinkCmd.Flags().String("name", "", "Name of the CDC Sink")
	editCdcSinkCmd.Flags().String("new-name", "", "Name of the new CDC Sink")
	editCdcSinkCmd.Flags().String("username", "", "Username of the CDC Sink")
	editCdcSinkCmd.Flags().String("password", "", "Password of the CDC Sink")

	deleteCmd.AddCommand(deleteCdcSinkCmd)
	deleteCdcSinkCmd.Flags().String("name", "", "Name of the CDC Sink")

}
