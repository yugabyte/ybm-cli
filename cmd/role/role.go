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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		roleListRequest := authApi.ListRbacRoles()
		// if user filters by display name, add it to the request
		// roleName, _ := cmd.Flags().GetString("role-name")
		// if roleName != "" {
		// 	roleistRequest = roleListRequest.DisplayName(roleName)
		// }

		resp, r, err := roleListRequest.Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		rolesCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewRoleFormat(viper.GetString("output")),
		}
		if len(resp.GetData()) < 1 {
			fmt.Println("No roles found")
			return
		}
		formatter.RoleWrite(rolesCtx, resp.GetData())
	},
}

var describeRoleCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a role",
	Long:  "Describe a role in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		// roleListRequest := authApi.ListRbacRoles()
		roleId, _ := cmd.Flags().GetString("role-name")

		resp, r, err := authApi.GetRole(roleId).Execute()

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

// createClusterCmd represents the cluster command
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
			// resource=CMK, operation=READ
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

		roleId, _ := cmd.Flags().GetString("role-name")

		resp, r, err := authApi.GetRole(roleId).Execute()

		roleName := resp.GetData().Info.DisplayName

		permissionsMap := map[string][]string{}

		permissionsList, _ := cmd.Flags().GetStringArray("permissions")
		for _, permissionsString := range permissionsList {
			// resource=CMK, operation=READ
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

		roleSpec, err := authApi.CreateRoleSpec(cmd, *roleName, permissionsMap)
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
	Short: "Delete a custom role in YugabyteDB Managed",
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		roleId, _ := cmd.Flags().GetString("role-name")
		// roleId, err := authApi.GetRoleIdByName(roleName)
		// if err != nil {
		// 	logrus.Fatal(err)
		// }
		
		response, err := authApi.DeleteRole(roleId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("The role %s has been successfully deleted.\n", formatter.Colorize(roleId, formatter.GREEN_COLOR))
	},
}

func init() {
	RoleCmd.AddCommand(deleteRoleCmd)

	deleteRoleCmd.Flags().String("role-name", "", "[REQUIRED] The name of the role to be deleted.")
	deleteRoleCmd.MarkFlagRequired("role-name")
	deleteRoleCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	RoleCmd.AddCommand(listRolesCmd)

	RoleCmd.AddCommand(describeRoleCmd)
	describeRoleCmd.Flags().String("role-name", "", "[REQUIRED] The name of the role to be deleted.")
	describeRoleCmd.MarkFlagRequired("role-name")

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
	// ./ybm cluster list
	// ./ybm cluster describe --cluster-name=adept-emu
	// ./ybm permissions list
	// ./ybm role list
	// ./ybm role create --role-name=aaa-testing-create-role-cli-4 --permissions resource-type=PROMO_CODE,operation-group=REDEEM
	// ./ybm role update --role-name=8426aa88-a0e2-42b9-9f20-59313f02fc74 --description="Some Description With Spaces" --permissions operation-group=CREATE,resource-type=BILLING_PROFILE --permissions operation-group=UPDATE,resource-type=BILLING_PROFILE --permissions resource-type=API_KEY,operation-group=REVOKE
	// ./ybm role describe --role-name=9dc63700-2982-4305-9378-cde72e2503b6
	// ./ybm role delete --role-name=<>
}

