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
		server          *ghttp.Server
		statusCode      int
		args            []string
		responseAccount openapi.AccountListResponse
		responseProject openapi.ProjectListResponse
		responseBackup  openapi.BackupListResponse
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
			err := loadJson("./test/fixtures/backups.json", &responseBackup)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/backups"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackup),
				),
			)
			cmd := exec.Command(compiledCLIPath, "backup", "get")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(fmt.Sprintf(
				`Created On         Expire On          Clusters                Description        State       Type      Retains\(day\)
%s   %s   proficient-parrotfish   scdasfdadfasdsad   SUCCEEDED   MANUAL    25`, formatter.FormatDate("2023-01-17T08:31:35.818Z"), formatter.FormatDate("2023-01-17T08:31:35.818Z"))))
			session.Kill()
		})

	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
