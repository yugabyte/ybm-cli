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
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var _ = Describe("Cluster", func() {

	var (
		server                   *ghttp.Server
		statusCode               int
		args                     []string
		responseAccount          openapi.AccountListResponse
		responseProject          openapi.ProjectListResponse
		responseListCluster      openapi.ClusterListResponse
		responseNetworkAllowList openapi.NetworkAllowListListResponse
		responseError            openapi.ApiError
		responseCluster          openapi.ClusterData
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

	Describe("Pausing cluster", func() {
		BeforeEach(func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-clusters.json", &responseListCluster)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListCluster),
				),
			)
		})
		Context("with a valid Api token and default output table", func() {
			It("should return success message", func() {
				statusCode = 200
				err := loadJson("./test/fixtures/pause-cluster.json", &responseCluster)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pause"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCluster),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "pause", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say("The cluster stunning-sole is being paused"))
				session.Kill()
			})
			It("should failed if cluster is already paused", func() {
				status := 409
				err := loadJson("./test/fixtures/pause-error.json", &responseError)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pause"),
						ghttp.RespondWithJSONEncodedPtr(&status, responseError),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "pause", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say("Cluster is not in an active state"))

				session.Kill()
			})
			It("should failed if cluster name is wrong", func() {
				status := 409
				err := loadJson("./test/fixtures/pause-error.json", &responseError)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pause"),
						ghttp.RespondWithJSONEncodedPtr(&status, responseError),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "pause", "--cluster-name", "stunnin-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say("Cluster is not in an active state"))

				session.Kill()
			})
		})
	})
	Describe("Resuming cluster", func() {
		BeforeEach(func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-clusters.json", &responseListCluster)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListCluster),
				),
			)
		})
		Context("with a valid Api token and default output table", func() {
			It("should return success message", func() {
				statusCode = 200
				err := loadJson("./test/fixtures/resume-cluster.json", &responseCluster)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/resume"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCluster),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "resume", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say("The cluster stunning-sole is being resumed"))
				session.Kill()
			})
			It("should failed if cluster is already paused", func() {
				status := 409
				err := loadJson("./test/fixtures/error-resume.json", &responseError)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/resume"),
						ghttp.RespondWithJSONEncodedPtr(&status, responseError),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "resume", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say("Cluster is not in paused state"))

				session.Kill()
			})
			It("should failed if cluster name is wrong", func() {
				status := 409
				err := loadJson("./test/fixtures/pause-error.json", &responseError)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pause"),
						ghttp.RespondWithJSONEncodedPtr(&status, responseError),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "pause", "--cluster-name", "stunnin-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say("Cluster is not in an active state"))

				session.Kill()
			})
		})
	})

	Describe("Get Cluster", func() {
		BeforeEach(func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-clusters.json", &responseListCluster)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListCluster),
				),
			)
		})
		Context("with a valid Api token and default output table", func() {
			It("should return list of cluster", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "get")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(
					`Name            Version       State     Health    Regions     Nodes     Total Res.\(Vcpu/Mem/Disk\)
stunning-sole   2.16.0.1-b7   ACTIVE    💚        us-west-2   1         2 / 8GB / 100GB`))
				session.Kill()
			})

			It("should return detailed summary of cluster if cluster-name is specified", func() {
				statusCode = 200
				err := loadJson("./test/fixtures/allow-list.json", &responseNetworkAllowList)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/allow-lists"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNetworkAllowList),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "get", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(
					`General
Name            ID                                     Version       State     Health
stunning-sole   5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8   2.16.0.1-b7   ACTIVE    💚
Provider   Fault Tolerance   Data Distribution   Nodes     Total Res.\(Vcpu/Mem/Disk\)
AWS        NONE              No idea             1         2 / 8GB / 100GB
Network AllowList
Name              Description       Allow List
device-ip-gween   device-ip-gween   152.165.26.42/32
Regions
Region      Nodes     vCPU/Node   Mem/Node   Disk/Node   VPC
us-west-2   1         2           8GB        100GB`))
				session.Kill()
			})
			It("should return only header when cluster-name is wrong", func() {
				statusCode = 200
				err := loadJson("./test/fixtures/no-clusters.json", &responseCluster)
				Expect(err).ToNot(HaveOccurred())
				server.SetHandler(2,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCluster),
						ghttp.VerifyFormKV("name", "test"),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "get", "--cluster-name", "test")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(
					`Name      Version   State     Health    Regions   Nodes     Total Res.\(Vcpu/Mem/Disk\)`))
				session.Kill()
			})
		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
