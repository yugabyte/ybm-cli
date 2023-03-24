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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/yugabyte/ybm-cli/internal/log"
	"github.com/yugabyte/ybm-cli/internal/releases"
)

var _ = Describe("Releases", func() {
	BeforeEach(func() {
		log.SetLogLevel("", true)
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
})
