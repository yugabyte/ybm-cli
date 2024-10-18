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

var _ = Describe("DB Query Logging", func() {

	var (
		server                          *ghttp.Server
		statusCode                      int
		args                            []string
		responseAccount                 openapi.AccountListResponse
		responseProject                 openapi.AccountListResponse
		pgLogExporterResponse           openapi.PgLogExporterConfigResponse
		pgLogExporterConfigListResponse openapi.PgLogExporterConfigListResponse
		responseListClusters            openapi.ClusterListResponse
		responseIntegrationList         openapi.TelemetryProviderListResponse
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
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

	Describe("When enabling query log exporter", func() {
		Context("with integration-name not set", func() {
			It("should throw error, integration-name not set", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "db-query-logging", "enable", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`\bError: required flag\(s\) "integration-name" not set\b`))
				session.Kill()
			})
		})
		Context("without log config params", func() {
			It("should enable db query logs with default log configs", func() {
				err := loadJson("./test/fixtures/db-query-log-exporter.json", &pgLogExporterResponse)
				Expect(err).ToNot(HaveOccurred())

				err = loadJson("./test/fixtures/list-telemetry-provider.json", &responseIntegrationList)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegrationList),
					),
				)

				statusCode = 202
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/cluster/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-query-log-exporter-configs"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, pgLogExporterResponse),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "db-query-logging", "enable",
					"--cluster-name", "stunning-sole",
					"--integration-name", "datadog-tp",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				// Specials characters like $ % [ ] etc must be escaped using backslash to do exact matching
				Expect(session.Out).Should(gbytes.Say(`State      Integration ID                         Log Config
ENABLING   9e740000-331b-4dec-89b0-4e59b81e9019   {"debug_print_plan":false,"log_connections":false,"log_disconnections":false,"log_duration":false,"log_error_verbosity":"DEFAULT","log_line_prefix":"\%m :\%r :\%u @ \%d :\[\%p\] :","log_min_duration_statement":-1,"log_min_error_statement":"ERROR","log_statement":"NONE"}`))
				Expect(server.ReceivedRequests()).Should(HaveLen(5))
				session.Kill()
			})
		})
		Context("with log config params", func() {
			It("should enable db query logs with provided log configs", func() {
				err := loadJson("./test/fixtures/db-query-log-exporter-custom.json", &pgLogExporterResponse)
				Expect(err).ToNot(HaveOccurred())

				err = loadJson("./test/fixtures/list-telemetry-provider.json", &responseIntegrationList)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegrationList),
					),
				)

				statusCode = 202
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/cluster/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-query-log-exporter-configs"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, pgLogExporterResponse),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "db-query-logging", "enable",
					"--cluster-name", "stunning-sole",
					"--integration-name", "datadog-tp",
					"--debug-print-plan", "true",
					"--log-connections", "true",
					"--log-disconnections", "false",
					"--log-duration", "false",
					"--log-error-verbosity", "TERSE",
					"--log-line-prefix", "%m :%r :%u @ %d :[%p] : %a :",
					"--log-min-duration-statement", "50",
					"--log-min-error-statement", "ERROR",
					"--log-statement", "MOD",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`State      Integration ID                         Log Config
ENABLING   9e740000-331b-4dec-89b0-4e59b81e9019   {"debug_print_plan":true,"log_connections":true,"log_disconnections":false,"log_duration":false,"log_error_verbosity":"TERSE","log_line_prefix":"\%m :\%r :\%u @ \%d :\[\%p\] : \%a :","log_min_duration_statement":50,"log_min_error_statement":"ERROR","log_statement":"MOD"}`))
				Expect(server.ReceivedRequests()).Should(HaveLen(5))
				session.Kill()
			})
		})
	})

	Describe("When disabling query log exporter", func() {
		Context("with logs enabled", func() {
			It("should disable query logs exporter", func() {
				err := loadJson("./test/fixtures/db-query-log-exporter-describe-resp.json", &pgLogExporterConfigListResponse)
				Expect(err).ToNot(HaveOccurred())
				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/cluster/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-query-log-exporter-configs"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, pgLogExporterConfigListResponse),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodDelete, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/cluster/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-query-log-exporter-configs/388c41b9-81f7-4ed9-8239-ab22f4aaca91"),
						ghttp.RespondWith(202, nil),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "db-query-logging", "disable",
					"--cluster-name", "stunning-sole", "-f",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Request submitted to disable DB query logging for the cluster, this may take a few minutes...
You can check the status via \$ ybm cluster db-query-logging describe --cluster-name stunning-sole`))
				Expect(server.ReceivedRequests()).Should(HaveLen(5))
				session.Kill()
			})
		})
	})
	Describe("When describe query log exporter", func() {
		It("should show a summary for query logs exporter", func() {
			err := loadJson("./test/fixtures/db-query-log-exporter-describe-resp.json", &pgLogExporterConfigListResponse)
			Expect(err).ToNot(HaveOccurred())
			statusCode = 200
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/cluster/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-query-log-exporter-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, pgLogExporterConfigListResponse),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "db-query-logging", "describe",
				"--cluster-name", "stunning-sole",
			)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`State     Integration ID                         Log Config
ACTIVE    9e740000-331b-4dec-89b0-4e59b81e9019   {"debug_print_plan":false,"log_connections":true,"log_disconnections":false,"log_duration":false,"log_error_verbosity":"DEFAULT","log_line_prefix":"\%m :\%r :\%u @ \%d :\[\%p\] : \%a :","log_min_duration_statement":30,"log_min_error_statement":"ERROR","log_statement":"MOD"}`))
			Expect(server.ReceivedRequests()).Should(HaveLen(4))
			session.Kill()
		})
	})

	Describe("When updating query log exporter config", func() {
		It("should update log config with provided args", func() {
			// Load existing log export config
			err := loadJson("./test/fixtures/db-query-log-exporter-describe-resp.json", &pgLogExporterConfigListResponse)
			Expect(err).ToNot(HaveOccurred())

			statusCode = 200
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/cluster/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-query-log-exporter-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, pgLogExporterConfigListResponse),
				),
			)

			// Load updated log export config response
			err = loadJson("./test/fixtures/db-query-log-exporter-update-config-resp.json", &pgLogExporterResponse)
			Expect(err).ToNot(HaveOccurred())

			statusCode = 202
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/cluster/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-query-log-exporter-configs/388c41b9-81f7-4ed9-8239-ab22f4aaca91"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, pgLogExporterResponse),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "db-query-logging", "update",
				"--cluster-name", "stunning-sole",
				// Change few query log configs
				"--debug-print-plan", "true",
				"--log-connections", "false",
				"--log-min-duration-statement", "60",
				"--log-statement", "ALL",
			)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`State     Integration ID                         Log Config
ACTIVE    9e740000-331b-4dec-89b0-4e59b81e9019   {"debug_print_plan":true,"log_connections":false,"log_disconnections":false,"log_duration":false,"log_error_verbosity":"DEFAULT","log_line_prefix":"\%m :\%r :\%u @ \%d :\[\%p\] : \%a :","log_min_duration_statement":60,"log_min_error_statement":"ERROR","log_statement":"ALL"}`))
			Expect(server.ReceivedRequests()).Should(HaveLen(5))
			session.Kill()
		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
