// Copyright (c) YugaByte, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022- Yugabyte, Inc.

package util

import (
	"errors"
	"fmt"

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
