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
		roleName, _ := cmd.Flags().GetString("role-name")
		if roleName != "" {
			roleListRequest = roleListRequest.DisplayName(roleName)
		}

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

func init() {
	RoleCmd.AddCommand(listRolesCmd)
	listRolesCmd.Flags().String("role-name", "", "[OPTIONAL] To filter by role name.")

}
