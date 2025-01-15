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

	Describe("When listing API keys with allow list FF disabled", func() {
		It("should hide allow list column", func() {
			err := loadJson("./test/fixtures/list-api-keys.json", &apiKeyListResponse)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/api-keys"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, apiKeyListResponse),
				),
			)
			os.Setenv("YBM_FF_API_KEY_ALLOW_LIST", "false")
			cmd := exec.Command(compiledCLIPath, "api-key", "list")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say(`Name       Role      Status    Created By    Date Created               Last Used                     Expiration
apikey-1   Admin     ACTIVE    user@yb.com   2025-01-13T13:55:14.024Z   2025-01-13T14:21:40.713599Z   2099-12-31T00:00Z
apikey-2   Admin     ACTIVE    user@yb.com   2025-01-08T09:06:35.077Z   Not yet used                  2025-02-07T09:06:35.077071Z`))
			session.Kill()
		})
	})

	Describe("When listing API keys with allow list FF enabled", func() {
		It("should show allow list column", func() {
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

			os.Setenv("YBM_FF_API_KEY_ALLOW_LIST", "true")
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

	Describe("When creating an API key", func() {
		Context("with allow list feature flag disabled", func() {
			os.Setenv("YBM_FF_API_KEY_ALLOW_LIST", "false")

			It("should error when passing allow list flag", func() {
				os.Setenv("YBM_FF_API_KEY_ALLOW_LIST", "false")
				cmd := exec.Command(compiledCLIPath, "api-key", "create", "--name", "apikey-1", "--duration", "30", "--unit", "DAYS", "--network-allow-lists", "device-ip")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`\bError: unknown flag: --network-allow-lists\b`))
				session.Kill()
			})

			It("should create API key", func() {
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
				Expect(session.Out).Should(gbytes.Say(`Name       Role      Status    Created By   Date Created               Last Used      Expiration
apikey-1   Admin     ACTIVE    admin        2025-01-13T15:55:55.825Z   Not yet used   2025-02-12T15:55:55.824833Z

API Key: test-jwt`))
				session.Kill()
			})
		})

		Context("with allow list feature flag enabled", func() {
			It("should create API key", func() {
				err := loadJson("./test/fixtures/get-or-create-api-key.json", &apiKeyResponse)
				Expect(err).ToNot(HaveOccurred())
				err = loadJson("./test/fixtures/allow-list.json", &responseNetworkAllowList)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/allow-lists"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNetworkAllowList),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/api-keys"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, apiKeyResponse),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/allow-lists"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseNetworkAllowList),
					),
				)

				os.Setenv("YBM_FF_API_KEY_ALLOW_LIST", "true")
				cmd := exec.Command(compiledCLIPath, "api-key", "create", "--name", "apikey-1", "--duration", "30", "--unit", "DAYS", "--network-allow-lists", "device-ip-gween")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Name       Role      Status    Created By   Date Created               Last Used      Expiration                    Allow List
apikey-1   Admin     ACTIVE    admin        2025-01-13T15:55:55.825Z   Not yet used   2025-02-12T15:55:55.824833Z   device-ip-gween

API Key: test-jwt`))
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
