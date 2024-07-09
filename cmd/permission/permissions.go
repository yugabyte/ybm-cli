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

package permission

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var ResourcePermissionsCmd = &cobra.Command{
	Use:   "permission",
	Short: "View available permissions for roles",
	Long:  "View available permissions for your YugabyteDB Aeon roles",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listResourcePermissionsCmd = &cobra.Command{
	Use:   "list",
	Short: "List Available Permissions for Custom Roles",
	Long:  `List Available Permissions for Custom Roles`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		resourcePermissionsResp, resp, err := authApi.ListResourcePermissions().Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", resp)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		resourcePermissionData := resourcePermissionsResp.Data

		resourcePermissionCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewResourcePermissionFormat(viper.GetString("output")),
		}

		formatter.ResourcePermissionWrite(resourcePermissionCtx, resourcePermissionData)
	},
}

func init() {
	ResourcePermissionsCmd.AddCommand(listResourcePermissionsCmd)
}
