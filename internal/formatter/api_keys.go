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
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultApiKeyListing = "table {{.ApiKeyName}}\t{{.ApiKeyRole}}\t{{.ApiKeyStatus}}\t{{.Issuer}}\t{{.CreatedAt}}\t{{.LastUsed}}\t{{.ExpiryTime}}"
	apiKeyListingV2      = "table {{.ApiKeyName}}\t{{.ApiKeyRole}}\t{{.ApiKeyStatus}}\t{{.Issuer}}\t{{.CreatedAt}}\t{{.LastUsed}}\t{{.ExpiryTime}}\t{{.AllowList}}"
	createdAtHeader      = "Date Created"
	expiryTimeHeader     = "Expiration"
	apiKeyRoleHeader     = "Role"
	lastUsedHeader       = "Last Used"
)

type ApiKeyDataAllowListInfo struct {
	ApiKey     *ybmclient.ApiKeyData
	AllowLists []string
}

type ApiKeyContext struct {
	HeaderContext
	Context
	a          ybmclient.ApiKeyData
	allowLists []string
}

func NewApiKeyFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultApiKeyListing
		if util.IsFeatureFlagEnabled(util.API_KEY_ALLOW_LIST) {
			format = apiKeyListingV2
		}
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// ApiKeyWrite renders the context for a list of API Keys
func ApiKeyWrite(ctx Context, keys []ApiKeyDataAllowListInfo) error {
	render := func(format func(subContext SubContext) error) error {
		for _, key := range keys {
			err := format(&ApiKeyContext{a: *key.ApiKey, allowLists: key.AllowLists})
			if err != nil {
				logrus.Debugf("Error rendering API key: %v", err)
				return err
			}
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
		"ApiKeyRole":   apiKeyRoleHeader,
		"Issuer":       apiKeyIssuerHeader,
		"LastUsed":     lastUsedHeader,
		"CreatedAt":    createdAtHeader,
		"AllowList":    "Allow List",
	}
	return &apiKeyCtx
}

func (a *ApiKeyContext) ApiKeyName() string {
	return a.a.Spec.GetName()
}

func (a *ApiKeyContext) ExpiryTime() string {
	return a.a.Info.ExpiryTime
}

func (a *ApiKeyContext) ApiKeyStatus() string {
	return string(a.a.Info.Status)
}

func (a *ApiKeyContext) ID() string {
	return a.a.Info.Id
}

func (a *ApiKeyContext) ApiKeyRole() string {
	return a.a.Info.Role.Info.GetDisplayName()
}

func (a *ApiKeyContext) Issuer() string {
	return a.a.Info.Issuer
}

func (a *ApiKeyContext) LastUsed() string {
	return a.a.Info.GetLastUsedTime()
}

func (a *ApiKeyContext) CreatedAt() string {
	return a.a.Info.Metadata.Get().GetCreatedOn()
}

func (a *ApiKeyContext) AllowList() string {
	if len(a.allowLists) == 0 {
		return "N/A"
	}
	return strings.Join(a.allowLists, ", ")
}

func (a *ApiKeyContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.a)
}
