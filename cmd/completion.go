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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

// AutocompleteClusterName - Autocomplete all container names.
func AutocompleteClusterName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	initConfig()
	clusterName := []string{}
	authApi, err := ybmAuthClient.NewAuthApiClient()
	if err != nil {
		cobra.CompErrorln(fmt.Sprintf("could not initiate api client: %s", err.Error()))
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	authApi.GetInfo("", "")

	resp, _, err := authApi.ListClusters().Execute()

	if err != nil {
		cobra.CompErrorln(ybmAuthClient.GetApiErrorDetails(err))
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(resp.GetData()) < 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	for _, data := range resp.GetData() {
		clusterName = append(clusterName, data.Spec.Name)
	}
	return clusterName, cobra.ShellCompDirectiveNoFileComp
}

func AutocompleteVPCName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	initConfig()
	vpcList := []string{}
	authApi, err := ybmAuthClient.NewAuthApiClient()
	if err != nil {
		cobra.CompErrorln(fmt.Sprintf("could not initiate api client: %s", err.Error()))
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	authApi.GetInfo("", "")

	vpcListRequest := authApi.ListSingleTenantVpcsByName("")

	resp, _, err := vpcListRequest.Execute()
	if err != nil {
		cobra.CompErrorln(ybmAuthClient.GetApiErrorDetails(err))
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(resp.GetData()) < 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	for _, data := range resp.GetData() {
		vpcList = append(vpcList, data.Spec.Name)
	}
	return vpcList, cobra.ShellCompDirectiveNoFileComp
}
