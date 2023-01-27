package cmd_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var (
	compiledCLIPath string
)

var _ = BeforeSuite(func() {
	var err error
	compiledCLIPath, err = gexec.Build("github.com/yugabyte/ybm-cli")
	Expect(compiledCLIPath).ToNot(BeEmpty())
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

func loadJson(filePath string, v any) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return err
	}
	return err
}

func newGhttpServer(responseAccount any, responseProject any) (*ghttp.Server, error) {
	server := ghttp.NewServer()
	err := loadJson("./test/fixtures/account.json", &responseAccount)
	if err != nil {
		return nil, err
	}
	err = loadJson("./test/fixtures/projects.json", &responseProject)
	if err != nil {
		return nil, err
	}
	statusCode := 200
	server.AppendHandlers(
		ghttp.CombineHandlers(
			ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts"),
			ghttp.RespondWithJSONEncodedPtr(&statusCode, responseAccount),
			ghttp.VerifyHeaderKV("Authorization", "Bearer test-token"),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects"),
			ghttp.RespondWithJSONEncodedPtr(&statusCode, responseProject),
		),
	)
	return server, nil
}
