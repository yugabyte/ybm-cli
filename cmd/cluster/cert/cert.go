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

package cert

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

var CertCmd = &cobra.Command{
	Use:   "cert",
	Short: "Get the root CA certificate",
	Long:  "Get the root CA certificate for your YBM clusters",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var downloadCertificate = &cobra.Command{
	Use:   "download",
	Short: "Download the root CA certificate",
	Long:  `Download the root CA certificate`,
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")
		certificate, err := authApi.GetConnectionCertificate()
		if err != nil {
			logrus.Fatal("Fail to retrieve connection certtificate: ", err)
		}

		if output, _ := cmd.Flags().GetString("output"); output != "" {
			// check if the file exists
			if _, err := os.Stat(output); err == nil {
				// Only overwrite if force is set
				if !cmd.Flags().Changed("force") {
					logrus.Fatalf("File %s already exists", output)
				}
			}

			f, err := os.Create(output)
			if err != nil {
				logrus.Fatal("Fail to create output file: ", err)
			}
			defer f.Close()
			_, err = f.WriteString(certificate)
			if err != nil {
				logrus.Fatal("Fail to write to output file: ", err)
			}
		} else {
			fmt.Println(certificate)
		}

	},
}

func init() {
	CertCmd.AddCommand(downloadCertificate)
	downloadCertificate.Flags().StringP("output", "o", "", "[OPTIONAL] Output file name (default: stdout)")
	downloadCertificate.Flags().BoolP("force", "f", false, "[OPTIONAL] Overwrite the output file if it exists")
}
