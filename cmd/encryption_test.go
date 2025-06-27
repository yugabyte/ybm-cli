package cmd_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var _ = Describe("Customer Managed Keys Test", func() {

	var (
		server              *ghttp.Server
		statusCode          int
		args                []string
		responseAccount     openapi.AccountResponse
		responseProject     openapi.AccountResponse
		responseListCluster openapi.ClusterListResponse
		responseCMK         openapi.CMKResponse
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
	})

	var _ = Describe("Describe Cluster CMK", func() {

		testCases := []struct {
			jsonFilePath string
			provider     string
			expected     string
		}{
			{
				jsonFilePath: "./test/fixtures/aws_cmk.json",
				provider:     "AWS",
				expected: `Provider   Key Alias                              Last Rotated               Security Principals                                                           CMK Status
AWS        0a80e409-e690-42fc-b209-baf969930b2c   2023-11-03T07:37:26.351Z   arn:aws:kms:us-east-1:745846189716:key/41c64d5g-c97d-472c-889e-0d9f80d2c754   ACTIVE` + "\n",
			},
			{
				jsonFilePath: "./test/fixtures/azure_cmk.json",
				provider:     "AZURE",
				expected: `Provider   Key Alias                              Last Rotated   Security Principals                      CMK Status
AZURE      8aXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX5b   -              https://test-azure-gj.vault.azure.net/   ACTIVE` + "\n",
			},
			{
				jsonFilePath: "./test/fixtures/gcp_cmk.json",
				provider:     "GCP",
				expected: `Provider   Key Alias      Last Rotated               Security Principals                                                                              CMK Status
GCP        GCP-test-key   2023-11-03T07:37:26.351Z   projects/<your-project-id>/locations/global/keyRings/GCP-test-key-ring/cryptoKeys/GCP-test-key   ACTIVE` + "\n",
			},
			{
				jsonFilePath: "./test/fixtures/azure_cmk_not_rotated.json",
				provider:     "AZURE",
				expected: `Provider   Key Alias                              Last Rotated   Security Principals                      CMK Status
AZURE      8aXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX5b   -              https://test-azure-gj.vault.azure.net/   ACTIVE` + "\n",
			},
		}

		for _, tc := range testCases {
			tc := tc
			It(fmt.Sprintf("should successfully get CMK details for %s", tc.provider), func() {
				err := loadJson(tc.jsonFilePath, &responseCMK)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "list", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				o := string(session.Out.Contents()[:])
				Expect(o).Should(Equal(tc.expected))
				session.Kill()
			})
		}

		It("should fail if no EAR configuration found for the cluster", func() {
			responseCMK = *openapi.NewCMKResponse()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "list", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No Encryption at rest configuration found for this cluster"))
			session.Kill()
		})
	})

	var _ = Describe("Update cluster CMK", func() {

		testCases := []struct {
			cloudProvider string
			spec          string
		}{
			{
				cloudProvider: "AWS",
				spec:          "cloud-provider=AWS,aws-secret-key=<your-secret-key>,aws-access-key=<your-access-key>",
			},
			{
				cloudProvider: "AZURE",
				spec:          "cloud-provider=AZURE,azu-client-id=<your-client-id>,azu-client-secret=<your-client-secret>,azu-tenant-id=<your-tenant-id>,azu-key-name=<your-key-name>,azu-key-vault-uri=<your-key-vault-uri>",
			},
			{
				cloudProvider: "GCP",
				spec:          "cloud-provider=GCP,gcp-resource-id=projects/<your-project>/locations/<your-location>/keyRings/<your-key-ring-name>/cryptoKeys/<your-key-name>,gcp-service-account-path=creds.json",
			},
		}

		BeforeEach(func() {
			fileName := "creds.json"
			os.WriteFile(fileName, []byte("{}"), 0644)
		})

		AfterEach(func() {
			os.Remove("creds.json")
		})

		for _, tc := range testCases {
			tc := tc

			It(fmt.Sprintf("should successfully update CMK for %s", tc.cloudProvider), func() {
				err := loadJson(fmt.Sprintf("./test/fixtures/%s_cmk.json", strings.ToLower(tc.cloudProvider)), &responseCMK)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
					),
				)
				cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update", "--cluster-name", "stunning-sole", "--encryption-spec", tc.spec)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say("Successfully updated encryption at rest for cluster stunning-sole"))
				session.Kill()
			})
		}

		It("should fail if invalid cloud provider in encryption at rest", func() {
			cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update", "--cluster-name", "stunning-sole", "--encryption-spec", "cloud-provider=TEST_PROVIDER,access-key=<your-access-key>,secret-key=<your-secret-key>")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Incorrect format in CMK spec: invalid cloud-provider"))
			session.Kill()
		})

		It("should fail if missing parameters in encryption at rest", func() {
			cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update", "--cluster-name", "stunning-sole", "--encryption-spec", "cloud-provider=AWS,aws-access-key=<your-access-key>,secret-key=<your-secret-key>")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Could not read AWS Secret key:"))
			session.Kill()
		})

		It("should succeed if no EAR configuration found for the cluster", func() {
			// In this case, a new EAR configuration would be added to this cluster
			responseCMK = *openapi.NewCMKResponse()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update", "--cluster-name", "stunning-sole", "--encryption-spec", "cloud-provider=AWS,aws-secret-key=<your-secret-key>,aws-access-key=<your-access-key>")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say("Successfully updated encryption at rest for cluster stunning-sole"))
			session.Kill()
		})
	})

	var _ = Describe("Update cluster CMK state", func() {

		testCases := []struct {
			action   string
			expected string
		}{
			{
				action:   "--enable",
				expected: "Successfully ENABLED encryption at rest status for cluster stunning-sole",
			},
			{
				action:   "--disable",
				expected: "Successfully DISABLED encryption at rest status for cluster stunning-sole",
			},
		}

		for _, tc := range testCases {
			tc := tc

			It(fmt.Sprintf("should successfully %s the CMK state for azure", tc.action), func() {

				err := loadJson("./test/fixtures/aws_cmk.json", &responseCMK)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks/41c64d5g-c97d-472c-889e-0d9f80d2c754/state"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update-state", "--cluster-name", "stunning-sole", tc.action)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(tc.expected))
				session.Kill()
			})
		}

		It("should fail if no or both flags are specified", func() {
			err := loadJson("./test/fixtures/aws_cmk.json", &responseCMK)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update-state", "--cluster-name", "stunning-sole")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Please enter valid input. Specify either enable or disable flag."))

			session.Kill()
		})

		It("should fail if both flags are specified", func() {
			err := loadJson("./test/fixtures/aws_cmk.json", &responseCMK)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update-state", "--cluster-name", "stunning-sole", "--enable", "--disable")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("Please enter valid input. Specify either enable or disable flag."))
			session.Kill()
		})

		It("should fail if no EAR configuration found for the cluster", func() {
			responseCMK = *openapi.NewCMKResponse()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/cmks"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCMK),
				),
			)
			cmd := exec.Command(compiledCLIPath, "cluster", "encryption", "update-state", "--cluster-name", "stunning-sole", "--enable")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("No Encryption at rest configuration found for this cluster"))
			session.Kill()
		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})
})
