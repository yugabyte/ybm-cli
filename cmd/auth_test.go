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

var _ = Describe("Auth", func() {

	var (
		server     *ghttp.Server
		statusCode int
		args       []string
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
		var ApiErr openapi.ApiErrorError
		var responseError openapi.ApiError
		statusCode = 401
		server = ghttp.NewServer()
		ApiErr.SetDetail("JWT has expired")
		ApiErr.SetStatus(401)
		responseError.SetError(ApiErr)
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseError),
				ghttp.VerifyHeaderKV("Authorization", "Bearer test-token"),
			),
		)
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
	})

	Context("When Token is expired   ", func() {
		It("should ask to re-auth", func() {
			cmd := exec.Command(compiledCLIPath, "backup", "list")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say(
				`JWT has expired. Please run "ybm auth" again and provide a new API key`))
			session.Kill()
		})

	})
	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
