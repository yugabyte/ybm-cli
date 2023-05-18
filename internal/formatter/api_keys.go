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
	"fmt"

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultApiKeyListing = "table {{.ApiKeyName}}\t{{.RoleName}}\t{{.ApiKeyStatus}}\t{{.Issuer}}\t{{.CreatedAt}}\t{{.LastUsed}}\t{{.ExpiryTime}}"
	createdAtHeader      = "Date Created"
	expiryTimeHeader     = "Expiration"
	roleNameHeader       = "Role"
	lastUsedHeader       = "Last Used"
)

type ApiKeyContext struct {
	HeaderContext
	Context
	a ybmclient.ApiKeyData
}

func NewApiKeyFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultApiKeyListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// ApiKeyWrite renders the context for a list of API Keys
func ApiKeyWrite(ctx Context, keys []ybmclient.ApiKeyData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, key := range keys {
			err := format(&ApiKeyContext{a: key})
			if err != nil {
				logrus.Debugf("Error rendering API key: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewApiKeyContext(), render)
}

func SingleApiKeyWrite(ctx Context, key ybmclient.ApiKeyData) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&ApiKeyContext{a: key})
		if err != nil {
			logrus.Debugf("Error rendering API key: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewApiKeyContext(), render)
}

// NewApiKeyContext creates a new context for rendering API Keys
func NewApiKeyContext() *ApiKeyContext {
	apiKeyCtx := ApiKeyContext{}
	apiKeyCtx.Header = SubHeaderContext{
		"ApiKeyName":   apiKeyNameHeader,
		"ExpiryTime":   expiryTimeHeader,
		"ApiKeyStatus": apiKeyStatusHeader,
		"ID":           "ID",
		"RoleName":     roleNameHeader,
		"Issuer":       apiKeyIssuerHeader,
		"LastUsed":     lastUsedHeader,
		"CreatedAt":    createdAtHeader,
	}
	return &apiKeyCtx
}

func (a *ApiKeyContext) ApiKeyName() string {
	return fmt.Sprintf("%s", a.a.Spec.GetName())
}

func (a *ApiKeyContext) ExpiryTime() string {
	return fmt.Sprintf("%s", a.a.Info.ExpiryTime)
}

func (a *ApiKeyContext) ApiKeyStatus() string {
	return fmt.Sprintf("%s", a.a.Info.Status)
}

func (a *ApiKeyContext) ID() string {
	return a.a.Info.Id
}

func (a *ApiKeyContext) RoleName() string {
	return a.a.Info.Role.Info.GetDisplayName()
}

func (a *ApiKeyContext) Issuer() string {
	return a.a.Info.Issuer
}

func (a *ApiKeyContext) LastUsed() string {
	return a.a.Info.GetLastUsedTime()
}

func (a *ApiKeyContext) CreatedAt() string {
	return a.a.Info.Metadata.GetCreatedOn()
}

func (a *ApiKeyContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.a)
}
