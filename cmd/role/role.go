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
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// createRoleCmd represents the create role command
var createRoleCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom role",
	Long:  "Create a custom role in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		roleName, _ := cmd.Flags().GetString("role-name")

		permissionsMap := map[string][]string{}

		permissionsList, _ := cmd.Flags().GetStringArray("permissions")
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

			if(kvpOne[0]=="resource-type" && kvpTwo[0]=="operation-group"){
				if ops, ok := permissionsMap[kvpOne[1]]; ok {
					ops = append(ops, kvpTwo[1])
					permissionsMap[kvpOne[1]] = ops
				} else {
					permissionsMap[kvpOne[1]] = []string{kvpTwo[1]}
				}
			} else if (kvpTwo[0]=="resource-type" && kvpOne[0]=="operation-group") {
				if ops, ok := permissionsMap[kvpTwo[1]]; ok {
					ops = append(ops, kvpOne[1])
					permissionsMap[kvpTwo[1]] = ops
				} else {
					permissionsMap[kvpTwo[1]] = []string{kvpOne[1]}
				}
			} else {
				logrus.Fatalln("Resource type and Operation group must be specified in permissions")
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

var updateRoleCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a custom role",
	Long:  "Update a custom role in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		roleName, _ := cmd.Flags().GetString("role-name")
		roleId, err := authApi.GetRoleIdByName(roleName)
		if err != nil {
			logrus.Fatal(err)
		}

		permissionsMap := map[string][]string{}

		permissionsList, _ := cmd.Flags().GetStringArray("permissions")
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

			if(kvpOne[0]=="resource-type" && kvpTwo[0]=="operation-group"){
				if ops, ok := permissionsMap[kvpOne[1]]; ok {
					ops = append(ops, kvpTwo[1])
					permissionsMap[kvpOne[1]] = ops
				} else {
					permissionsMap[kvpOne[1]] = []string{kvpTwo[1]}
				}
			} else if (kvpTwo[0]=="resource-type" && kvpOne[0]=="operation-group") {
				if ops, ok := permissionsMap[kvpTwo[1]]; ok {
					ops = append(ops, kvpOne[1])
					permissionsMap[kvpTwo[1]] = ops
				} else {
					permissionsMap[kvpTwo[1]] = []string{kvpOne[1]}
				}
			} else {
				logrus.Fatalln("Resource type and Operation group must be specified in permissions")
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


func init() {
	RoleCmd.AddCommand(createRoleCmd)

	createRoleCmd.Flags().String("role-name", "", "[REQUIRED] Name of the role to be created.")
	createRoleCmd.MarkFlagRequired("role-name")
	createRoleCmd.Flags().String("description", "", "[OPTIONAL] Description of the role to be created.")
	createRoleCmd.Flags().StringArray("permissions", []string{}, `[REQUIRED] Permissions for the role. Please provide key value pairs resource-type=<resource-type>,operation-group=<operation-group> as the value. Both resource-type and operation-group are mandatory. Information about multiple permissions can be specified by using multiple --permissions arguments.`)
	createRoleCmd.MarkFlagRequired("permissions")

	RoleCmd.AddCommand(updateRoleCmd)

	updateRoleCmd.Flags().String("role-name", "", "[REQUIRED] Name of the role.")
	updateRoleCmd.MarkFlagRequired("role-name")
	updateRoleCmd.Flags().String("new-name", "", "[OPTIONAL] New name of the role to be updated.")
	updateRoleCmd.Flags().String("description", "", "[OPTIONAL] New description of the role to be updated.")
	updateRoleCmd.Flags().StringArray("permissions", []string{}, `[REQUIRED] Permissions for the role. Please provide key value pairs resource-type=<resource-type>,operation-group=<operation-group> as the value. Both resource-type and operation-group are mandatory. Information about multiple permissions can be specified by using multiple --permissions arguments.`)
	updateRoleCmd.MarkFlagRequired("permissions")
}
