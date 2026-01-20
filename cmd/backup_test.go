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

package cmd_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var _ = Describe("Backup", func() {

	var (
		server             *ghttp.Server
		statusCode         int
		args               []string
		responseAccount    openapi.AccountResponse
		responseProject    openapi.AccountResponse
		responseBackupList openapi.BackupListResponse
		responseBackup     openapi.BackupResponse

		//cbr        *cobra.Command
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
		var err error
		server, err = newGhttpServer(responseAccount, responseProject)
		Expect(err).ToNot(HaveOccurred())
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
	})

	Context("When running with a valid Api token", func() {

		It("should return list of available backup", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/backups.json", &responseBackupList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/backups"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupList),
				),
			)
			cmd := exec.Command(compiledCLIPath, "backup", "list")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(fmt.Sprintf(`ID                                     Created On         Inc       Expire On          Clusters        State     Type      Roles & Grants
c7742a97-cee0-449d-9c7c-4b934d9cf940   %s   🍕        %s   mirthful-mole   ✅        🧑        Included
faaca956-b542-49ee-92a8-9f1e138d1311   %s   🟡        %s   mirthful-mole   ✅        🧑        Not Included`, formatter.FormatDate("2024-03-05T03:33:23.532Z"), formatter.FormatDateAndAddDays("2024-03-05T03:33:23.532Z", 8), formatter.FormatDate("2024-03-04T20:28:32.982Z"), formatter.FormatDateAndAddDays("2024-03-04T20:28:32.982Z", 1))))
			session.Kill()
		})

		It("should describe the given backup", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/backup.json", &responseBackup)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/backups/5574d58f-68f1-4762-baa1-c2421cdb38b0"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackup),
				),
			)
			cmd := exec.Command(compiledCLIPath, "backup", "describe", "--backup-id", "5574d58f-68f1-4762-baa1-c2421cdb38b0")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			expected := fmt.Sprintf(`General
ID                                     Created On         Inc       Clusters        State
5574d58f-68f1-4762-baa1-c2421cdb38b0   %s   🟡        mirthful-mole   ✅

Type      Size(bytes)   Expire On          Duration   Roles & Grants
🧑        1176035       %s   2 mins     Included


Databases/Keyspaces
Database/Keyspace   API Type
yugabyte            YSQL
my_keyspace         YCQL
`, formatter.FormatDate("2024-03-07T17:56:14.553Z"), formatter.FormatDateAndAddDays("2024-03-07T17:56:14.553Z", 1))

			o := string(session.Out.Contents()[:])
			Expect(o).Should(Equal(expected))
			session.Kill()

		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
