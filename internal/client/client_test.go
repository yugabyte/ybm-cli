// Copyright (c) YugaByte, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022- Yugabyte, Inc.

package client_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/sirupsen/logrus"
	"github.com/yugabyte/ybm-cli/internal/client"
)

var _ = Describe("Client", func() {
	Context("When Parsing URL", func() {
		It("should return host and https scheme when no scheme is specified", func() {
			url, err := client.ParseURL("myurl.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(url.String()).To(Equal("https://myurl.com"))
		})
		It("should return same url scheme is specified", func() {
			buffer := gbytes.NewBuffer()
			logrus.SetOutput(buffer)
			url, err := client.ParseURL("http://myurl.com")
			Expect(buffer).To(gbytes.Say("level=warning msg=\"you are using insecure api endpoint http://myurl.com\""))
			Expect(err).ToNot(HaveOccurred())
			Expect(url.String()).To(Equal("http://myurl.com"))
		})
		It("should return same url scheme is specified", func() {
			url, err := client.ParseURL("https://myurl.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(url.String()).To(Equal("https://myurl.com"))
		})
	})
})
