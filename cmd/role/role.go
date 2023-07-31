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

package role

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

var RoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage roles",
	Long:  "Manage roles",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listRolesCmd = &cobra.Command{
	Use:   "list",
	Short: "List roles",
	Long:  "List roles in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		roleListRequest := authApi.ListAllRbacRoles()

		// if user filters by role type, add it to the request
		roleType, _ := cmd.Flags().GetString("type")
		if roleType != "" {
			validType := false
			for k, v := range GetRoleTypeFilterMap() {
				if strings.ToUpper(roleType) == k {
					validType = true
					roleListRequest = roleListRequest.RoleTypes(v)
				}
			}
			if !validType {
				logrus.Fatalln("Only BUILT-IN and CUSTOM role types filters are allowed.")
			}
		}

		// if user filters by display name, add it to the request
		roleName, _ := cmd.Flags().GetString("role-name")
		if roleName != "" {
			roleListRequest = roleListRequest.DisplayName(roleName)
		}

		roleResponse, roleResp, roleErr := roleListRequest.Execute()

		if roleErr != nil {
			if strings.TrimSpace(ybmAuthClient.GetApiErrorDetails(roleErr)) == strings.TrimSpace(util.GetCustomRoleFeatureFlagDisabledError()) {
				systemRoleListRequest := authApi.ListSystemRbacRoles()
				if roleName != "" {
					systemRoleListRequest = systemRoleListRequest.DisplayName(roleName)
				}
				systemRoleResponse, systemRoleResp, systemRoleErr := systemRoleListRequest.Execute()

				if systemRoleErr != nil {
					logrus.Debugf("Full HTTP response: %v", systemRoleResp)
					logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(systemRoleErr))
				} else {
					roleResponse = systemRoleResponse
				}
			} else {
				logrus.Debugf("Full HTTP response: %v", roleResp)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(roleErr))
			}
		}

		rolesCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewRoleFormat(viper.GetString("output")),
		}
		if len(roleResponse.GetData()) < 1 {
			fmt.Println("No roles found")
			return
		}
		formatter.RoleWrite(rolesCtx, roleResponse.GetData())
	},
}

func GetRoleTypeFilterMap() map[string]string {
	return map[string]string{
		"BUILT-IN": "SYSTEM",
		"CUSTOM":   "USER_DEFINED",
	}
}

var describeRoleCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a role",
	Long:  "Describe a role in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		roleListRequest := authApi.ListAllRbacRolesWithPermissions()
		roleName, _ := cmd.Flags().GetString("role-name")
		roleListRequest = roleListRequest.DisplayName(roleName)

		roleResponse, roleResp, roleErr := roleListRequest.Execute()

		if roleErr != nil {
			if strings.TrimSpace(ybmAuthClient.GetApiErrorDetails(roleErr)) == strings.TrimSpace(util.GetCustomRoleFeatureFlagDisabledError()) {
				systemRoleListRequest := authApi.ListSystemRbacRolesWithPermissions()
				systemRoleListRequest = systemRoleListRequest.DisplayName(roleName)
				systemRoleResponse, systemRoleResp, systemRoleErr := systemRoleListRequest.Execute()

				if systemRoleErr != nil {
					logrus.Debugf("Full HTTP response: %v", systemRoleResp)
					logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(systemRoleErr))
				} else {
					roleResponse = systemRoleResponse
				}
			} else {
				logrus.Debugf("Full HTTP response: %v", roleResp)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(roleErr))
			}
		}

		if len(roleResponse.GetData()) < 1 {
			fmt.Println("No role found")
			return
		}

		if viper.GetString("output") == "table" {
			fullRoleContext := *formatter.NewFullRoleContext()
			fullRoleContext.Output = os.Stdout
			fullRoleContext.Format = formatter.NewFullRoleFormat(viper.GetString("output"))
			fullRoleContext.SetFullRole(roleResponse.GetData()[0])
			fullRoleContext.Write()
			return
		}

		rolesCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewFullRoleFormat(viper.GetString("output")),
		}

		formatter.SingleRoleWrite(rolesCtx, roleResponse.GetData()[0])
	},
}

