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

var _ = Describe("Backup Replication", func() {

	var (
		server                            *ghttp.Server
		statusCode                        int
		args                              []string
		responseAccount                   openapi.AccountResponse
		responseProject                   openapi.AccountResponse
		responseListClusters              openapi.ClusterListResponse
		responseBackupReplication         openapi.GCPBackupReplicationResponse
		responseBackupReplicationDisabled openapi.GCPBackupReplicationResponse
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
		var err error
		server, err = newGhttpServer(responseAccount, responseProject)
		Expect(err).ToNot(HaveOccurred())
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
		os.Setenv("YBM_FF_BACKUP_REPLICATION_GCP_TARGET", "true")
		statusCode = 200
		err = loadJson("./test/fixtures/list-clusters.json", &responseListClusters)
		Expect(err).ToNot(HaveOccurred())
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
				ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListClusters),
			),
		)
	})

	Describe("When describing GCP backup replication", func() {
		Context("with cluster-name not set", func() {
			It("should throw error, cluster-name not set", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`\bError: required flag\(s\) "cluster-name" not set\b`))
				session.Kill()
			})
		})

		Context("with valid cluster-name and active replication", func() {
			It("should show backup replication configuration in table format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe",
					"--cluster-name", "stunning-sole",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Overall State: ENABLED`))
				Expect(session.Out).Should(gbytes.Say(`Region`))
				Expect(session.Out).Should(gbytes.Say(`Config State`))
				Expect(session.Out).Should(gbytes.Say(`Bucket Name`))
				Expect(session.Out).Should(gbytes.Say(`Latest Operation Status`))
				Expect(session.Out).Should(gbytes.Say(`Last Run`))
				Expect(session.Out).Should(gbytes.Say(`Next Run`))
				Expect(session.Out).Should(gbytes.Say(`us-west1`))
				Expect(session.Out).Should(gbytes.Say(`backup-bucket-us-west1`))
				Expect(session.Out).Should(gbytes.Say(`SUCCESS`))
				Expect(session.Out).Should(gbytes.Say(`2024-01-15,09:30`))
				Expect(session.Out).Should(gbytes.Say(`2024-01-15,10:00`))
				Expect(session.Out).Should(gbytes.Say(`us-east1`))
				Expect(session.Out).Should(gbytes.Say(`backup-bucket-us-east1`))
				Expect(session.Out).Should(gbytes.Say(`IN_PROGRESS`))
				Expect(session.Out).Should(gbytes.Say(`N/A`))
				Expect(session.Out).Should(gbytes.Say(`2024-01-15,11:00`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})

			It("should show backup replication configuration in json format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe",
					"--cluster-name", "stunning-sole",
					"--output", "json",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state":"ENABLED"`))
				Expect(session.Out).Should(gbytes.Say(`"region":"us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"target":"backup-bucket-us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"region":"us-east1"`))
				Expect(session.Out).Should(gbytes.Say(`"target":"backup-bucket-us-east1"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})

			It("should show backup replication configuration in pretty format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe",
					"--cluster-name", "stunning-sole",
					"--output", "pretty",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state": "ENABLED"`))
				Expect(session.Out).Should(gbytes.Say(`"region": "us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"target": "backup-bucket-us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"region": "us-east1"`))
				Expect(session.Out).Should(gbytes.Say(`"target": "backup-bucket-us-east1"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})
		})

		Context("with disabled replication", func() {
			It("should show DISABLED state when no active reports", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication-disabled.json", &responseBackupReplicationDisabled)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplicationDisabled),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe",
					"--cluster-name", "stunning-sole",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Overall State: DISABLED`))
				Expect(session.Out).Should(gbytes.Say(`DISABLED`))
				Expect(session.Out).Should(gbytes.Say(`backup-bucket-us-west1`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})
		})

		Context("with show-all flag", func() {
			It("should show active and expired configurations in table format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe",
					"--cluster-name", "stunning-sole",
					"--show-all",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`=== Active Configurations ===`))
				Expect(session.Out).Should(gbytes.Say(`Region`))
				Expect(session.Out).Should(gbytes.Say(`Config State`))
				Expect(session.Out).Should(gbytes.Say(`Bucket Name`))
				Expect(session.Out).Should(gbytes.Say(`Latest Operation Status`))
				Expect(session.Out).Should(gbytes.Say(`Last Run`))
				Expect(session.Out).Should(gbytes.Say(`Next Run`))
				Expect(session.Out).Should(gbytes.Say(`us-west1`))
				Expect(session.Out).Should(gbytes.Say(`backup-bucket-us-west1`))
				Expect(session.Out).Should(gbytes.Say(`SUCCESS`))
				Expect(session.Out).Should(gbytes.Say(`us-east1`))
				Expect(session.Out).Should(gbytes.Say(`backup-bucket-us-east1`))
				Expect(session.Out).Should(gbytes.Say(`IN_PROGRESS`))
				Expect(session.Out).Should(gbytes.Say(`=== Configurations Set for Expiry ===`))
				Expect(session.Out).Should(gbytes.Say(`Expiry Time`))
				Expect(session.Out).Should(gbytes.Say(`2024-01-15,08:00`))
				Expect(session.Out).Should(gbytes.Say(`2024-01-15,09:00`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})

			It("should show backup replication configuration in json format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe",
					"--cluster-name", "stunning-sole",
					"--output", "json",
					"--show-all",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state":"ENABLED"`))
				Expect(session.Out).Should(gbytes.Say(`"region":"us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"target":"backup-bucket-us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"region":"us-east1"`))
				Expect(session.Out).Should(gbytes.Say(`"target":"backup-bucket-us-east1"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})

			It("should show backup replication configuration in pretty format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "describe",
					"--cluster-name", "stunning-sole",
					"--output", "pretty",
					"--show-all",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state": "ENABLED"`))
				Expect(session.Out).Should(gbytes.Say(`"region": "us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"target": "backup-bucket-us-west1"`))
				Expect(session.Out).Should(gbytes.Say(`"region": "us-east1"`))
				Expect(session.Out).Should(gbytes.Say(`"target": "backup-bucket-us-east1"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})
		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})
