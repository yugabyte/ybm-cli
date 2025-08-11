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

var _ = Describe("Billing", func() {

	var (
		server                  *ghttp.Server
		statusCode              int
		args                    []string
		responseAccount         openapi.AccountResponse
		responseProject         openapi.AccountResponse
		billingEstimateResponse openapi.BillingEstimateResponse
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
		os.Unsetenv("YBM_FF_BILLING")
	})

	Context("When BILLING feature flag is disabled", func() {
		It("should not recognize billing command", func() {
			cmd := exec.Command(compiledCLIPath, "billing")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("unknown command \"billing\""))
			session.Kill()
		})

		It("should not recognize billing estimate command", func() {
			cmd := exec.Command(compiledCLIPath, "billing", "estimate")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("unknown command \"billing\""))
			session.Kill()
		})
	})

	Context("When BILLING feature flag is enabled", func() {
		BeforeEach(func() {
			os.Setenv("YBM_FF_BILLING", "true")
		})

		Describe("When running billing help", func() {
			It("should show billing help", func() {
				cmd := exec.Command(compiledCLIPath, "billing")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say("Billing operations for YugabyteDB Aeon"))
				Expect(session.Out).Should(gbytes.Say("estimate"))
				session.Kill()
			})
		})

		Describe("When running billing estimate", func() {
			BeforeEach(func() {
				err := loadJson("./test/fixtures/billing-estimate-response.json", &billingEstimateResponse)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/billing-estimate"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, billingEstimateResponse),
					),
				)
			})

			It("with no params should get billing estimate", func() {
				cmd := exec.Command(compiledCLIPath, "billing", "estimate")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Start Date   End Date     Total Amount
2025-01-01   2025-01-31   \$245.67

Account Name   Amount
account-1      \$123.45
account-2      \$122.22`))
				session.Kill()
			})

			It("should get billing estimate with all parameters", func() {
				cmd := exec.Command(compiledCLIPath, "billing", "estimate",
					"--start-date", "2025-01-01",
					"--end-date", "2025-01-31",
					"--account-names", "account-1,account-2")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Start Date   End Date     Total Amount
2025-01-01   2025-01-31   \$245.67

Account Name   Amount
account-1      \$123.45
account-2      \$122.22`))
				session.Kill()
			})
		})

		Describe("When running billing estimate with no billing account", func() {
			BeforeEach(func() {
				err := loadJson("./test/fixtures/billing-estimate-empty-accounts.json", &billingEstimateResponse)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/billing-estimate"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, billingEstimateResponse),
					),
				)
			})

			It("should show no account data message when accounts are empty", func() {
				cmd := exec.Command(compiledCLIPath, "billing", "estimate")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Start Date   End Date     Total Amount
2025-08-01   2025-08-11   \$0.00

No account data available.`))
				session.Kill()
			})
		})

		Describe("When running billing estimate with account not belonging to the user", func() {
			BeforeEach(func() {
				errorResponse := map[string]interface{}{
					"error": map[string]interface{}{
						"detail": "Cannot find account with name 'account-10'",
						"status": 400,
					},
				}

				statusCode = 404
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/billing-estimate"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, errorResponse),
					),
				)
			})

			It("should show error message when account not found", func() {
				cmd := exec.Command(compiledCLIPath, "billing", "estimate", "--account-names", "account-10")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say("Cannot find account with name 'account-10'"))
				session.Kill()
			})
		})
	})
})
