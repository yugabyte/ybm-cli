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

package util

import (
	"errors"
	"fmt"
	"net"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func FindNetworkAllowList(nals []ybmclient.NetworkAllowListData, name string) (ybmclient.NetworkAllowListData, error) {
	for _, allowList := range nals {
		if allowList.Spec.Name == name {
			return allowList, nil
		}
	}
	return ybmclient.NetworkAllowListData{}, errors.New("Unable to find NetworkAllowList " + name)
}

func GetClusterTier(tierCli string) (string, error) {

	if tierCli == "Dedicated" {
		return "PAID", nil
	} else if tierCli == "Sandbox" {
		return "FREE", nil
	}

	return "", fmt.Errorf("The tier must be either 'Sandbox' or 'Dedicated'")
}

func ValidateCIDR(cidr string) (bool, error) {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("%s is not a valid CIDR", cidr)
	}
	return true, nil
}
