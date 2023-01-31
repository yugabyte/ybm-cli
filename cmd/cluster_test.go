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
		server              *ghttp.Server
		statusCode          int
		args                []string
		responseAccount     openapi.AccountListResponse
		responseProject     openapi.ProjectListResponse
		responseListCluster openapi.ClusterListResponse
		responseError       openapi.ApiError
		responseCluster     openapi.ClusterData
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
			// 			It("should return only header when cluster-name is wrong", func() {
			// 				statusCode = 200
			// 				err := loadJson("./test/fixtures/no-clusters.json", &responseCluster)
			// 				Expect(err).ToNot(HaveOccurred())
			// 				server.AppendHandlers(
			// 					ghttp.CombineHandlers(
			// 						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
			// 						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCluster),
			// 						ghttp.VerifyFormKV("name", "test"),
			// 					),
			// 				)
			// 				cmd := exec.Command(compiledCLIPath, "get", "cluster", "--cluster-name", "test")
			// 				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			// 				Expect(err).NotTo(HaveOccurred())
			// 				session.Wait(2)
			// 				Expect(session.Out).Should(gbytes.Say(
			// 					`Name      Version   State     Health    Regions   Nodes     Total Res.\(Vcpu/Mem/Disk\)`))
			// 				session.Kill()
			// 			})
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
			// 			It("should return only header when cluster-name is wrong", func() {
			// 				statusCode = 200
			// 				err := loadJson("./test/fixtures/no-clusters.json", &responseCluster)
			// 				Expect(err).ToNot(HaveOccurred())
			// 				server.AppendHandlers(
			// 					ghttp.CombineHandlers(
			// 						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
			// 						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCluster),
			// 						ghttp.VerifyFormKV("name", "test"),
			// 					),
			// 				)
			// 				cmd := exec.Command(compiledCLIPath, "get", "cluster", "--cluster-name", "test")
			// 				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			// 				Expect(err).NotTo(HaveOccurred())
			// 				session.Wait(2)
			// 				Expect(session.Out).Should(gbytes.Say(
			// 					`Name      Version   State     Health    Regions   Nodes     Total Res.\(Vcpu/Mem/Disk\)`))
			// 				session.Kill()
			// 			})
		})
	})

	Describe("Cluster", func() {
		Context("with a valid Api token and default output table", func() {
			It("should return only header when cluster-name is wrong", func() {
				statusCode = 200
				err := loadJson("./test/fixtures/no-clusters.json", &responseCluster)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
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
