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
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type FeatureFlag string

const (
	CDC FeatureFlag = "CDC"
)

func (f FeatureFlag) String() string {
	return string(f)
}

func AddCommandIfFeatureFlag(rootCmd *cobra.Command, cmd *cobra.Command, featureFlag FeatureFlag) {
	envVarName := "YBM_FF_" + featureFlag.String()
	if strings.ToLower(os.Getenv(envVarName)) == "true" {
		rootCmd.AddCommand(cmd)
	}
}
