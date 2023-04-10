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
package releases_test

import (
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/internal/log"
	"github.com/yugabyte/ybm-cli/internal/releases"
)

var _ = Describe("Releases", func() {
	BeforeEach(func() {
		log.SetLogLevel("", false)
		log.SetDebugFormatter()
	})

	Context("when checking releases", func() {
		DescribeTable("to check if should upgrade",
			func(release1, release2 string, expected bool) {
				Expect(releases.ShouldUpgrade(release1, release2)).To(Equal(expected))
			},
			Entry("Empty release", "", "", false),
			Entry("Equal releases", "1.2.3", "1.2.3", false),
			Entry("Release 1 is greater than release 2", "1.2.3", "1.2.2", false),
			Entry("Release 1 is less than release 2", "1.2.3", "1.2.4", true),
			Entry("[v1] Equal releases", "v1.2.3", "1.2.3", false),
			Entry("[v2] Equal releases", "1.2.3", "v1.2.3", false),
			Entry("[vv1] Equal releases", "vv1.2.3", "1.2.3", false),
			Entry("[vv2] Equal releases", "1.2.3", "vv1.2.3", false),
			Entry("[vv1vv2] Equal releases", "vv1.2.3", "vv1.2.3", false),
			Entry("[v1]Release 1 is greater than release 2", "v1.2.3", "1.2.2", false),
			Entry("[v2]Release 1 is less than release 2", "1.2.3", "v1.2.4", true),
			Entry("[vv1]Release 1 is greater than release 2", "vv1.2.3", "1.2.2", false),
			Entry("[vv2]Release 1 is less than release 2", "1.2.3", "vv1.2.4", true),
		)
	})

	Context("When fetching the latest releases from Github", func() {
		It("should not fetch if the current timestamp is less than 24 hours from the existing timestamp", func() {
			fetchFromGithub, err := releases.ShouldFetchLatestRelease("0", 3600)
			Expect(err).ToNot(HaveOccurred())
			Expect(fetchFromGithub).To(Equal(false))
		})
		It("fetch if the current timestamp is more than 24 hours from the existing timestamp", func() {
			fetchFromGithub, err := releases.ShouldFetchLatestRelease("0", 3600*25)
			Expect(err).ToNot(HaveOccurred())
			Expect(fetchFromGithub).To(Equal(true))
		})
		It("Throw error if the timestamp contains a character", func() {
			_, err := releases.ShouldFetchLatestRelease("0v", 3600*25)
			Expect(err).To(HaveOccurred())
		})
		It("Throw error if the timestamp is null", func() {
			_, err := releases.ShouldFetchLatestRelease("", 3600*25)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When fetching the release config from viper", func() {
		It("Fetch release config properly from viper", func() {
			lastVersionAvailable := "v0.0.1"
			lastCheckedTime := "10"
			viper.GetViper().Set("lastVersionAvailable", &lastVersionAvailable)
			viper.GetViper().Set("lastCheckedTime", &lastCheckedTime)
			releaseConfig, err := releases.GetReleaseConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(releaseConfig.LastVersionAvailable).To(Equal(lastVersionAvailable))
			Expect(releaseConfig.LastCheckedTime).To(Equal(lastCheckedTime))
		})
		It("Fetch release config with invalid version", func() {
			lastVersionAvailable := "0.0.1"
			lastCheckedTime := "10"
			viper.GetViper().Set("lastVersionAvailable", &lastVersionAvailable)
			viper.GetViper().Set("lastCheckedTime", &lastCheckedTime)
			_, err := releases.GetReleaseConfig()
			Expect(err).To(HaveOccurred())
		})
		It("Fetch release config with null string as version", func() {
			lastVersionAvailable := ""
			lastCheckedTime := "10"
			viper.GetViper().Set("lastVersionAvailable", &lastVersionAvailable)
			viper.GetViper().Set("lastCheckedTime", &lastCheckedTime)
			_, err := releases.GetReleaseConfig()
			Expect(err).To(HaveOccurred())
		})
		It("Fetch release config with negative time", func() {
			lastVersionAvailable := "v0.1.1"
			lastCheckedTime := "-10"
			viper.GetViper().Set("lastVersionAvailable", &lastVersionAvailable)
			viper.GetViper().Set("lastCheckedTime", &lastCheckedTime)
			_, err := releases.GetReleaseConfig()
			Expect(err).To(HaveOccurred())
		})
		It("Fetch release config with time having a character", func() {
			lastVersionAvailable := "v0.1.1"
			lastCheckedTime := "10v"
			viper.GetViper().Set("lastVersionAvailable", &lastVersionAvailable)
			viper.GetViper().Set("lastCheckedTime", &lastCheckedTime)
			_, err := releases.GetReleaseConfig()
			Expect(err).To(HaveOccurred())
		})
		It("Fetch release config with time being a null string", func() {
			lastVersionAvailable := "v0.1.1"
			lastCheckedTime := ""
			viper.GetViper().Set("lastVersionAvailable", &lastVersionAvailable)
			viper.GetViper().Set("lastCheckedTime", &lastCheckedTime)
			_, err := releases.GetReleaseConfig()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When writing release config to viper", func() {
		It("Write release config to viper", func() {
			version := "v0.0.1"
			timestamp := 10
			timestampString := strconv.Itoa(timestamp)
			viper.SetConfigFile(".tmp")
			viper.SetConfigType("yaml")
			err := releases.WriteReleaseConfig(int64(timestamp), version)
			Expect(err).ToNot(HaveOccurred())
			Expect(viper.GetString("lastCheckedTime")).To(Equal(timestampString))
			Expect(viper.GetString("lastVersionAvailable")).To(Equal(version))
			os.Remove(".tmp")
		})
		It("Write release config to viper with invalid version", func() {
			version := "0.0.1"
			timestamp := 10
			viper.SetConfigFile(".tmp")
			viper.SetConfigType("yaml")
			err := releases.WriteReleaseConfig(int64(timestamp), version)
			Expect(err).To(HaveOccurred())
			os.Remove(".tmp")
		})
		It("Write release config to viper with null version", func() {
			version := ""
			timestamp := 10
			viper.SetConfigFile(".tmp")
			viper.SetConfigType("yaml")
			err := releases.WriteReleaseConfig(int64(timestamp), version)
			Expect(err).To(HaveOccurred())
			os.Remove(".tmp")
		})
		It("Write release config to viper with negative timestamp", func() {
			version := ""
			timestamp := 10
			viper.SetConfigFile(".tmp")
			viper.SetConfigType("yaml")
			err := releases.WriteReleaseConfig(int64(timestamp), version)
			Expect(err).To(HaveOccurred())
			os.Remove(".tmp")
		})

	})
})
