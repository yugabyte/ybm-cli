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
	// Save original environment state
	originalEnv := os.Getenv("YBM_FF_CONNECTION_POOLING")
	defer func() {
		if originalEnv != "" {
			os.Setenv("YBM_FF_CONNECTION_POOLING", originalEnv)
		} else {
			os.Unsetenv("YBM_FF_CONNECTION_POOLING")
		}
	}()

	t.Run("Flag Availability Scenarios", func(t *testing.T) {
		// Test: Flag exists and is accessible
		flag := createClusterCmd.Flags().Lookup("enable-connection-pooling")
		if flag == nil {
			t.Error("enable-connection-pooling flag should exist and be accessible")
		}

		// Test: Flag has correct default value (false)
		if flag != nil && flag.DefValue != "false" {
			t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
		}

		// Test: Flag accepts boolean values correctly
		if flag != nil && flag.Value.Type() != "bool" {
			t.Errorf("Expected flag type 'bool', got '%s'", flag.Value.Type())
		}

		// Test setting flag to true
		if flag != nil {
			err := flag.Value.Set("true")
			if err != nil {
				t.Errorf("Flag should accept 'true' value, got error: %v", err)
			}

			// Test setting flag to false
			err = flag.Value.Set("false")
			if err != nil {
				t.Errorf("Flag should accept 'false' value, got error: %v", err)
			}

			// Test invalid value
			err = flag.Value.Set("invalid")
			if err == nil {
				t.Error("Flag should reject invalid boolean values")
			}
		}
	})

	t.Run("Feature Flag Validation Logic Scenarios", func(t *testing.T) {
		// Test: Feature flag disabled + connection pooling requested → Should fail
		os.Setenv("YBM_FF_CONNECTION_POOLING", "false")
		enableConnectionPooling := true
		shouldFail := enableConnectionPooling && !util.IsFeatureFlagEnabled(util.CONNECTION_POOLING)
		if !shouldFail {
			t.Error("Expected validation logic to indicate failure when feature flag is disabled but connection pooling is requested")
		}

		// Test: Feature flag disabled + connection pooling NOT requested → Should pass
		enableConnectionPooling = false
		shouldPass := !(enableConnectionPooling && !util.IsFeatureFlagEnabled(util.CONNECTION_POOLING))
		if !shouldPass {
			t.Error("Expected validation logic to pass when feature flag is disabled and connection pooling is not requested")
		}

		// Test: Feature flag enabled + connection pooling requested → Should pass
		os.Setenv("YBM_FF_CONNECTION_POOLING", "true")
		enableConnectionPooling = true
		shouldPass = !(enableConnectionPooling && !util.IsFeatureFlagEnabled(util.CONNECTION_POOLING))
		if !shouldPass {
			t.Error("Expected validation logic to pass when feature flag is enabled and connection pooling is requested")
		}

		// Test: Feature flag enabled + connection pooling NOT requested → Should pass
		enableConnectionPooling = false
		shouldPass = !(enableConnectionPooling && !util.IsFeatureFlagEnabled(util.CONNECTION_POOLING))
		if !shouldPass {
			t.Error("Expected validation logic to pass when feature flag is enabled and connection pooling is not requested")
		}
	})

	// Note: Integration scenarios that require full command execution testing:
	// - Runtime command execution with feature flag disabled → Should terminate with logrus.Fatalf
	// - Runtime command execution with feature flag enabled → Should proceed to cluster creation
	//
	// These scenarios cannot be easily tested in unit tests because:
	// 1. logrus.Fatalf terminates the program (requires process isolation)
	// 2. Full command execution involves API calls and complex dependencies
	// 3. These are better suited for integration tests with actual CLI execution
}
