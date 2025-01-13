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

var _ = Describe("API Key Management", func() {

	var (
		server                   *ghttp.Server
		statusCode               int
		args                     []string
		responseAccount          openapi.AccountListResponse
		responseProject          openapi.AccountListResponse
		apiKeyListResponse       openapi.ApiKeyListResponse
		apiKeyResponse           openapi.CreateApiKeyResponse
		responseNetworkAllowList openapi.NetworkAllowListListResponse
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
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

	Describe("When listing API keys", func() {
		Context("with valid request", func() {
			It("should list all API keys", func() {
				err := loadJson("./test/fixtures/list-api-keys.json", &apiKeyListResponse)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/api-keys"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, apiKeyListResponse),
					),
				)

				cmd := exec.Command(compiledCLIPath, "api-key", "list")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Name       Role      Status    Created By    Date Created               Last Used                     Expiration                    Allow List
apikey-1   Admin     ACTIVE    user@yb.com   2025-01-13T13:55:14.024Z   2025-01-13T14:21:40.713599Z   2099-12-31T00:00Z             N/A
apikey-2   Admin     ACTIVE    user@yb.com   2025-01-08T09:06:35.077Z   Not yet used                  2025-02-07T09:06:35.077071Z   N/A`))
				// Expect(server.ReceivedRequests()).Should(HaveLen(2))
				session.Kill()
			})
		})
	})

	Describe("When listing API keys with allow list", func() {
		Context("with valid request", func() {
			It("should list all API keys", func() {
				err := loadJson("./test/fixtures/list-api-keys-with-allow-list.json", &apiKeyListResponse)
				Expect(err).ToNot(HaveOccurred())
				err = loadJson("./test/fixtures/allow-list.json", &responseNetworkAllowList)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/api-keys"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, apiKeyListResponse),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/allow-lists"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNetworkAllowList),
					),
					func(w http.ResponseWriter, req *http.Request) {
						queryParams := req.URL.Query()
						Expect(queryParams.Get("status")).To(Equal("ACTIVE"))
					},
				)

				cmd := exec.Command(compiledCLIPath, "api-key", "list")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Name       Role      Status    Created By    Date Created               Last Used                     Expiration                    Allow List
apikey-1   Admin     ACTIVE    user@yb.com   2025-01-13T13:55:14.024Z   2025-01-13T14:21:40.713599Z   2099-12-31T00:00Z             device-ip-gween
apikey-2   Admin     ACTIVE    user@yb.com   2025-01-08T09:06:35.077Z   Not yet used                  2025-02-07T09:06:35.077071Z   N/A`))
				session.Kill()
			})
		})
	})

	Describe("When creating an API key", func() {
		Context("with valid request", func() {
			It("should create a new API key", func() {
				err := loadJson("./test/fixtures/get-or-create-api-key.json", &apiKeyResponse)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/api-keys"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, apiKeyResponse),
					),
				)

				cmd := exec.Command(compiledCLIPath, "api-key", "create", "--name", "apikey-1", "--duration", "30", "--unit", "DAYS")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Name       Role      Status    Created By   Date Created               Last Used      Expiration                    Allow List
apikey-1   Admin     ACTIVE    admin        2025-01-13T15:55:55.825Z   Not yet used   2025-02-12T15:55:55.824833Z   N/A

API Key: test-jwt 

The API key is only shown once after creation. Copy and store it securely.`))
				session.Kill()
			})
		})
	})

	Describe("When revoking an API key", func() {
		Context("with valid request", func() {
			It("should revoke the API key", func() {
				err := loadJson("./test/fixtures/get-or-create-api-key.json", &apiKeyResponse)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/api-keys"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, apiKeyResponse),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/api-keys/440af43a-8a7c-4659-9258-4876fd6a207b/revoke"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)

				cmd := exec.Command(compiledCLIPath, "api-key", "revoke", "--name", "apikey-1", "-f")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(server.ReceivedRequests()).Should(HaveLen(3))
				session.Kill()
			})
		})
	})
})
