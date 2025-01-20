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
	defaultNalListing      = "table {{.Name}}\t{{.Desc}}\t{{.AllowedList}}\t{{.Clusters}}\t{{.ApiKeys}}"
	networkAllowListHeader = "Allow List"
)

type NetworkAllowListContext struct {
	HeaderContext
	Context
	c ybmclient.NetworkAllowListData
}

func NewNetworkAllowListFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultNalListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// NetworkAllowListWrite renders the context for a list of network allow lists
func NetworkAllowListWrite(ctx Context, nals []ybmclient.NetworkAllowListData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, nal := range nals {
			err := format(&NetworkAllowListContext{c: nal})
			if err != nil {
				logrus.Debugf("Error rendering network allow list: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewNetworkAllowListContext(), render)
}

// NewNetworkAllowListContext creates a new context for rendering nal
func NewNetworkAllowListContext() *NetworkAllowListContext {
	nalCtx := NetworkAllowListContext{}
	nalCtx.Header = SubHeaderContext{
		"AllowedList": networkAllowListHeader,
		"Clusters":    clustersHeader,
		"Desc":        descriptionHeader,
		"Name":        nameHeader,
		"ApiKeys":     "API keys",
	}
	return &nalCtx
}

func (c *NetworkAllowListContext) Name() string {
	return c.c.Spec.Name
}

func (c *NetworkAllowListContext) AllowedList() string {
	return strings.Join(c.c.GetSpec().AllowList, ",")
}

func (c *NetworkAllowListContext) Desc() string {
	return c.c.GetSpec().Description
}
func (c *NetworkAllowListContext) Clusters() string {
	var clusterNameList []string
	for _, cluster := range c.c.GetInfo().ClusterList {
		clusterNameList = append(clusterNameList, cluster.Name)
	}
	return strings.Join(clusterNameList, ",")
}

func (c *NetworkAllowListContext) ApiKeys() string {
	var apiKeyNames []string
	for _, apiKey := range c.c.GetInfo().ApiKeyList {
		if apiKey.Status != nil && *apiKey.Status == ybmclient.APIKEYSTATUSENUM_ACTIVE {
			apiKeyNames = append(apiKeyNames, apiKey.Name)
		}
	}
	return strings.Join(apiKeyNames, ",")
}

func (c *NetworkAllowListContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
