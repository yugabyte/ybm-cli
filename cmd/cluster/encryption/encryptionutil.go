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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	"golang.org/x/term"
)

func parseKMSKeyResourceID(resourceID string) (keyRingName string, keyName string, location string, protectionLevel string) {
	parts := strings.Split(resourceID, "/")
	if len(parts) != 8 || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "keyRings" || parts[6] != "cryptoKeys" {
		logrus.Fatalln("Invalid resource ID. Expected format: projects/PROJECT_ID/locations/LOCATION/keyRings/KEY_RING/cryptoKeys/KEY_NAME")
	}
	keyRingName = parts[5]
	keyName = parts[7]
	location = parts[3]
	protectionLevel = parts[1] + "/" + parts[4] + "/" + parts[5] + "/cryptoKeyVersions/1"
	return keyRingName, keyName, location, protectionLevel
}

func GetCmkSpecFromCommand(cmd *cobra.Command) (*ybmclient.CMKSpec, error) {

	var cmkSpec *ybmclient.CMKSpec = nil
	if cmd.Flags().Changed("encryption-spec") {
		cmkString, _ := cmd.Flags().GetString("encryption-spec")
		cmkProvider := ""
		cmkAwsSecretKey := ""
		cmkAwsAccessKey := ""
		cmkAwsArnList := []string{}
		cmkGcpResourceId := ""
		cmkGcpServiceAccountPath := ""

		for _, cmkInfo := range strings.Split(cmkString, ",") {
			kvp := strings.Split(cmkInfo, "=")
			if len(kvp) != 2 {
				logrus.Fatalln("Incorrect format in cmk spec: configuration not provided as key=value pairs.")
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
			case "gcp-resource-id":
				if len(strings.TrimSpace(val)) != 0 {
					cmkGcpResourceId = val
				}
			case "gcp-service-account-path":
				if len(strings.TrimSpace(val)) != 0 {
					cmkGcpServiceAccountPath = val
				}
			}
		}

		if cmkProvider == "" {
			logrus.Fatalln("Incorrect format in cmk spec: please provide a cloud-provider.")
		}

		cmkSpec = ybmclient.NewCMKSpec(ybmclient.CMKProviderEnum(cmkProvider))

		switch cmkProvider {
		case "AWS":
			if cmkAwsAccessKey == "" {
				logrus.Fatalln("Incorrect format in cmk spec: AWS provider specified, but no aws-access-key provided.")
			}

			// The password/secret was not provided.
			if cmkAwsSecretKey == "" {
				// We should first check the environment variables.
				envVarName := "YBM_AWS_SECRET_KEY"
				value, exists := os.LookupEnv(envVarName)
				if exists {
					cmkAwsSecretKey = value
				} else {
					// If not found, prompt the user.
					fmt.Print("Please provide the AWS Secret Key for Encryption at Rest: ")

					data, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						logrus.Fatalln("Could not read AWS Secret key: ", err)
					}
					cmkAwsSecretKey = string(data)

					// Validate non-empty key
					if strings.TrimSpace(cmkAwsSecretKey) == "" {
						logrus.Fatalln("The AWS Secret Key cannot be empty")
					}
				}
			}
			cmkSpec.AwsCmkSpec.Set(ybmclient.NewAWSCMKSpec(cmkAwsAccessKey, cmkAwsSecretKey, cmkAwsArnList))
		case "GCP":
			if cmkGcpResourceId == "" {
				logrus.Fatalln("Incorrect format in CMK spec: GCP provider specified, but no gcp-resource-id provided")
			}
			keyRingName, keyName, location, protectionLevel := parseKMSKeyResourceID(cmkGcpResourceId)

			if cmkGcpServiceAccountPath == "" {
				logrus.Fatalln("Incorrect format in CMK spec: GCP provider specified, but no gcp-service-account-path provided")
			}
			cmkGcpServiceAccount, err := os.ReadFile(cmkGcpServiceAccountPath)
			if err != nil {
				logrus.Fatal("Incorrect file path for gcp service account file: ", err)
			}
			gcpCmkSpec := ybmclient.NewGCPCMKSpec(keyRingName, keyName, location, protectionLevel)

			var gcpServiceAccount ybmclient.GCPServiceAccount
			err = json.Unmarshal([]byte(cmkGcpServiceAccount), &gcpServiceAccount)

			if err != nil {
				logrus.Fatal("Failed to parse GCP service account credentials: invalid JSON format")
			}
			gcpCmkSpec.SetGcpServiceAccount(gcpServiceAccount)
			cmkSpec.SetGcpCmkSpec(*gcpCmkSpec)
		default:
			logrus.Fatalln("Incorrect format in CMK spec: invalid cloud-provider")
		}
	}
	return cmkSpec, nil
}
