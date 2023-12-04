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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var _ = Describe("BackupSchedules", func() {

	var (
		server                            *ghttp.Server
		statusCode                        int
		args                              []string
		responseAccount                   openapi.AccountListResponse
		responseProject                   openapi.AccountListResponse
		responseListBackupSchedules       openapi.ScheduleListResponse
		responseListPausedBackupSchedules openapi.ScheduleListResponse
		responseListCronBackupSchedules   openapi.ScheduleListResponse
		responseListClusters              openapi.ClusterListResponse
	)

	BeforeEach(func() {
		args = os.Args
		os.Args = []string{}
		var err error
		server, err = newGhttpServer(responseAccount, responseProject)
		Expect(err).ToNot(HaveOccurred())
		os.Setenv("YBM_HOST", fmt.Sprintf("http://%s", server.Addr()))
		os.Setenv("YBM_APIKEY", "test-token")
	})

	Describe("List backup schedules", func() {
		BeforeEach(func() {
			statusCode = 200
			err := loadJson("./test/fixtures/list-backup-schedules.json", &responseListBackupSchedules)
			Expect(err).ToNot(HaveOccurred())
			err = loadJson("./test/fixtures/list-clusters.json", &responseListClusters)
			Expect(err).ToNot(HaveOccurred())
			err = loadJson("./test/fixtures/list-backup-schedules-cron.json", &responseListCronBackupSchedules)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/clusters"),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListClusters),
				),
			)
		})
		Context("with a valid Api token and default output table", func() {
			It("should return list of backup schedules", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/backup-schedules"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListBackupSchedules),
					),
				)
				cmd := exec.Command(compiledCLIPath, "backup", "policy", "list", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				o := string(session.Out.Contents()[:])
				expected := `Time Interval(days)   Days of the Week   Backup Start Time
1                     NA                 NA` + "\n"
				Expect(o).Should(Equal(expected))

				session.Kill()
			})
			It("should not return list of backup schedules if policy is disabled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/backup-schedules"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListPausedBackupSchedules),
					),
				)
				cmd := exec.Command(compiledCLIPath, "backup", "policy", "list", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				o := string(session.Out.Contents()[:])
				// no backup schedule must be returned if it is paused
				expected := ``
				Expect(o).Should(Equal(expected))

				session.Kill()
			})
			It("should not return list of backup schedules if policy is disabled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/api/public/v1/accounts/340af43a-8a7c-4659-9258-4876fd6a207b/projects/78d4459c-0f45-47a5-899a-45ddf43eba6e/backup-schedules"),
						ghttp.RespondWithJSONEncodedPtr(&statusCode, responseListCronBackupSchedules),
					),
				)
				cmd := exec.Command(compiledCLIPath, "backup", "policy", "list", "--cluster-name", "stunning-sole")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(2)
				o := string(session.Out.Contents()[:])

				// no backup schedule must be returned if it is paused
				expected := `Time Interval(days)   Days of the Week   Backup Start Time
NA                    Su,We,Fr           ` + getLocalTime("2 3 * * *") + "\n"
				Expect(o).Should(Equal(expected))

				session.Kill()
			})

		})
	})

	AfterEach(func() {
		os.Args = args
		server.Close()
	})

})

func getLocalTime(cronExpression string) string {
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := cronParser.Parse(cronExpression)
	if err != nil {
		logrus.Debugln("Error parsing cron expression:\n", err)
		return ""
	}

	// Get the next scheduled time in UTC
	utcTime := schedule.Next(time.Now().UTC())
	localTimeZone, err := time.LoadLocation("Local")
	if err != nil {
		logrus.Debugln("Error loading local time zone:\n", err)
		return utcTime.Format("15:04") + "UTC"
	}
	localTime := utcTime.In(localTimeZone)
	localTimeString := localTime.Format("15:04")
	return localTimeString
}
