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

package formatter

import (
	"encoding/json"
	"strings"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	keyAliasHeader = "Key Alias"
)

type CMKContext struct {
	HeaderContext
	Context
	c ybmclient.CMKSpec
}

func NewCMKContext() *CMKContext {
	cmkContext := CMKContext{}
	cmkContext.Header = SubHeaderContext{
		"Provider":           providerHeader,
		"KeyAlias":           keyAliasHeader,
		"SecurityPrincipals": securityPrincipalsHeader,
	}
	return &cmkContext
}

func (c *CMKContext) Provider() string {
	return c.c.ProviderType
}

func (c *CMKContext) KeyAlias() string {
	//TODO: fix this with extracting GCP values when we support that
	return c.c.AwsCmkSpec.GetAliasName()
}

func (c *CMKContext) SecurityPrincipals() string {
	// TODO: fix this to pick up GCP later
	return strings.Join(c.c.AwsCmkSpec.GetArnList(), ", ")
}

func (c *CMKContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
