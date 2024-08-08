package cmd_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	// "strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var _ = Describe("PITR Configs Test", func() {

	var (
		server                       *ghttp.Server
		statusCode                   int
		args                         []string
		responseAccount              openapi.AccountListResponse
		responseProject              openapi.AccountListResponse
		responseListCluster          openapi.ClusterListResponse
		responseListPITRConfig       openapi.ClusterPitrConfigListResponse
		responseCreatePITRConfig     openapi.CreateDatabasePitrConfigResponse
		responseRestoreViaPITRConfig openapi.RestoreDatabaseViaPitrResponse
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
		var err error
		server, err = newGhttpServer(responseAccount, responseProject)
		Expect(err).ToNot(HaveOccurred())
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
		os.Setenv("YBM_FF_PITR_CONFIG", "true")

		statusCode = 200
		err = loadJson("./test/fixtures/list-clusters.json", &responseListCluster)
		Expect(err).ToNot(HaveOccurred())
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListCluster),
			),
		)
	})

	var _ = Describe("List Cluster PITR Configs", func() {
		It("Should successfully list PITR Configs for the cluster", func() {
			err := loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "list", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`Namespace      Table Type   Retention Period in Days   Backup Interval in Seconds   State     Date Created
test_ycql_db   YCQL         6                          86400                        ACTIVE    2024-08-03T11:38:10.838Z
test_ysql_db   YSQL         5                          86400                        QUEUED    2024-08-03T11:34:06.456Z`))
			session.Kill()
		})

		It("should fail if no PITR Configs found for the cluster", func() {
			responseListPITRConfig = *openapi.NewClusterPitrConfigListResponse()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "list", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No PITR Configs found for cluster.\n"))
			session.Kill()
		})
	})

	var _ = Describe("Create PITR config", func() {
		It("Should successfully create PITR config", func() {
			err := loadJson("./test/fixtures/create-cluster-pitr-config.json", &responseCreatePITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCreatePITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "--retention-period-in-days", "5")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`Successfully created PITR configuration.

Namespace      Table Type   Retention Period in Days   Backup Interval in Seconds   State     Date Created
test_ysql_db   YSQL         5                          86400                        QUEUED    2024-08-07T16:26:08.435Z`))
			session.Kill()
		})
	})

	It("Should fail if invalid namespace type in PITR Config", func() {
		cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "PGSQL", "--retention-period-in-days", "5")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		session.Wait(2)
		Expect(session.Err).Should(gbytes.Say("Only YCQL or YSQL namespace types are allowed."))
		session.Kill()
	})

	var _ = Describe("Restore cluster namespace via PITR config", func() {
		It("Should successfully restore namespace via PITR Config", func() {
			err := loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			restoreErr := loadJson("./test/fixtures/restore-cluster-database-via-pitr-config.json", &responseRestoreViaPITRConfig)
			Expect(restoreErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/07de8b5e-ce57-4ab5-8e29-97e57b20e76f/restore"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseRestoreViaPITRConfig),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "restore", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--restore-at-millis", "4567", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say("Successfully restored namespace test_ysql_db at 4567 ms."))
			session.Kill()
		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})
})