// createRoleCmd represents the create role command
var createRoleCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom role",
	Long:  "Create a custom role in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		roleName, _ := cmd.Flags().GetString("role-name")

		permissionsMap := map[string][]string{}

		permissionsList, _ := cmd.Flags().GetStringArray("permissions")
		CreatePermissionsMap(permissionsList, permissionsMap)

		containsSensitivePermissions, err := authApi.ContainsSensitivePermissions(permissionsMap)
		if err != nil {
			logrus.Fatal(err)
		}
		if containsSensitivePermissions {
			viper.BindPFlag("force", cmd.Flags().Lookup("force"))
			err := util.ConfirmCommand(fmt.Sprintf(util.GetSensitivePermissionsConfirmationMessage()+" Are you sure you want to proceed with role creation?", roleName), viper.GetBool("force"))
			if err != nil {
				logrus.Fatal(err)
			}
		}

		roleSpec, err := authApi.CreateRoleSpec(cmd, roleName, permissionsMap)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.CreateRole().RoleSpec(*roleSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if viper.GetString("output") == "table" {
			fullRoleContext := *formatter.NewFullRoleContext()
			fullRoleContext.Output = os.Stdout
			fullRoleContext.Format = formatter.NewFullRoleFormat(viper.GetString("output"))
			fullRoleContext.SetFullRole(resp.GetData())
			fullRoleContext.Write()
			return
		}

		rolesCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewFullRoleFormat(viper.GetString("output")),
		}

		formatter.SingleRoleWrite(rolesCtx, resp.GetData())
	},
}

func PopulatePermissionsMap(resourceType string, operationGroup string, permissionsMap map[string][]string) {
	if ops, ok := permissionsMap[resourceType]; ok {
		ops = append(ops, operationGroup)
		permissionsMap[resourceType] = ops
	} else {
		permissionsMap[resourceType] = []string{operationGroup}
	}
}

func CreatePermissionsMap(permissionsList []string, permissionsMap map[string][]string) {
	for _, permissionsString := range permissionsList {
		permission := strings.Split(permissionsString, ",")

		if len(permission) != 2 {
			logrus.Fatalln("All necessary fields not provided for permissions")
		}

		kvpOne := strings.Split(permission[0], "=")
		if len(kvpOne) != 2 {
			logrus.Fatalln("Incorrect format in permissions fields")
		}

		kvpTwo := strings.Split(permission[1], "=")
		if len(kvpTwo) != 2 {
			logrus.Fatalln("Incorrect format in permissions fields")
		}

		if kvpOne[0] == "resource-type" && kvpTwo[0] == "operation-group" {
			PopulatePermissionsMap(kvpOne[1], kvpTwo[1], permissionsMap)
		} else if kvpTwo[0] == "resource-type" && kvpOne[0] == "operation-group" {
			PopulatePermissionsMap(kvpTwo[1], kvpOne[1], permissionsMap)
		} else {
			logrus.Fatalln("Resource type and Operation group must be specified in permissions")
		}
	}
}

var updateRoleCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a custom role",
	Long:  "Update a custom role in YugabyteDB Managed",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		roleName, _ := cmd.Flags().GetString("role-name")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to modify %s: %s. Editing a role with assigned users or API Keys can have major repercussions on automation as well as user’s ability to access resources.", "role", roleName), viper.GetBool("force"))
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

		roleName, _ := cmd.Flags().GetString("role-name")
		roleId, err := authApi.GetRoleIdByName(roleName)
		if err != nil {
			logrus.Infof("Could not find role with name: %s.\n", roleName)
			logrus.Fatal(err)
		}

		permissionsMap := map[string][]string{}

		permissionsList, _ := cmd.Flags().GetStringArray("permissions")
		CreatePermissionsMap(permissionsList, permissionsMap)

		containsSensitivePermissions, err := authApi.ContainsSensitivePermissions(permissionsMap)
		if err != nil {
			logrus.Fatal(err)
		}
		if containsSensitivePermissions {
			err := util.ConfirmCommand(fmt.Sprintf(util.GetSensitivePermissionsConfirmationMessage()+" Are you sure you want to proceed with updating this role?", roleName), viper.GetBool("force"))
			if err != nil {
				logrus.Fatal(err)
			}

		}

		roleSpec, err := authApi.CreateRoleSpec(cmd, roleName, permissionsMap)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if cmd.Flags().Lookup("new-name").Changed {
			newName, _ := cmd.Flags().GetString("new-name")
			roleSpec.SetName(newName)
		}

		updatedResp, r, err := authApi.UpdateRole(roleId).RoleSpec(*roleSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if viper.GetString("output") == "table" {
			fullRoleContext := *formatter.NewFullRoleContext()
			fullRoleContext.Output = os.Stdout
			fullRoleContext.Format = formatter.NewFullRoleFormat(viper.GetString("output"))
			fullRoleContext.SetFullRole(updatedResp.GetData())
			fullRoleContext.Write()
			return
		}

		rolesCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewFullRoleFormat(viper.GetString("output")),
		}

		formatter.SingleRoleWrite(rolesCtx, updatedResp.GetData())
	},
}

var deleteRoleCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a custom role",
	Long:  "Delete a custom role in YugabyteDB Managed",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		roleName, _ := cmd.Flags().GetString("role-name")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to delete %s: %s", "role", roleName), viper.GetBool("force"))
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
		roleName, _ := cmd.Flags().GetString("role-name")
		roleId, err := authApi.GetRoleIdByName(roleName)
		if err != nil {
			logrus.Infof("Could not find role with name: %s.\n", roleName)
			logrus.Fatal(err)
		}

		response, err := authApi.DeleteRole(roleId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("The role %s has been successfully deleted.\n", formatter.Colorize(roleName, formatter.GREEN_COLOR))
	},
}

func init() {
	RoleCmd.AddCommand(listRolesCmd)
	listRolesCmd.Flags().SortFlags = false
	listRolesCmd.Flags().String("role-name", "", "[OPTIONAL] To filter by role name.")
	listRolesCmd.Flags().String("type", "", "[OPTIONAL] To filter by role type. BUILT-IN and CUSTOM options are available to list only built-in or custom roles.")

	RoleCmd.AddCommand(describeRoleCmd)
	describeRoleCmd.Flags().SortFlags = false
	describeRoleCmd.Flags().String("role-name", "", "[REQUIRED] The name of the role.")
	describeRoleCmd.MarkFlagRequired("role-name")

	RoleCmd.AddCommand(createRoleCmd)
	createRoleCmd.Flags().SortFlags = false
	createRoleCmd.Flags().String("role-name", "", "[REQUIRED] Name of the role to be created.")
	createRoleCmd.MarkFlagRequired("role-name")
	createRoleCmd.Flags().StringArray("permissions", []string{}, `[REQUIRED] Permissions for the role. Please provide key value pairs resource-type=<resource-type>,operation-group=<operation-group> as the value. Both resource-type and operation-group are mandatory. Information about multiple permissions can be specified by using multiple --permissions arguments.`)
	createRoleCmd.MarkFlagRequired("permissions")
	createRoleCmd.Flags().String("description", "", "[OPTIONAL] Description of the role to be created.")
	createRoleCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	RoleCmd.AddCommand(updateRoleCmd)
	updateRoleCmd.Flags().SortFlags = false
	updateRoleCmd.Flags().String("role-name", "", "[REQUIRED] Name of the role.")
	updateRoleCmd.MarkFlagRequired("role-name")
	updateRoleCmd.Flags().StringArray("permissions", []string{}, `[REQUIRED] Permissions for the role. Please provide key value pairs resource-type=<resource-type>,operation-group=<operation-group> as the value. Both resource-type and operation-group are mandatory. Information about multiple permissions can be specified by using multiple --permissions arguments.`)
	updateRoleCmd.MarkFlagRequired("permissions")
	updateRoleCmd.Flags().String("description", "", "[OPTIONAL] New description of the role to be updated.")
	updateRoleCmd.Flags().String("new-name", "", "[OPTIONAL] New name of the role to be updated.")
	updateRoleCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	RoleCmd.AddCommand(deleteRoleCmd)
	deleteRoleCmd.Flags().SortFlags = false
	deleteRoleCmd.Flags().String("role-name", "", "[REQUIRED] The name of the role to be deleted.")
	deleteRoleCmd.MarkFlagRequired("role-name")
	deleteRoleCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

}
