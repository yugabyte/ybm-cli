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
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/golang-jwt/jwt/v5"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const customRoleFeatureFlagDisabled string = "Requested API not found"
const sensitivePermissionsConfirmationMessage string = "Some of the permissions assigned to role '%s' have security implications (such as user, API key, and role management operations)."

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

	return "", fmt.Errorf("the tier must be either 'Sandbox' or 'Dedicated'")
}

func SetPreferredRegion(clusterRegionInfo []ybmclient.ClusterRegionInfo, preferredRegion string) error {

	regionFound := false
	for _, info := range clusterRegionInfo {
		if info.PlacementInfo.CloudInfo.GetRegion() == preferredRegion {
			regionFound = true
		}
	}

	if !regionFound {
		return fmt.Errorf("the preferred region is not found in the list of regions")
	}

	for i, info := range clusterRegionInfo {
		if info.PlacementInfo.CloudInfo.GetRegion() == preferredRegion {
			clusterRegionInfo[i].SetIsAffinitized(true)
		} else {
			clusterRegionInfo[i].SetIsAffinitized(false)
		}
	}

	return nil

}

func SetDefaultRegion(clusterRegionInfo []ybmclient.ClusterRegionInfo, defaultRegion string) error {

	regionFound := false
	for _, info := range clusterRegionInfo {
		if info.PlacementInfo.CloudInfo.GetRegion() == defaultRegion {
			regionFound = true
		}
	}

	if !regionFound {
		return fmt.Errorf("the default region is not found in the list of regions")
	}

	for i, info := range clusterRegionInfo {
		if info.PlacementInfo.CloudInfo.GetRegion() == defaultRegion {
			clusterRegionInfo[i].SetIsDefault(true)
			break
		}
	}

	return nil

}

func ValidateNumFaultsToTolerate(numFaultsToTolerate int32, faultTolerance ybmclient.ClusterFaultTolerance) (bool, error) {
	if numFaultsToTolerate < 0 || numFaultsToTolerate > 3 {
		return false, fmt.Errorf("number of faults to tolerate must be between 0 and 3")
	}
	if faultTolerance == ybmclient.CLUSTERFAULTTOLERANCE_NONE && numFaultsToTolerate != 0 {
		return false, fmt.Errorf("number of faults to tolerate must be 0 for fault tolerance level 'NONE'")
	}
	if faultTolerance == ybmclient.CLUSTERFAULTTOLERANCE_NODE && numFaultsToTolerate < 1 {
		return false, fmt.Errorf("number of faults to tolerate must be greater than 0 for fault tolerance level 'NODE'")
	}
	if faultTolerance == ybmclient.CLUSTERFAULTTOLERANCE_REGION && numFaultsToTolerate < 1 {
		return false, fmt.Errorf("number of faults to tolerate must be greater than 0 for fault tolerance level 'REGION'")
	}
	if faultTolerance == ybmclient.CLUSTERFAULTTOLERANCE_ZONE && numFaultsToTolerate != 1 {
		return false, fmt.Errorf("number of faults to tolerate must be 1 for fault tolerance level 'ZONE'")
	}

	return true, nil
}

func ValidateCIDR(cidr string) (bool, error) {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("%s is not a valid CIDR", cidr)
	}
	return true, nil
}

func ExtractJwtClaims(tokenStr string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok {
		return token.Claims.(jwt.MapClaims), nil
	}
	return nil, errors.New("unable to extract claims from token")
}

func IsJwtTokenExpiredWithTime(tokenStr string, now time.Time) (bool, error) {
	claims, err := ExtractJwtClaims(tokenStr)
	if err != nil {
		return false, err
	}

	exp := claims["exp"].(float64)
	if exp < float64(now.Unix()) {
		return true, nil
	}

	iat := claims["iat"].(float64)
	if iat > float64(now.Unix()) {
		return true, nil
	}
	return false, nil
}

func IsJwtTokenExpired(tokenStr string) (bool, error) {
	return IsJwtTokenExpiredWithTime(tokenStr, time.Now())
}

// Inspired from here:
// https://stackoverflow.com/questions/37562873/most-idiomatic-way-to-select-elements-from-an-array-in-golang
// This allows us to filter a slice of any type using a function that returns a bool
func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func SplitAndIgnoreEmpty(str string, sep string) []string {
	split := Filter(strings.Split(str, sep), func(s string) bool {
		return s != ""
	})
	// If the string is empty, we want to return an empty slice
	if split == nil {
		return []string{}
	}
	return split
}

// this function will add an interactive comfirmation with the message provided
func ConfirmCommand(message string, bypass bool) error {
	errAborted := errors.New("command aborted")
	if bypass {
		return nil
	}
	response := false
	prompt := &survey.Confirm{
		Message: message,
	}
	err := survey.AskOne(prompt, &response)
	if err != nil {
		return err
	}
	if !response {
		return errAborted
	}
	return nil
}

func GetCustomRoleFeatureFlagDisabledError() string {
	return customRoleFeatureFlagDisabled
}

func GetSensitivePermissionsConfirmationMessage() string {
	return sensitivePermissionsConfirmationMessage
}
