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

var _ = Describe("Db Audit", func() {

	var (
		server              	*ghttp.Server
		statusCode          	int
		args                	[]string
		responseAccount     	openapi.AccountListResponse
		responseProject     	openapi.AccountListResponse
		responseDbAudit     	openapi.DbAuditExporterConfigResponse
		responseDbAuditList 	openapi.DbAuditExporterConfigListResponse
		responseIntegrationList openapi.TelemetryProviderListResponse
		responseListClusters    openapi.ClusterListResponse
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
		var err error
		server, err = newGhttpServer(responseAccount, responseProject)
		Expect(err).ToNot(HaveOccurred())
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
		os.Setenv("YBM_FF_DB_AUDIT_LOGS", "true")
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

	Context("When associating DB Audit config", func() {
		It("should associate cluster with DB Audit", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-telemetry-provider.json", &responseIntegrationList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegrationList),
				),
			)
			err = loadJson("./test/fixtures/db-audit-data.json", &responseDbAudit)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-audit-log-exporter-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseDbAudit),
				),
			)
			cmd := exec.Command(compiledCLIPath, "db-audit-logs-exporter", "assign", "--cluster-name", "stunning-sole", "--integration-name", "datadog-tp", "--ysql-config", "log_catalog=true,log_client=true,log_level=NOTICE,log_parameter=false,log_statement_once=false,log_relation=false", "--statement_classes", "READ,WRITE")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`The db audit exporter config 9e3fabbc-849c-4a77-bdb2-9422e712e7dc is being created
ID                                     Date Created               Cluster ID                             Integration ID                         State     Ysql Config
9e3fabbc-849c-4a77-bdb2-9422e712e7dc   2024-02-27T06:30:51.304Z   5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8   7c07c103-e3b2-48b6-ac30-764e9b5275e1   ACTIVE    {\"log_settings\":{\"log_catalog\":true,\"log_client\":true,\"log_level\":\"LOG\",\"log_parameter\":false,\"log_relation\":false,\"log_statement_once\":false},\"statement_classes\":\[\"READ\",\"WRITE\"]}`))
			session.Kill()
		})
		It("should return required field name and type when not set", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-telemetry-provider.json", &responseIntegrationList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegrationList),
				),
			)
			err = loadJson("./test/fixtures/db-audit-data.json", &responseDbAudit)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-audit-log-exporter-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseDbAudit),
				),
			)
			cmd := exec.Command(compiledCLIPath, "db-audit-logs-exporter", "assign")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say(`\bError: required flag\(s\) "integration-name", "ysql-config", "statement_classes", "cluster-name" not set\b`))
			session.Kill()
		})
		It("should return required log setting when not set", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-telemetry-provider.json", &responseIntegrationList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegrationList),
				),
			)
			err = loadJson("./test/fixtures/db-audit-data.json", &responseDbAudit)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-audit-log-exporter-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseDbAudit),
				),
			)
			cmd := exec.Command(compiledCLIPath, "db-audit-logs-exporter", "assign", "--cluster-name", "stunning-sole", "--integration-name", "datadog-tp", "--ysql-config", "log_catalog=true,log_client=true,log_level=NOTICE,log_parameter=false,log_relation=false", "--statement_classes", "READ,WRITE")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:log_statement_once required for log settings$)"))
			session.Kill()
		})
	})

	Context("When listing db audit exporter config", func() {
		It("should return the list of config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-db-audit.json", &responseDbAuditList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-audit-log-exporter-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseDbAuditList),
				),
			)
			cmd := exec.Command(compiledCLIPath, "db-audit-logs-exporter", "list", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`ID                                     Date Created               Cluster ID                             Integration ID                         State     Ysql Config
9e3fabbc-849c-4a77-bdb2-9422e712e7dc   2024-02-27T06:30:51.304Z   5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8   7c07c103-e3b2-48b6-ac30-764e9b5275e1   ACTIVE    {\"log_settings\":{\"log_catalog\":true,\"log_client\":true,\"log_level\":\"LOG\",\"log_parameter\":false,\"log_relation\":false,\"log_statement_once\":false},\"statement_classes\":\[\"READ\",\"WRITE\"]}`))
			session.Kill()
		})
		It("should return required field name and type when not set", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-telemetry-provider.json", &responseIntegrationList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegrationList),
				),
			)
			cmd := exec.Command(compiledCLIPath, "db-audit-logs-exporter", "list")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:Error: required flag\\(s\\) \"cluster-name\" not set$)"))
			session.Kill()
		})

	})

	Context("When removing db audit exporter config", func() {
		It("should delete the config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-db-audit.json", &responseDbAuditList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodDelete, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/db-audit-log-exporter-configs/123e4567-e89b-12d3-a456-426614174000"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseDbAuditList),
				),
			)
			cmd := exec.Command(compiledCLIPath, "db-audit-logs-exporter", "unassign", "--cluster-name", "stunning-sole", "--export-config-id", "123e4567-e89b-12d3-a456-426614174000", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`Deleting Db Audit Logs Exporter Config 123e4567-e89b-12d3-a456-426614174000`))
			session.Kill()
		})
		It("should return required field name and type when not set", func() {

			cmd := exec.Command(compiledCLIPath, "db-audit-logs-exporter", "unassign", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			exec.Command(compiledCLIPath, "y")
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:Error: required flag\\(s\\) \"export-config-id\", \"cluster-name\" not set$)"))
			session.Kill()
		})

	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
