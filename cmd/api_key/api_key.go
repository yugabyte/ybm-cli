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

package api_key

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var ApiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage API Keys",
	Long:  "Manage API Keys in your YBM account",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listApiKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List API Keys",
	Long:  `List API Keys in your YBM account`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		apiKeyListRequest := authApi.ListApiKeys()

		// if user filters by api key name, add it to the request
		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			apiKeyListRequest = apiKeyListRequest.ApiKeyName(name)
		}

		// if user filters by key status, add it to the request
		keyStatus, _ := cmd.Flags().GetString("status")
		if keyStatus != "" {
			validStatus := false
			for _, v := range GetKeyStatusFilters() {
				if strings.ToUpper(keyStatus) == v {
					validStatus = true
					apiKeyListRequest = apiKeyListRequest.Status([]string{v})
				}
			}
			if !validStatus {
				logrus.Fatalln("Only ACTIVE, EXPIRED, REVOKED status filters are allowed.")
			}
		}

		resp, r, err := apiKeyListRequest.Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		apiKeyCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewApiKeyFormat(viper.GetString("output")),
		}

		if len(resp.GetData()) < 1 {
			logrus.Info("No API Keys found")
			return
		}

		formatter.ApiKeyWrite(apiKeyCtx, resp.GetData())
	},
}

func GetKeyStatusFilters() []string {
	return []string{"ACTIVE", "EXPIRED", "REVOKED"}
}

func GetKeyTimeConversionMap() map[string]int {
	return map[string]int{
		"HOURS":  1,
		"DAYS":   24,
		"MONTHS": 24 * 30,
	}
}

// inviteUsersCmd represents the create role command
var createApiKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API Key",
	Long:  "Create an API Key",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		name, _ := cmd.Flags().GetString("name")
		duration, _ := cmd.Flags().GetInt32("duration")

		unit, _ := cmd.Flags().GetString("unit")
		var expiryHours int
		validTimeUnit := false
		for k, v := range GetKeyTimeConversionMap() {
			if strings.ToUpper(unit) == k {
				validTimeUnit = true
				expiryHours = int(duration) * v
			}
		}
		if !validTimeUnit {
			logrus.Fatalln("Only Hours, Days, and Months time units are allowed.")
		}

		apiKeySpec, err := authApi.CreateApiKeySpec(name, expiryHours)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			apiKeySpec.SetDescription(description)
		}

		if cmd.Flags().Changed("role-name") {
			roleName, _ := cmd.Flags().GetString("role-name")
			roleId, err := authApi.GetRoleIdByName(roleName)
			if err != nil {
				logrus.Fatal(err)
			}
			apiKeySpec.SetRoleId(roleId)
		}

		resp, r, err := authApi.CreateApiKey().ApiKeySpec(*apiKeySpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		apiKeyCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewApiKeyFormat(viper.GetString("output")),
		}

		formatter.SingleApiKeyWrite(apiKeyCtx, resp.GetData())

		fmt.Printf("\nAPI Key: %s \n", formatter.Colorize(resp.GetJwt(), formatter.GREEN_COLOR))
		fmt.Printf("\nThe API key is only shown once after creation. Copy and store it securely.\n")
	},
}

var revokeApiKeyCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke an API Key",
	Long:  "Revoke an API Key",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		name, _ := cmd.Flags().GetString("name")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to revoke the %s: %s", "API Key", name), viper.GetBool("force"))
		if err != nil {
			logrus.Fatal(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		name, _ := cmd.Flags().GetString("name")
		keyId, err := authApi.GetKeyIdByName(name)
		if err != nil {
			logrus.Fatal(err)
		}

		response, err := authApi.RevokeApiKey(keyId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("The API key %s has been successfully revoked.\n", formatter.Colorize(name, formatter.GREEN_COLOR))
	},
}

func init() {
	ApiKeyCmd.AddCommand(listApiKeysCmd)
	listApiKeysCmd.Flags().SortFlags = false
	listApiKeysCmd.Flags().String("name", "", "[OPTIONAL] To filter by API Key name.")
	listApiKeysCmd.Flags().String("status", "", "[OPTIONAL] To filter by API Key status. Available options are ACTIVE, EXPIRED, REVOKED.")

	ApiKeyCmd.AddCommand(createApiKeyCmd)
	createApiKeyCmd.Flags().SortFlags = false
	createApiKeyCmd.Flags().String("name", "", "[REQUIRED] The name of the API Key.")
	createApiKeyCmd.MarkFlagRequired("name")
	createApiKeyCmd.Flags().Int32("duration", 0, "[REQUIRED] The duration for which the API Key will be valid. 0 denotes that the key will never expire.")
	createApiKeyCmd.MarkFlagRequired("duration")
	createApiKeyCmd.Flags().String("unit", "", "[REQUIRED] The time units for which the API Key will be valid. Available options are Hours, Days, and Months.")
	createApiKeyCmd.MarkFlagRequired("unit")
	createApiKeyCmd.Flags().String("description", "", "[OPTIONAL] Description of the API Key to be created.")
	createApiKeyCmd.Flags().String("role-name", "", "[OPTIONAL] The name of the role to be assigned to the API Key. If not provided, an Admin API Key will be generated.")

	ApiKeyCmd.AddCommand(revokeApiKeyCmd)
	revokeApiKeyCmd.Flags().SortFlags = false
	revokeApiKeyCmd.Flags().String("name", "", "[REQUIRED] The name of the API Key.")
	revokeApiKeyCmd.MarkFlagRequired("name")
	revokeApiKeyCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
}
