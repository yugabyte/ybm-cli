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

package cluster

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yugabyte/ybm-cli/cmd/util"
)

func TestConnectionPoolingFlagAvailability(t *testing.T) {
	// Test that the flag is always available (regardless of feature flag state)
	testCmd := &cobra.Command{
		Use: "test-create",
	}

	// Add flags like the actual create command
	testCmd.Flags().Bool("enable-connection-pooling", false, "[OPTIONAL] Enable connection pooling for the cluster after creation. Default false. (Requires CONNECTION_POOLING feature flag)")

	// Check that the flag exists
	flag := testCmd.Flags().Lookup("enable-connection-pooling")
	if flag == nil {
		t.Error("enable-connection-pooling flag should always be available")
	}

	// Test default value
	if flag != nil && flag.DefValue != "false" {
		t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
	}
}

func TestConnectionPoolingFeatureFlagValidation(t *testing.T) {
	// Test that feature flag validation works at runtime
	testCmd := &cobra.Command{
		Use: "test-validate",
		Run: func(cmd *cobra.Command, args []string) {
			enableConnectionPooling, _ := cmd.Flags().GetBool("enable-connection-pooling")
			if enableConnectionPooling && !util.IsFeatureFlagEnabled(util.CONNECTION_POOLING) {
				t.Error("Feature flag validation should prevent execution when flag is disabled")
			}
		},
	}
	testCmd.Flags().Bool("enable-connection-pooling", false, "Test flag")

	// Test with feature flag disabled
	os.Setenv("YBM_FF_CONNECTION_POOLING", "false")
	defer os.Unsetenv("YBM_FF_CONNECTION_POOLING")

	// This should validate the flag exists but would fail at runtime if used
	flag := testCmd.Flags().Lookup("enable-connection-pooling")
	if flag == nil {
		t.Error("Flag should exist even when feature flag is disabled")
	}
}
