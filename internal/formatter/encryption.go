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

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	keyAliasHeader    = "Key Alias"
	cmkStatusHeader   = "CMK Status"
	lastRotatedHeader = "Last Rotated"
	defaultCmkFormat  = "table {{.Provider}}\t{{.KeyAlias}}\t{{.LastRotated}}\t{{.SecurityPrincipals}}\t{{.CMKStatus}}"
)

type CMKContext struct {
	HeaderContext
	Context
	c ybmclient.CMKData
}

func NewCMKContext() *CMKContext {
	cmkContext := CMKContext{}
	cmkContext.Header = SubHeaderContext{
		"Provider":           providerHeader,
		"KeyAlias":           keyAliasHeader,
		"SecurityPrincipals": securityPrincipalsHeader,
		"CMKStatus":          cmkStatusHeader,
		"LastRotated":        lastRotatedHeader,
	}
	return &cmkContext
}

func NewCMKFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultCmkFormat
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func CMKWrite(ctx Context, cmkData ybmclient.CMKData) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&CMKContext{c: cmkData})
		if err != nil {
			logrus.Debug(err)
			return err
		}
		return nil
	}
	return ctx.Write(NewCMKContext(), render)
}

func (c *CMKContext) Provider() ybmclient.CMKProviderEnum {
	return c.c.Spec.Get().ProviderType
}

func (c *CMKContext) LastRotated() string {
	if c.c.Info.GetRotatedOn() != "" {
		return c.c.Info.GetRotatedOn()
	}
	return "-"
}

func (c *CMKContext) CMKStatus() ybmclient.CMKStatusEnum {
	return *c.c.GetSpec().Status.Get()
}

func (c *CMKContext) KeyAlias() string {
	if c.c.GetSpec().GcpCmkSpec.Get().GetKeyName() != "" {
		return c.c.GetSpec().GcpCmkSpec.Get().GetKeyName()
	} else if c.c.GetSpec().AzureCmkSpec.Get().GetKeyName() != "" {
		return c.c.GetSpec().AzureCmkSpec.Get().GetClientId()
	} else {
		return c.c.GetSpec().AwsCmkSpec.Get().GetAliasName()
	}
}

func (c *CMKContext) ResourceId() string {
	// Resource id: projects/{PROJECT_ID}/locations/{LOCATION}/keyRings/{KEY_RING_NAME}/cryptoKeys/{KEY_NAME}
	keyName := c.c.GetSpec().GcpCmkSpec.Get().GetKeyName()
	location := c.c.GetSpec().GcpCmkSpec.Get().GetLocation()
	keyRingName := c.c.GetSpec().GcpCmkSpec.Get().GetKeyRingName()
	projectId := c.c.GetSpec().GcpCmkSpec.Get().GetGcpServiceAccount().ProjectId
	resourceId := "projects/" + projectId + "/locations/" + location + "/keyRings/" + keyRingName + "/cryptoKeys/" + keyName
	return resourceId
}

func (c *CMKContext) SecurityPrincipals() string {
	if c.c.GetSpec().GcpCmkSpec.Get().GetKeyName() != "" {
		return c.ResourceId()
	} else if c.c.GetSpec().AzureCmkSpec.Get().GetKeyName() != "" {
		return c.c.GetSpec().AzureCmkSpec.Get().GetKeyVaultUri()
	}
	return strings.Join(c.c.GetSpec().AwsCmkSpec.Get().GetArnList(), ", ")
}

func (c *CMKContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
