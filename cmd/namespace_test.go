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

var _ = Describe("Namespaces", func() {

	var (
		server              *ghttp.Server
		statusCode          int
		args                []string
		responseAccount     openapi.AccountResponse
		responseProject     openapi.AccountResponse
		responseListCluster openapi.ClusterListResponse
		responseNamespace   openapi.ClusterNamespacesListResponse
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
		err = loadJson("./test/fixtures/list-clusters.json", &responseListCluster)
		Expect(err).ToNot(HaveOccurred())
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListCluster),
			),
		)

		statusCode = 200
		err = loadJson("./test/fixtures/namespaces.json", &responseNamespace)
		Expect(err).ToNot(HaveOccurred())
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/namespaces"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNamespace),
			),
		)
	})

	Context("When listing namespaces", func() {
		It("should return the list of namespaces", func() {
			cmd := exec.Command(compiledCLIPath, "cluster", "namespace", "list", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`Namespace      Table Type
postgres       YSQL
test_ycql_db   YCQL
test_ysql_db   YSQL
yugabyte       YSQL`))
			session.Kill()
		})

	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})
})
