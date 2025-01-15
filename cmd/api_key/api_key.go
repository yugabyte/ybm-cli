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
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var ApiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage API Keys",
	Long:  "Manage API Keys in your YugabyteDB Aeon account",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listApiKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List API Keys",
	Long:  `List API Keys in your YugabyteDB Aeon account`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		apiKeyListRequest := authApi.ListApiKeys()

		// if user filters by api key name, add it to the request
		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			apiKeyListRequest = apiKeyListRequest.ApiKeyName(name)
		}

		isNameSpecified := cmd.Flags().Changed("name")
		keyStatus := "ACTIVE"
		// If --name arg is specified, don't set default filter for key-status
		// because if an API key is revoked/expired and user do $ ybm api-key list --name <key-name>
		// it will lead to empty response if we filter key by ACTIVE status.
		if isNameSpecified {
			keyStatus = ""
		}

		// if user filters by key status, add it to the request
		if cmd.Flags().Changed("status") {
			keyStatus, _ = cmd.Flags().GetString("status")
		}
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

		apiKeyOutputList := *enrichApiKeyDataWithAllowListInfo(&resp.Data, authApi)
		formatter.ApiKeyWrite(apiKeyCtx, apiKeyOutputList)
	},
}

func enrichApiKeyDataWithAllowListInfo(apiKeys *[]ybmclient.ApiKeyData, authApi *ybmAuthClient.AuthApiClient) *[]formatter.ApiKeyDataAllowListInfo {
	apiKeyOutputList := make([]formatter.ApiKeyDataAllowListInfo, 0)

	// For each API key, fetch the allow list(s) associated with it
	for _, apiKey := range *apiKeys {
		apiKeyId := apiKey.GetInfo().Id
		allowListsNames := make([]string, 0)

		if util.IsFeatureFlagEnabled(util.API_KEY_ALLOW_LIST) {
			allowListIds := apiKey.GetSpec().AllowListInfo
			if allowListIds != nil && len(*allowListIds) > 0 {
				apiKeyAllowLists, resp, err := authApi.ListApiKeyNetworkAllowLists(apiKeyId).Execute()
				if err != nil {
					logrus.Debugf("Full HTTP response: %v", resp)
					logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
				}
				for _, allowList := range apiKeyAllowLists.GetData() {
					allowListsNames = append(allowListsNames, allowList.GetSpec().Name)
				}
			}
		}

		apiKeyOutputList = append(apiKeyOutputList, formatter.ApiKeyDataAllowListInfo{
			ApiKey:     &apiKey,
			AllowLists: allowListsNames,
		})
	}
	return &apiKeyOutputList
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

var createApiKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API Key",
	Long:  "Create an API Key",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
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

			roleContainsSensitivePermissions, err := authApi.RoleContainsSensitivePermissions(roleId)
			if err != nil {
				logrus.Fatal(err)
			}

			if roleContainsSensitivePermissions {
				viper.BindPFlag("force", cmd.Flags().Lookup("force"))
				err := util.ConfirmCommand(fmt.Sprintf(util.GetSensitivePermissionsConfirmationMessage()+" Are you sure you want to proceed with creating API Key '%s' with this role?", roleName, name), viper.GetBool("force"))
				if err != nil {
					logrus.Fatal(err)
				}
			}

			apiKeySpec.SetRoleId(roleId)
		}

		if util.IsFeatureFlagEnabled(util.API_KEY_ALLOW_LIST) && cmd.Flags().Changed("network-allow-lists") {
			allowLists, _ := cmd.Flags().GetString("network-allow-lists")

			allowListNames := strings.Split(allowLists, ",")
			allowListIds := make([]string, 0)

			for _, allowList := range allowListNames {
				allowListId, err := authApi.GetNetworkAllowListIdByName(strings.TrimSpace(allowList))
				if err != nil {
					logrus.Fatalln(err)
				}
				allowListIds = append(allowListIds, allowListId)
			}
			apiKeySpec.SetAllowListInfo(allowListIds)
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

		apiKeyOutput := *enrichApiKeyDataWithAllowListInfo(&[]ybmclient.ApiKeyData{resp.GetData()}, authApi)
		formatter.ApiKeyWrite(apiKeyCtx, apiKeyOutput)

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
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
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

	if util.IsFeatureFlagEnabled(util.API_KEY_ALLOW_LIST) {
		createApiKeyCmd.Flags().String("network-allow-lists", "", "[OPTIONAL] The network allow lists(comma separated names) to assign to the API key.")
	}
	createApiKeyCmd.Flags().String("role-name", "", "[OPTIONAL] The name of the role to be assigned to the API Key. If not provided, an Admin API Key will be generated.")
	createApiKeyCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	ApiKeyCmd.AddCommand(revokeApiKeyCmd)
	revokeApiKeyCmd.Flags().SortFlags = false
	revokeApiKeyCmd.Flags().String("name", "", "[REQUIRED] The name of the API Key.")
	revokeApiKeyCmd.MarkFlagRequired("name")
	revokeApiKeyCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
}
