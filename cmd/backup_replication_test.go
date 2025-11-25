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

	Describe("When enabling GCP backup replication", func() {
		var responseCluster openapi.ClusterResponse

		BeforeEach(func() {
			err := loadJson("./test/fixtures/one-cluster.json", &responseCluster)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCluster),
				),
			)
		})

		Context("with cluster-name not set", func() {
			It("should throw error, cluster-name not set", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`Error: required flag\(s\)`))
				Expect(session.Err).Should(gbytes.Say(`"cluster-name"`))
				session.Kill()
			})
		})

		Context("with region-target not set", func() {
			It("should throw error, region-target not set", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`Error: required flag\(s\) "region-target" not set`))
				session.Kill()
			})
		})

		Context("with invalid format in region-target", func() {
			It("should throw error for incorrect format", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "invalid-format",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`incorrect format in region target`))
				session.Kill()
			})
		})

		Context("with region not specified in region-target", func() {
			It("should throw error for missing region", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "bucket-name=my-bucket",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`region not specified in region target`))
				session.Kill()
			})
		})

		Context("with bucket-name not specified in region-target", func() {
			It("should throw error for missing bucket-name", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "region=us-west-2",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`bucket name not specified in region target`))
				session.Kill()
			})
		})

		Context("with region not in cluster", func() {
			It("should throw error for invalid region", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "region=invalid-region,bucket-name=my-bucket",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`region 'invalid-region' is not a backup region in cluster stunning-sole`))
				session.Kill()
			})
		})

		Context("with duplicate region target", func() {
			It("should throw error for duplicate region", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "region=us-west-2,bucket-name=bucket1",
					"--region-target", "region=us-west-2,bucket-name=bucket2",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`duplicate region target provided for region 'us-west-2'`))
				session.Kill()
			})
		})

		Context("with valid region-target", func() {
			It("should enable backup replication in table format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{"regional_targets":[{"region":"us-west-2","target":"my-bucket"}]}`),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "region=us-west-2,bucket-name=my-bucket",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`GCP backup replication for cluster stunning-sole is being enabled`))
				Expect(session.Out).Should(gbytes.Say(`Overall State: ENABLED`))
				Expect(server.ReceivedRequests()).Should(HaveLen(6))
				session.Kill()
			})

			It("should enable backup replication in json format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{"regional_targets":[{"region":"us-west-2","target":"my-bucket"}]}`),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "region=us-west-2,bucket-name=my-bucket",
					"--output", "json",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state":"ENABLED"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(6))
				session.Kill()
			})

			It("should enable backup replication in pretty format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{"regional_targets":[{"region":"us-west-2","target":"my-bucket"}]}`),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "region=us-west-2,bucket-name=my-bucket",
					"--output", "pretty",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state": "ENABLED"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(6))
				session.Kill()
			})

			It("should handle spaces in region-target", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication.json", &responseBackupReplication)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{"regional_targets":[{"region":"us-west-2","target":"my-bucket"}]}`),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplication),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "enable",
					"--cluster-name", "stunning-sole",
					"--region-target", "region=us-west-2, bucket-name=my-bucket",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`GCP backup replication for cluster stunning-sole is being enabled`))
				Expect(server.ReceivedRequests()).Should(HaveLen(6))
				session.Kill()
			})
		})
	})

	Describe("When disabling GCP backup replication", func() {
		var responseCluster openapi.ClusterResponse

		BeforeEach(func() {
			err := loadJson("./test/fixtures/one-cluster.json", &responseCluster)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseCluster),
				),
			)
		})

		Context("with cluster-name not set", func() {
			It("should throw error, cluster-name not set", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "disable", "--force")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`Error: required flag\(s\)`))
				Expect(session.Err).Should(gbytes.Say(`"cluster-name"`))
				session.Kill()
			})
		})

		Context("with valid cluster-name and --force flag", func() {
			It("should disable backup replication in table format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication-disabled.json", &responseBackupReplicationDisabled)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{"regional_targets":[{"region":"us-west-2"}]}`),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplicationDisabled),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplicationDisabled),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "disable",
					"--cluster-name", "stunning-sole",
					"--force",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`GCP backup replication for cluster stunning-sole is being disabled`))
				Expect(session.Out).Should(gbytes.Say(`Overall State: DISABLED`))
				Expect(server.ReceivedRequests()).Should(HaveLen(6))
				session.Kill()
			})

			It("should disable backup replication in json format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication-disabled.json", &responseBackupReplicationDisabled)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{"regional_targets":[{"region":"us-west-2"}]}`),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplicationDisabled),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplicationDisabled),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "disable",
					"--cluster-name", "stunning-sole",
					"--force",
					"--output", "json",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state":"DISABLED"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(6))
				session.Kill()
			})

			It("should disable backup replication in pretty format", func() {
				err := loadJson("./test/fixtures/gcp-backup-replication-disabled.json", &responseBackupReplicationDisabled)
				Expect(err).ToNot(HaveOccurred())

				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{"regional_targets":[{"region":"us-west-2"}]}`),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplicationDisabled),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseBackupReplicationDisabled),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "disable",
					"--cluster-name", "stunning-sole",
					"--force",
					"--output", "pretty",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`"info"`))
				Expect(session.Out).Should(gbytes.Say(`"region_configs"`))
				Expect(session.Out).Should(gbytes.Say(`"state": "DISABLED"`))
				Expect(server.ReceivedRequests()).Should(HaveLen(6))
				session.Kill()
			})
		})

	})

	Describe("When disabling GCP backup replication with cluster having no backup regions", func() {
		BeforeEach(func() {
			var responseClusterNoBackupRegions openapi.ClusterResponse
			err := loadJson("./test/fixtures/one-cluster.json", &responseClusterNoBackupRegions)
			Expect(err).ToNot(HaveOccurred())

			clusterDataPtr, ok := responseClusterNoBackupRegions.GetDataOk()
			Expect(ok).To(BeTrue())
			clusterInfo := clusterDataPtr.GetInfo()
			clusterInfo.ClusterRegionInfoDetails = []openapi.ClusterRegionInfoDetails{
				{
					Region:       "us-west-2",
					Id:           "dc35fede-02f2-45e4-b763-f8ea24b74cf9",
					BackupRegion: false,
				},
			}
			clusterDataPtr.SetInfo(clusterInfo)

			statusCode = 200
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseClusterNoBackupRegions),
				),
			)
		})

		It("should throw error for no backup regions found", func() {
			cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "disable",
				"--cluster-name", "stunning-sole",
				"--force",
			)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			session.Wait(2)
			Expect(session.Err).Should(gbytes.Say(`no backup regions found for cluster stunning-sole`))
			session.Kill()
		})
	})

	Describe("When syncing GCP backup replication", func() {
		BeforeEach(func() {
			statusCode = 200
		})

		Context("with cluster-name not set", func() {
			It("should throw error, cluster-name not set", func() {
				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "sync")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`Error: required flag\(s\)`))
				Expect(session.Err).Should(gbytes.Say(`"cluster-name"`))
				session.Kill()
			})
		})

		Context("with valid cluster-name", func() {
			It("should trigger resync successfully", func() {
				statusCode = 200
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{}`),
						ghttp.RespondWith(statusCode, nil),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "sync",
					"--cluster-name", "stunning-sole",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Out).Should(gbytes.Say(`Resync triggered for all backup replication configs in cluster stunning-sole`))
				Expect(server.ReceivedRequests()).Should(HaveLen(4))
				session.Kill()
			})
		})

		Context("with API error", func() {
			It("should handle API error", func() {
				errorStatusCode := 400
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters/5f80730f-ba3f-4f7e-8c01-f8fa4c90dad8/backup-replication/gcp"),
						ghttp.VerifyJSON(`{}`),
						ghttp.RespondWithJSONEncoded(errorStatusCode, map[string]interface{}{
							"error": map[string]interface{}{
								"detail": "Invalid request",
								"status": 400,
							},
						}),
					),
				)

				cmd := exec.Command(compiledCLIPath, "cluster", "backup-replication", "gcp", "sync",
					"--cluster-name", "stunning-sole",
				)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				Expect(session.Err).Should(gbytes.Say(`Invalid request`))
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
