package cmd_test

import (
	"encoding/csv"
	"encoding/json"
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

var _ = Describe("Spend Tracking", func() {

	var (
		server                *ghttp.Server
		statusCode            int
		args                  []string
		responseAccount       openapi.AccountListResponse
		responseProject       openapi.AccountListResponse
		responseListClusters  openapi.ListClustersByDateRangeResponse
		responseSpendTracking openapi.BillingUsageResponse
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
		err = loadJson("./test/fixtures/list-clusters-date-range.json", &responseListClusters)
		Expect(err).ToNot(HaveOccurred())
		err = loadJson("./test/fixtures/usage-data.json", &responseSpendTracking)
		Expect(err).ToNot(HaveOccurred())
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/billing-usage/clusters"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListClusters),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/billing-usage"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseSpendTracking),
			),
		)
	})

	Context("Validating output", func() {
		It("should return json file in output", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", "usage", "--output-format", "json")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)

			usageContents, err := os.ReadFile("usage.json")
			Expect(err).ToNot(HaveOccurred())

			expectedContents, err := os.ReadFile("./test/fixtures/usage-data.json")
			Expect(err).ToNot(HaveOccurred())

			// Unmarshal the JSON data into a map for both actual and expected
			var actualData map[string]interface{}
			var expectedData map[string]interface{}
			err = json.Unmarshal(usageContents, &actualData)
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(expectedContents, &expectedData)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualData).To(Equal(expectedData))
			Expect(session.Out).Should(gbytes.Say("JSON data written to usage.json\n"))
			os.Remove("usage.json")
			session.Kill()
		})

		It("should return csv file in output", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", "usage", "--output-format", "csv")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)

			usageContents, err := os.ReadFile("usage.csv")
			Expect(err).ToNot(HaveOccurred())

			reader := csv.NewReader(strings.NewReader(string(usageContents)))

			records, err := reader.ReadAll()
			Expect(err).ToNot(HaveOccurred())

			// Join all rows from the csv
			actual := ""
			for _, row := range records {
				actual += strings.Join(row, ",") + "\n"
			}

			expected := "Date,Clusters,VCPUS_Daily,VCPUS_Cumulative,DISK_STORAGE_Daily,DISK_STORAGE_Cumulative,DISK_IOPS_Daily,DISK_IOPS_Cumulative,BACKUP_STORAGE_Daily,BACKUP_STORAGE_Cumulative,DATA_TRANSFER_WITHIN_REGION_Daily,DATA_TRANSFER_WITHIN_REGION_Cumulative,DATA_TRANSFER_CROSS_REGION_APAC_Daily,DATA_TRANSFER_CROSS_REGION_APAC_Cumulative,DATA_TRANSFER_CROSS_REGION_NON_APAC_Daily,DATA_TRANSFER_CROSS_REGION_NON_APAC_Cumulative,DATA_TRANSFER_INTERNET_Daily,DATA_TRANSFER_INTERNET_Cumulative\n" +
				"2023-08-13,'courageous-jellyfish','willing-walrus',120.520000,120.520000,6026.000000,6026.000000,0.000000,0.000000,0.000000,0.000000,1.962233,1.962233,0.000000,0.000000,0.173324,0.173324,0.118982,0.118982\n" +
				"2023-08-14,'courageous-jellyfish','willing-walrus',0.000000,120.520000,0.000000,6026.000000,0.000000,0.000000,0.000000,0.000000,0.000000,1.962233,0.000000,0.000000,0.000000,0.173324,0.000023,0.119005\n" +
				"2023-08-15,'courageous-jellyfish','willing-walrus',0.000000,120.520000,0.000000,6026.000000,0.000000,0.000000,0.000000,0.000000,0.000000,1.962233,0.000000,0.000000,0.000000,0.173324,0.000018,0.119023\n"

			Expect(actual).To(Equal(expected))
			Expect(session.Out).Should(gbytes.Say("CSV data written to usage.csv\n"))
			os.Remove("usage.csv")
			session.Kill()
		})

		It("should fail if the clusters is not present", func() {
			clusterName := "charming-canid, passionate-manatee"
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", "usage", "--output-format", "csv", "--cluster-name", "charming-canid", "--cluster-name", "passionate-manatee")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("clusters '%s' not found within the specified time range", clusterName))
			session.Kill()
		})
	})

	Context("Date related", func() {
		It("should succeed if start date or end date is in yyyy-mm-dd format", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-18", "--end", "2023-10-15", "--output-file", "usage", "--output-format", "csv")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say("CSV data written to usage.csv\n"))
			os.Remove("usage.csv")
			session.Kill()
		})
		It("should fail if start date or end date not present", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--output-file", "usage", "--output-format", "csv")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("both start date and end date are required"))
			session.Kill()
		})

		It("should fail if start date is before end date", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-18T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", "usage", "--output-format", "csv")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("start date must be before end date"))
			session.Kill()
		})

		It("should fail if the start date or end date is in invalid format", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "18-02-2022", "--end", "2022-10-19T15:30:00Z", "--output-file", "usage", "--output-format", "csv")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("invalid start date format"))
			session.Kill()
		})
	})

	Context("Filename related", func() {
		It("should succeed if file name and format not specified", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			// It should take default fileame and format
			expectedFileName := "usage_20221012T153000_20221015T153000.csv"
			_, err = os.Stat(expectedFileName)
			Expect(err).NotTo(HaveOccurred())
			os.Remove(expectedFileName)
			session.Kill()
		})

		It("should fail if the file extension is invalid", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", "usage", "--output-format", "abc")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("file extension abc is not supported"))
			session.Kill()
		})

		It("should succeed if file already exists but force flag is present", func() {
			fileName := "usage.json"
			err := os.WriteFile(fileName, []byte("dummy content"), 0644)
			Expect(err).NotTo(HaveOccurred())
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", fileName, "--output-format", "json", "--force")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)

			usageContents, err := os.ReadFile("usage.json")
			Expect(err).ToNot(HaveOccurred())

			expectedContents, err := os.ReadFile("./test/fixtures/usage-data.json")
			Expect(err).ToNot(HaveOccurred())

			var actualData map[string]interface{}
			var expectedData map[string]interface{}
			err = json.Unmarshal(usageContents, &actualData)
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(expectedContents, &expectedData)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualData).To(Equal(expectedData))
			Expect(session.Out).Should(gbytes.Say("JSON data written to usage.json\n"))
			os.Remove("usage.json")
			session.Kill()
		})

		It("should fail if the filename already exists without force flag", func() {
			dummyFilename := "usage.csv"
			err := os.WriteFile(dummyFilename, []byte("dummy content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", "usage", "--output-format", "csv")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say("file usage.csv already exists"))

			os.Remove(dummyFilename)
			session.Kill()
		})

		It("should succeed if the filename has a valid extension", func() {
			cmd := exec.Command(compiledCLIPath, "usage", "get", "--start", "2022-10-12T15:30:00Z", "--end", "2022-10-15T15:30:00Z", "--output-file", "usage.json", "--output-format", "csv")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Out).Should(gbytes.Say("JSON data written to usage.json\n"))
			os.Remove("usage.json")
			session.Kill()
		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})
})
