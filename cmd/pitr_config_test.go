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

var _ = Describe("PITR Configs Test", func() {

	var (
		server                       *ghttp.Server
		statusCode                   int
		args                         []string
		responseAccount              openapi.AccountListResponse
		responseProject              openapi.AccountListResponse
		responseListCluster          openapi.ClusterListResponse
		responseNamespace            openapi.ClusterNamespacesListResponse
		responseListPITRConfig       openapi.ClusterPitrConfigListResponse
		responseGetPITRConfig        openapi.DatabasePitrConfigResponse
		responseCreatePITRConfig     openapi.BulkCreateDatabasePitrConfigResponse
		responseUpdatePITRConfig     openapi.UpdateDatabasePitrConfigResponse
		responseRestoreViaPITRConfig openapi.RestoreDatabaseViaPitrResponse
		responseClone                openapi.DatabaseCloneResponse
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
			Expect(session.Out).Should(gbytes.Say(`Namespace      Table Type   Retention Period in Days   State     Earliest Recovery Time in Millis   Latest Recovery Time in Millis
test_ycql_db   YCQL         6                          ACTIVE    123456                             123456789
test_ysql_db   YSQL         5                          QUEUED    654321                             987654321`))
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
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/create-cluster-pitr-config.json", &responseCreatePITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCreatePITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name=test_ysql_db, namespace-type=YSQL, retention-period-in-days=5", "--pitr-config", "namespace-name=test_ycql_db, namespace-type=YCQL, retention-period-in-days=3")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`The requested PITR Configurations are being created`))
			session.Kill()
		})

		It("Should fail if invalid namespace type in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name=test_ysql_db, namespace-type=PGSQL, retention-period-in-days=5")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Only YCQL or YSQL namespace types are allowed."))
			session.Kill()
		})

		It("Should fail if empty namespace name in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name=, namespace-type=YSQL, retention-period-in-days=5", "--pitr-config", "namespace-name=test_ycql_db, namespace-type=YCQL, retention-period-in-days=3")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Namespace name must be provided."))
			session.Kill()
		})

		It("Should fail if invalid key value pairs in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name test_ysql_db, namespace-type=YSQL, retention-period-in-days=5", "--pitr-config", "namespace-name=test_ycql_db, namespace-type=YCQL, retention-period-in-days=3")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("namespace-name, namespace-type and retention-period-in-days must be provided as key value pairs for each PITR Config to be created"))
			session.Kill()
		})

		It("Should fail if all required params are not in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name=test_ysql_db, namespace-type=YSQL ", "--pitr-config", "namespace-name=test_ycql_db, namespace-type=YCQL, retention-period-in-days=3")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("namespace-name, namespace-type and retention-period-in-days must be provided for each PITR Config to be created"))
			session.Kill()
		})

		It("Should fail if non int retention period in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name=test_ysql_db, namespace-type=YSQL, retention-period-in-days=five", "--pitr-config", "namespace-name=test_ycql_db, namespace-type=YCQL, retention-period-in-days=3")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("invalid syntax"))
			session.Kill()
		})

		It("Should fail if less than one day retention period in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name=test_ysql_db, namespace-type=YSQL, retention-period-in-days=1", "--pitr-config", "namespace-name=test_ycql_db, namespace-type=YCQL, retention-period-in-days=3")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Retention period must be between 2 and 14 days."))
			session.Kill()
		})

		It("Should fail if more than 14 days retention period in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "create", "--cluster-name", "stunning-sole", "--pitr-config", "namespace-name=test_ysql_db, namespace-type=YSQL, retention-period-in-days=15", "--pitr-config", "namespace-name=test_ycql_db, namespace-type=YCQL, retention-period-in-days=3")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Retention period must be between 2 and 14 days."))
			session.Kill()
		})
	})

	var _ = Describe("Restore cluster namespace via PITR config", func() {
		It("Should successfully restore YSQL namespace via PITR Config", func() {
			os.Setenv("YBM_FF_PITR_RESTORE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			restoreErr := loadJson("./test/fixtures/restore-ysql-database-via-pitr-config.json", &responseRestoreViaPITRConfig)
			Expect(restoreErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/07de8b5e-ce57-4ab5-8e29-97e57b20e76f/restore"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseRestoreViaPITRConfig),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "restore", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "--restore-at-millis", "4567", "--force")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Out).Should(gbytes.Say("The YSQL namespace test_ysql_db in cluster stunning-sole is being restored via PITR Configuration."))
			ysqlSession.Kill()
		})

		It("Should successfully restore YCQL namespace via PITR Config", func() {
			os.Setenv("YBM_FF_PITR_RESTORE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			restoreErr := loadJson("./test/fixtures/restore-ycql-database-via-pitr-config.json", &responseRestoreViaPITRConfig)
			Expect(restoreErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/249f9bf1-4276-4c60-8ab3-2bf1b2f6f1aa/restore"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseRestoreViaPITRConfig),
				),
			)

			ycqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "restore", "--cluster-name", "stunning-sole", "--namespace-name", "test_ycql_db", "--namespace-type", "YCQL", "--restore-at-millis", "4567", "--force")
			ycqlSession, ycqlErr := gexec.Start(ycqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ycqlErr).NotTo(HaveOccurred())
			ycqlSession.Wait(2)
			Expect(ycqlSession.Out).Should(gbytes.Say("The YCQL namespace test_ycql_db in cluster stunning-sole is being restored via PITR Configuration."))
			ycqlSession.Kill()
		})

		It("Should fail if invalid namespace name and type combination is provided", func() {
			os.Setenv("YBM_FF_PITR_RESTORE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "restore", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YCQL", "--restore-at-millis", "4567", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No YCQL namespace found with name test_ysql_db in cluster stunning-sole.\n"))
			session.Kill()
		})
	})

	var _ = Describe("Describe Cluster PITR Config", func() {
		It("Should successfully describe PITR Configs for the YSQL namespace in the cluster", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			getErr := loadJson("./test/fixtures/get-ysql-pitr-config.json", &responseGetPITRConfig)
			Expect(getErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/07de8b5e-ce57-4ab5-8e29-97e57b20e76f"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseGetPITRConfig),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "describe", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Out).Should(gbytes.Say(`Namespace      Table Type   Retention Period in Days   State     Earliest Recovery Time in Millis   Latest Recovery Time in Millis
test_ysql_db   YSQL         5                          QUEUED    654321                             987654321`))
			ysqlSession.Kill()
		})

		It("Should successfully describe PITR Configs for the YCQL namespace in the cluster", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			getErr := loadJson("./test/fixtures/get-ycql-pitr-config.json", &responseGetPITRConfig)
			Expect(getErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/249f9bf1-4276-4c60-8ab3-2bf1b2f6f1aa"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseGetPITRConfig),
				),
			)

			ycqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "describe", "--cluster-name", "stunning-sole", "--namespace-name", "test_ycql_db", "--namespace-type", "YCQL")
			ycqlSession, ycqlErr := gexec.Start(ycqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ycqlErr).NotTo(HaveOccurred())
			ycqlSession.Wait(2)
			Expect(ycqlSession.Out).Should(gbytes.Say(`Namespace      Table Type   Retention Period in Days   State     Earliest Recovery Time in Millis   Latest Recovery Time in Millis
test_ycql_db   YCQL         6                          ACTIVE    123456                             123456789`))
			ycqlSession.Kill()
		})

		It("should fail if no PITR Configs found for the cluster", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			responseListPITRConfig = *openapi.NewClusterPitrConfigListResponse()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "describe", "--cluster-name", "stunning-sole", "--namespace-name", "different-db", "--namespace-type", "YCQL")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No YCQL namespace found with name different-db in cluster stunning-sole.\n"))
			session.Kill()
		})

		It("Should fail if invalid namespace name and type combination is provided", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "describe", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YCQL")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No YCQL namespace found with name test_ysql_db in cluster stunning-sole.\n"))
			session.Kill()
		})
	})

	var _ = Describe("Delete Cluster PITR Config", func() {
		It("Should successfully delete PITR Configs for the YSQL namespace in the cluster", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodDelete, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/07de8b5e-ce57-4ab5-8e29-97e57b20e76f"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, "Successfully submitted remove Database PITR config request"),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "delete", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "-f")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Out).Should(gbytes.Say("The PITR Configuration for YSQL namespace test_ysql_db in cluster stunning-sole is being removed."))
			ysqlSession.Kill()
		})

		It("Should successfully delete PITR Configs for the YCQL namespace in the cluster", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodDelete, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/249f9bf1-4276-4c60-8ab3-2bf1b2f6f1aa"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, "Successfully submitted remove Database PITR config request"),
				),
			)

			ycqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "delete", "--cluster-name", "stunning-sole", "--namespace-name", "test_ycql_db", "--namespace-type", "YCQL", "-f")
			ycqlSession, ycqlErr := gexec.Start(ycqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ycqlErr).NotTo(HaveOccurred())
			ycqlSession.Wait(2)
			Expect(ycqlSession.Out).Should(gbytes.Say("The PITR Configuration for YCQL namespace test_ycql_db in cluster stunning-sole is being removed."))
			ycqlSession.Kill()
		})

		It("Should fail if invalid namespace name and type combination is provided", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "delete", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YCQL", "-f")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No YCQL namespace found with name test_ysql_db in cluster stunning-sole.\n"))
			session.Kill()
		})

		It("should fail if no PITR Configs found for the cluster", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			responseListPITRConfig = *openapi.NewClusterPitrConfigListResponse()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "delete", "--cluster-name", "stunning-sole", "--namespace-name", "different-db", "--namespace-type", "YCQL", "-f")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No YCQL namespace found with name different-db in cluster stunning-sole.\n"))
			session.Kill()
		})
	})

	var _ = Describe("Update PITR config", func() {
		It("Should successfully update PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			err = loadJson("./test/fixtures/update-cluster-pitr-config.json", &responseUpdatePITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs/07de8b5e-ce57-4ab5-8e29-97e57b20e76f"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseUpdatePITRConfig),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "update", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "--retention-period-in-days", "6", "--force")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Out).Should(gbytes.Say("The PITR Configuration for YSQL namespace test_ysql_db in cluster stunning-sole is being updated."))
			ysqlSession.Kill()
		})

		It("Should fail if invalid namespace name and type combination is provided", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)

			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "update", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YCQL", "--retention-period-in-days", "6", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No YCQL namespace found with name test_ysql_db in cluster stunning-sole.\n"))
			session.Kill()
		})

		It("Should fail if less than one day retention period in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "update", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "--retention-period-in-days", "1", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Retention period must be between 2 and 14 days."))
			session.Kill()
		})

		It("Should fail if more than 14 days retention period in PITR Config", func() {
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "update", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "--retention-period-in-days", "15", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Retention period must be between 2 and 14 days."))
			session.Kill()
		})
	})

	var _ = Describe("Clone cluster namespace via PITR config", func() {
		It("Should successfully clone YSQL namespace via pre existing PITR Config", func() {
			os.Setenv("YBM_FF_PITR_CLONE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			restoreErr := loadJson("./test/fixtures/clone-ysql-database-via-pitr-config.json", &responseClone)
			Expect(restoreErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/clone-database"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseClone),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "clone", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "--clone-as", "test_ysql_db_clone", "--clone-at-millis", "4567")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Out).Should(gbytes.Say("The YSQL namespace test_ysql_db in cluster stunning-sole is being cloned via PITR Configuration."))
			ysqlSession.Kill()
		})

		It("Should successfully clone YSQL namespace to current time via specified pre existing PITR Config", func() {
			os.Setenv("YBM_FF_PITR_CLONE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)

			restoreErr := loadJson("./test/fixtures/clone-now-ysql-db-existing-pitr.json", &responseClone)
			Expect(restoreErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/clone-database"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseClone),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "clone", "--cluster-name", "stunning-sole", "--namespace-name", "test_ysql_db", "--namespace-type", "YSQL", "--clone-as", "test_ysql_db_clone")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Out).Should(gbytes.Say("The YSQL namespace test_ysql_db in cluster stunning-sole is being cloned via PITR Configuration."))
			ysqlSession.Kill()
		})

		It("Should successfully clone YSQL namespace to current time with no pre existing PITR Config", func() {
			os.Setenv("YBM_FF_PITR_CLONE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			restoreErr := loadJson("./test/fixtures/clone-now-ysql-database.json", &responseClone)
			Expect(restoreErr).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/clone-database"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseClone),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "clone", "--cluster-name", "stunning-sole", "--namespace-name", "yugabyte", "--namespace-type", "YSQL", "--clone-as", "test_yugabyte_clone")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Out).Should(gbytes.Say("The YSQL namespace yugabyte in cluster stunning-sole is being cloned via PITR Configuration."))
			ysqlSession.Kill()
		})

		It("Should fail to clone YSQL namespace to current time with no pre existing PITR Config if clone time is specified", func() {
			os.Setenv("YBM_FF_PITR_CLONE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)

			ysqlCmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "clone", "--cluster-name", "stunning-sole", "--namespace-name", "yugabyte", "--namespace-type", "YSQL", "--clone-as", "test_ysql_db_clone", "--clone-at-millis", "4567")
			ysqlSession, ysqlErr := gexec.Start(ysqlCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Expect(ysqlErr).NotTo(HaveOccurred())
			ysqlSession.Wait(2)
			Expect(ysqlSession.Err).Should(gbytes.Say("No PITR configuration found for YSQL namespace yugabyte in cluster stunning-sole. The 'clone-at-millis' parameter cannot be used unless a valid PITR configuration is set up"))
			ysqlSession.Kill()
		})

		It("Should fail if invalid namespace name and type combination is provided", func() {
			os.Setenv("YBM_FF_PITR_CLONE", "true")
			err := loadJson("./test/fixtures/namespaces.json", &responseNamespace)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
				),
			)
			err = loadJson("./test/fixtures/list-cluster-pitr-configs.json", &responseListPITRConfig)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/pitr-configs"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPITRConfig),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "pitr-config", "clone", "--cluster-name", "stunning-sole", "--namespace-name", "yugabyte", "--namespace-type", "YCQL", "--clone-as", "test_yugabyte_clone")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No YCQL namespace found with name yugabyte in cluster stunning-sole.\n"))
			session.Kill()
		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})
})
