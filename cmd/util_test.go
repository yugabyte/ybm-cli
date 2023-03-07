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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/yugabyte/ybm-cli/cmd/util"
)

var _ = Describe("Utils", func() {
	Context("When ", func() {
		DescribeTable("validating CIDR",
			func(cidr string, valid bool) {

				isValidCiCR, err := util.ValidateCIDR(cidr)
				Expect(isValidCiCR).To(Equal(valid))
				if !valid {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).ToNot(HaveOccurred())
				}

			},
			Entry("When CIDR 10.0.0.0/16 is valid", "10.0.0.0/16", true),
			Entry("When CIDR 10.0.0.0/32 is valid", "10.0.0.0/32", true),
			Entry("When CIDR 10.0.0.0 is not valid", "10.0.0.0", false),
			Entry("When CIDR 10.0.0.0/23445 is not valid", "10.0.0.0/23445", false),
		)

	})
})
