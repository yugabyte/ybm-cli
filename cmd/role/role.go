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
		roleName, _ := cmd.Flags().GetString("role-name")
		roleId, err := authApi.GetRoleIdByName(roleName)
		if err != nil {
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
	RoleCmd.AddCommand(deleteRoleCmd)

	deleteRoleCmd.Flags().String("role-name", "", "[REQUIRED] The name of the role to be deleted.")
	deleteRoleCmd.MarkFlagRequired("role-name")
	deleteRoleCmd.Flags().BoolP("force", "f", false, "Bypass the prompt for non-interactive usage")

}