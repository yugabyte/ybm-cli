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
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var _ = Describe("Connection Pooling", func() {

	var (
		server               *ghttp.Server
		statusCode           int
		responseAccount      openapi.AccountResponse
		responseProject      openapi.AccountResponse
		responseListClusters openapi.ClusterListResponse
	)

	BeforeEach(func() {
		var err error
		server, err = newGhttpServer(responseAccount, responseProject)
		Expect(err).ToNot(HaveOccurred())
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
		statusCode = 200
		err = loadJson("./test/fixtures/list-clusters.json", &responseListClusters)
		Expect(err).ToNot(HaveOccurred())
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListClusters),
			),
		)
	})

	Context("When enabling connection pooling", func() {
		It("should enable connection pooling", func() {
			statusCode = 200

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/connection-pooling"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, "Successfully submitted cluster connection pooling operation request"),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "connection-pooling", "enable", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say("Connection Pooling for cluster stunning-sole is being enabled"))
			session.Kill()
		})
		It("should return required field name and type when not set", func() {
			statusCode = 200

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/connection-pooling"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, "Successfully submitted cluster connection pooling operation request"),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "connection-pooling", "enable")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:Error: required flag\\(s\\) \"cluster-name\" not set$)"))
			session.Kill()
		})

	})

	Context("When disabling connection pooling", func() {
		It("should disable connection pooling", func() {
			statusCode = 200

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/connection-pooling"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, "Successfully submitted cluster connection pooling operation request"),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "connection-pooling", "disable", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			exec.Command(compiledCLIPath, "y")
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say("Connection Pooling for cluster stunning-sole is being disabled"))
			session.Kill()
		})
		It("should return required field name and type when not set", func() {
			statusCode = 200

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/connection-pooling"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, "Successfully submitted cluster connection pooling operation request"),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "connection-pooling", "disable")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:Error: required flag\\(s\\) \"cluster-name\" not set$)"))
			session.Kill()
		})
	})
})
