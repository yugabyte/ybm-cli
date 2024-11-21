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

var _ = Describe("Integration", func() {

	var (
		server                  *ghttp.Server
		statusCode              int
		args                    []string
		responseAccount         openapi.AccountListResponse
		responseProject         openapi.AccountListResponse
		responseIntegration     openapi.TelemetryProviderResponse
		responseIntegrationList openapi.TelemetryProviderListResponse
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
		var err error
		server, err = newGhttpServer(responseAccount, responseProject)
		Expect(err).ToNot(HaveOccurred())
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
		os.Setenv("YBM_FF_GOOGLECLOUD_INTEGRATION", "true")
	})

	Context("When type is Datadog", func() {
		It("should create the config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/metrics-exporter-dd.json", &responseIntegration)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegration),
				),
			)
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "datadog", "--datadog-spec", "site=test,api-key=c4XXXXXXXXXXXXXXXXXXXXXXXXXXXX3d")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`The Integration test has been created
ID                                     Name      Type      Site      ApiKey
9e3fabbc-849c-4a77-bdb2-9422e712e7dc   ff        DATADOG   test      c4XXXXXXXXXXXXXXXXXXXXXXXXXXXX3d`))
			session.Kill()
		})
		It("should return required field name and type when not set", func() {

			cmd := exec.Command(compiledCLIPath, "integration", "create")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:Error: required flag\\(s\\) \"config-name\", \"type\" not set$)"))
			session.Kill()
		})
		It("should return required field", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "datadog", "--datadog-spec", "site=test")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:api-key is a required field for datadog-spec$)"))
			session.Kill()
		})

	})
	Context("When type is Prometheus", func() {
		It("should create the config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/metrics-exporter-prom.json", &responseIntegration)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegration),
				),
			)
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "prometheus", "--prometheus-spec", "endpoint=http://prometheus.yourcompany.com")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`The Integration test has been created
ID                                     Name      Type         Endpoint
9e3fabbc-849c-4a77-bdb2-9422e712e7dc   test      PROMETHEUS   http://prometheus.yourcompany.com/api/v1/otlp`))
			session.Kill()
		})
		It("should return error when arg prometheus-spec not set", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "prometheus")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("prometheus-spec is required for prometheus sink"))
			session.Kill()
		})
		It("should return error when field endpoint not set", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "prometheus", "--prometheus-spec", "invalid-key=val")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:endpoint is a required field for prometheus-spec$)"))
			session.Kill()
		})

	})

	Context("When type is VictoriaMetrics", func() {
		It("should create the config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/metrics-exporter-victoriametrics.json", &responseIntegration)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegration),
				),
			)
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "victoriametrics", "--victoriametrics-spec", "endpoint=http://victoriametrics.yourcompany.com")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`The Integration test has been created
ID                                     Name      Type              Endpoint
9e3fabbc-849c-4a77-bdb2-9422e712e7dc   test      VICTORIAMETRICS   http://victoriametrics.yourcompany.com`))
			session.Kill()
		})
		It("should return error when arg victoriametrics-spec not set", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "victoriametrics")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("victoriametrics-spec is required for victoriametrics sink"))
			session.Kill()
		})
		It("should return error when field endpoint not set", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "victoriametrics", "--victoriametrics-spec", "invalid-key=val")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:endpoint is a required field for victoriametrics-spec$)"))
			session.Kill()
		})

	})

	Context("When type is Grafana", func() {
		It("should create the config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/metrics-exporter-grafana.json", &responseIntegration)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegration),
				),
			)
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "grafana", "--grafana-spec", "org-slug=ybmclitest,instance-id=1234456,zone=test-endpoint,access-policy-token=glXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX==")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`The Integration test has been created
ID                                     Name      Type      Zone        Access Token Policy                InstanceId   OrgSlug
92ceaa26-bac7-4842-9b3c-831a18a4f813   grafana   GRAFANA   test-zone   glXXXXXXXXXX...XXXXXXXXXXXXXXX==   1234456      ybmclitest`))
			session.Kill()
		})
		It("should return required field", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "test", "--type", "grafana", "--grafana-spec", "zone=test")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:access-policy-token is a required field for grafana-spec$)"))
			session.Kill()
		})

	})
	Context("When type is sumologic", func() {
		It("should create the config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/metrics-exporter-sumologic.json", &responseIntegration)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegration),
				),
			)
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "testsumo", "--type", "sumologic", "--sumologic-spec", "access-id=ybmclitest,access-key=1234456,installation-token=test-endpoint")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`ID                                     Name         Type        Access Key                         Access ID        InstallationToken
4021aa44-b22d-4bb6-a7d3-0a1099a4f206   gwenn-sumo   SUMOLOGIC   FqXXXXXXXXXX...XXXXXXXXXXXXXXX9p   suXXXXXXXXXXJ9   U1XXXXXXXXXX...XXXXXXXXXXXXXXX==`))
			session.Kill()
		})
		It("should return required field", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "testsumo", "--type", "sumologic", "--sumologic-spec", "access-id=test")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("(?m:access-key is a required field for sumologic-spec$)"))
			session.Kill()
		})

	})
	Context("When type is googlecloud ff is true", func() {
		It("should create the config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/metrics-exporter-googlecloud.json", &responseIntegration)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegration),
				),
			)
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "testgcp", "--type", "googlecloud", "--googlecloud-cred-filepath", "./test/fixtures/googlecloud-test-creds.json")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`The Integration testgcp has been created
ID                                     Name      Type
7913c052-fcd0-4b37-8a90-b0e47320190b   ddd       GOOGLECLOUD`))
			session.Kill()
		})
		It("should return filepath error", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "testgcp", "--type", "googlecloud", "--googlecloud-cred-filepath", "./test/fixtures/invalid-filepath.json")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("failed to open file"))
			session.Kill()
		})
		It("should return required field", func() {
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "testgcp", "--type", "googlecloud", "--googlecloud-cred-filepath", "./test/fixtures/invalid-googlecloud-test-creds.json")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("type is a required field for googlecloud credentials"))
			session.Kill()
		})
	})
	Context("When type is googlecloud and ff is false", func() {
		It("should return unknown flag error", func() {
			os.Setenv("YBM_FF_GOOGLECLOUD_INTEGRATION", "false")
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "testgcp", "--type", "googlecloud", "--googlecloud-cred-filepath", "./test/fixtures/googlecloud-test-creds.json")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("unknown flag: --googlecloud-cred-filepath"))
			session.Kill()
		})
		It("should return 'unsupported integration' error", func() {
			os.Setenv("YBM_FF_GOOGLECLOUD_INTEGRATION", "false")
			cmd := exec.Command(compiledCLIPath, "integration", "create", "--config-name", "testgcp", "--type", "googlecloud", "--datadog-spec", "site=test,api-key=c4XXXXXXXXXXXXXXXXXXXXXXXXXXXX3d")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Integration of type GOOGLECLOUD is currently not supported"))
			session.Kill()
		})
	})
	Context("When listing telememtry providers", func() {
		It("should return the list of config", func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-metrics-exporter.json", &responseIntegrationList)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/telemetry-providers"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseIntegrationList),
				),
			)
			cmd := exec.Command(compiledCLIPath, "integration", "list")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`ID                                     Name         Type
129f7c97-81ae-47c7-8f9e-40ab4390093f   ff           DATADOG
92ceaa26-bac7-4842-9b3c-831a18a4f813   grafana      GRAFANA
4021aa44-b22d-4bb6-a7d3-0a1099a4f206   gwenn-sumo   SUMOLOGIC`))
			session.Kill()
		})

	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
