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

package formatter_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var _ = Describe("Formatter", func() {

	Context("when using utils function ", func() {
		DescribeTable("to Truncate",
			func(text string, lenght int, expected string) {
				Expect(formatter.Truncate(text, lenght)).To(Equal(expected))
			},
			Entry("with text and lenght", "derftd", 3, "der..."),
			Entry("with text and lenght = 0", "", 0, ""),
			Entry("with not text and lenght", "", 3, ""),
			Entry("with japanese  and lenght", "コニチワ", 3, "コニチ..."),
		)

		DescribeTable("convert bytes to  MB",
			func(value int64, expected string) {
				Expect(formatter.ConvertBytestoGb(value)).To(Equal(expected))
			},
			Entry("with 1000000", int64(1000000), "1MB"),
			Entry("with 42000000", int64(42000000), "40MB"),
		)
	})

})
