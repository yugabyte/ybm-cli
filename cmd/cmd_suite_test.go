// Licensed to Yugabyte, Inc. under one or more contributor license
// agreements. See the NOTICE file distributed with this work for
// additional information regarding copyright ownership. Yugabyte
// licenses this file to you under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
