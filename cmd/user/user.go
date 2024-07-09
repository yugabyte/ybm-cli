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

package user

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  "Manage users in your YugabyteDB Aeon account",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Long:  `List users in your YugabyteDB Aeon account`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		userListRequest := authApi.ListAccountUsers()

		// if user filters by email, add it to the request
		email, _ := cmd.Flags().GetString("email")
		if email != "" {
			userListRequest = userListRequest.Email(email)
		}

		resp, r, err := userListRequest.Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		userCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewUserFormat(viper.GetString("output")),
		}

		if len(resp.GetData()) < 1 {
			logrus.Info("No users found")
			return
		}

		formatter.UserWrite(userCtx, resp.GetData())
	},
}

var inviteUserCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a user",
	Long:  "Invite a user to your YugabyteDB Aeon account",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		email, _ := cmd.Flags().GetString("email")
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
			err := util.ConfirmCommand(fmt.Sprintf(util.GetSensitivePermissionsConfirmationMessage()+" Are you sure you want to proceed with inviting user '%s' with this role?", roleName, email), viper.GetBool("force"))
			if err != nil {
				logrus.Fatal(err)
			}
		}

		usersSpec, err := authApi.CreateBatchInviteUserSpec(email, roleId)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.BatchInviteAccountUsers().BatchInviteUserSpec(*usersSpec).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if resp.Data.GetUserList()[0].GetIsSuccessful() {
			email := resp.Data.GetUserList()[0].GetInviteUserData().Spec.GetEmail()
			role := resp.Data.GetUserList()[0].GetInviteUserData().Info.GetRoleList()[0].GetRoles()[0].Info.GetDisplayName()
			fmt.Printf("The user %s has been successfully invited with role: %s.\n", formatter.Colorize(email, formatter.GREEN_COLOR), formatter.Colorize(role, formatter.GREEN_COLOR))
		} else {
			fmt.Printf("%s \n", resp.Data.GetUserList()[0].GetErrorMessage())
		}

	},
}

var updateUserCmd = &cobra.Command{
	Use:   "update",
	Short: "Modify role of a user",
	Long:  "Modify role of a user in your YugabyteDB Aeon account",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		email, _ := cmd.Flags().GetString("email")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to modify the role of user with %s: %s", "email", email), viper.GetBool("force"))
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

		email, _ := cmd.Flags().GetString("email")
		userId, err := authApi.GetUserIdByEmail(email)
		if err != nil {
			logrus.Infof("Could not find user with email: %s.\n", email)
			logrus.Fatal(err)
		}

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
			err := util.ConfirmCommand(fmt.Sprintf(util.GetSensitivePermissionsConfirmationMessage()+" Are you sure you want to proceed with modifying role of user '%s'?", roleName, email), viper.GetBool("force"))
			if err != nil {
				logrus.Fatal(err)
			}
		}

		modifyUserRoleRequest := *ybmclient.NewModifyUserRoleRequest(roleId)

		request := authApi.ModifyUserRole(userId)
		request = request.ModifyUserRoleRequest(modifyUserRoleRequest)

		r, err := request.Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("The role of user %s has been successfully modified.\n", formatter.Colorize(email, formatter.GREEN_COLOR))
	},
}

var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  "Delete a user from your YugabyteDB Aeon account",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		email, _ := cmd.Flags().GetString("email")
		err := util.ConfirmCommand(fmt.Sprintf("Are you sure you want to delete the user with %s: %s", "email", email), viper.GetBool("force"))
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

		email, _ := cmd.Flags().GetString("email")
		userId, err := authApi.GetUserIdByEmail(email)
		if err != nil {
			logrus.Infof("Could not find user with email: %s.\n", email)
			logrus.Fatal(err)
		}

		response, err := authApi.RemoveAccountUser(userId).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", response)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		fmt.Printf("The user %s has been successfully deleted.\n", formatter.Colorize(email, formatter.GREEN_COLOR))
	},
}

func init() {
	UserCmd.AddCommand(listUsersCmd)
	listUsersCmd.Flags().String("email", "", "[OPTIONAL] To filter by user email.")

	UserCmd.AddCommand(inviteUserCmd)
	inviteUserCmd.Flags().String("email", "", "[REQUIRED] The email of the user to be invited.")
	inviteUserCmd.MarkFlagRequired("email")
	inviteUserCmd.Flags().String("role-name", "", "[REQUIRED] The name of the role to be assigned to the user.")
	inviteUserCmd.MarkFlagRequired("role-name")
	inviteUserCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	UserCmd.AddCommand(updateUserCmd)
	updateUserCmd.Flags().String("email", "", "[REQUIRED] The email of the user whose role is to be modified.")
	updateUserCmd.MarkFlagRequired("email")
	updateUserCmd.Flags().String("role-name", "", "[REQUIRED] The name of the role to be assigned to the user.")
	updateUserCmd.MarkFlagRequired("role-name")
	updateUserCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

	UserCmd.AddCommand(deleteUserCmd)
	deleteUserCmd.Flags().String("email", "", "[REQUIRED] The email of the user.")
	deleteUserCmd.MarkFlagRequired("email")
	deleteUserCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")
}
