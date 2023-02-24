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

package formatter

import (
	"encoding/json"
	"strings"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultNalListing      = "table {{.Name}}\t{{.Desc}}\t{{.AllowedList}}\t{{.Clusters}}"
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
func (c *NetworkAllowListContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
