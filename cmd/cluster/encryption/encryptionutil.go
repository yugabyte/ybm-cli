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
package ear

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	"golang.org/x/term"
)

func GetCmkSpecFromCommand(cmd *cobra.Command) (*ybmclient.CMKSpec, error) {
	var cmkSpec *ybmclient.CMKSpec = nil
	if cmd.Flags().Changed("encryption-spec") {
		cmkString, _ := cmd.Flags().GetString("encryption-spec")
		cmkProvider := ""
		cmkAwsSecretKey := ""
		cmkAwsAccessKey := ""
		cmkAwsArnList := []string{}

		for _, cmkInfo := range strings.Split(cmkString, ",") {
			kvp := strings.Split(cmkInfo, "=")
			if len(kvp) != 2 {
				logrus.Fatalln("Incorrect format in cmk spec")
			}
			key := kvp[0]
			val := kvp[1]
			switch key {
			case "cloud-provider":
				if len(strings.TrimSpace(val)) != 0 {
					cmkProvider = val
				}
			case "aws-secret-key":
				if len(strings.TrimSpace(val)) != 0 {
					cmkAwsSecretKey = val
				}
			case "aws-access-key":
				if len(strings.TrimSpace(val)) != 0 {
					cmkAwsAccessKey = val
				}
			case "aws-arn":
				if len(strings.TrimSpace(val)) != 0 {
					cmkAwsArnList = append(cmkAwsArnList, val)
				}
			}
		}

		if cmkProvider == "AWS" && cmkAwsAccessKey == "" {
			logrus.Fatalln("Incorrect format in cmk spec")
		}

		// The password/secret was not provided.
		if cmkProvider == "AWS" && cmkAwsSecretKey == "" {
			// We should first check the environment variables.
			envVarName := "YBM_AWS_SECRET_KEY"
			value, exists := os.LookupEnv(envVarName)
			if exists {
				cmkAwsSecretKey = value
			} else {
				// If not found, prompt the user.
				fmt.Print("Please provide the AWS Secret Key: ")

				data, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					logrus.Fatalln("Could not read apiKey: ", err)
				}
				cmkAwsSecretKey = string(data)

				// Validate non-empty key
				if strings.TrimSpace(cmkAwsSecretKey) == "" {
					logrus.Fatalln("The AWS Secret Key cannot be empty")
				}
			}

		}

		cmkSpec = ybmclient.NewCMKSpec(cmkProvider)
		if cmkProvider == "AWS" {
			cmkSpec.AwsCmkSpec.Set(ybmclient.NewAWSCMKSpec(cmkAwsAccessKey, cmkAwsSecretKey, cmkAwsArnList))
		}
	}

	return cmkSpec, nil
}
