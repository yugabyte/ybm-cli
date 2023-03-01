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

package signup

import (
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var SignUpCmd = &cobra.Command{
	Use:   "signup",
	Short: "Open a browser to sign up for YugabyteDB Managed",
	Long:  "Open a browser to sign up for YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		//util.OpenBrowser("https://cloud.yugabyte.com/signup")
		browser.OpenURL("https://cloud.yugabyte.com/signup")
	},
}
