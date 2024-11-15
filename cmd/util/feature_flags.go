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
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type FeatureFlag string

const (
	CDC                     FeatureFlag = "CDC"
	CONFIGURE_URL           FeatureFlag = "CONFIGURE_URL"
	NODE_OP                 FeatureFlag = "NODE_OPS"
	TOOLS                   FeatureFlag = "TOOLS"
	AZURE_CIDR_ALLOWED      FeatureFlag = "AZURE_CIDR_ALLOWED"
	ENTERPRISE_SECURITY     FeatureFlag = "ENTERPRISE_SECURITY"
	DB_AUDIT_LOGGING        FeatureFlag = "DB_AUDIT_LOGGING"
	PITR_CONFIG             FeatureFlag = "PITR_CONFIG"
	PITR_RESTORE            FeatureFlag = "PITR_RESTORE"
	CONNECTION_POOLING      FeatureFlag = "CONNECTION_POOLING"
	GOOGLECLOUD_INTEGRATION FeatureFlag = "GOOGLECLOUD_INTEGRATION"
	DR                      FeatureFlag = "DR"
	DB_QUERY_LOGS           FeatureFlag = "DB_QUERY_LOGGING"
)

func (f FeatureFlag) String() string {
	return string(f)
}

func IsFeatureFlagEnabled(featureFlag FeatureFlag) bool {
	envVarName := "YBM_FF_" + featureFlag.String()
	return strings.ToLower(os.Getenv(envVarName)) == "true"
}

func AddCommandIfFeatureFlag(rootCmd *cobra.Command, cmd *cobra.Command, featureFlag FeatureFlag) {
	// If the feature flag is enabled, add the command to the root command
	if IsFeatureFlagEnabled(featureFlag) {
		rootCmd.AddCommand(cmd)
	}
}
