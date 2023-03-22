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
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	Context("When", func() {
		DescribeTable("validating the token", func(now time.Time, valid bool) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"aud": "LoginJwt",
				"sub": "e31656e8-20a2-4605-8d68-fff83d24242f",
				"exp": time.Date(2023, 3, 9, 8, 0, 0, 0, time.UTC).Unix(),
				"iat": time.Date(2023, 3, 9, 7, 0, 0, 0, time.UTC).Unix(),
				"jti": "67b1ef70-2db2-4aa8-af10-8184de8bc61b",
			})

			hmacSampleSecret := []byte("secret")

			// Sign and get the complete encoded token as a string using the secret
			tokenString, _ := token.SignedString(hmacSampleSecret)

			claims, err := util.ExtractJwtClaims(tokenString)
			Expect(err).To(BeNil())
			Expect(claims).ToNot(BeNil())

			expired, err := util.IsJwtTokenExpiredWithTime(tokenString, now)
			Expect(err).To(BeNil())
			Expect(expired).To(Equal(valid))
		},
			Entry("When token is valid", time.Date(2023, 3, 9, 7, 30, 0, 0, time.UTC), false),
			Entry("When token is expired", time.Date(2023, 3, 9, 8, 30, 0, 0, time.UTC), true),
			Entry("When token is too early", time.Date(2023, 3, 9, 6, 30, 0, 0, time.UTC), true),
		)
	})
})
