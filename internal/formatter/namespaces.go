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
	"sort"

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultNamespaceListing = "table {{.Namespace}}\t{{.TableType}}"
	namespaceHeader         = "Namespace"
	tableTypeHeader         = "Table Type"
)

type NamespaceContext struct {
	HeaderContext
	Context
	n ybmclient.ClusterNamespaceData
}

func NewNamespaceFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultNamespaceListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NamespaceWrite(ctx Context, namespace []ybmclient.ClusterNamespaceData) error {
	//Sort by name
	sort.Slice(namespace, func(i, j int) bool {
		return string(namespace[i].Name) < string(namespace[j].Name)
	})

	render := func(format func(subContext SubContext) error) error {
		for _, namespace := range namespace {
			if namespace.GetTableType() == "REDIS_TABLE_TYPE" || namespace.GetTableType() == "TRANSACTION_STATUS_TABLE_TYPE" {
				continue
			}
			err := format(&NamespaceContext{n: namespace})
			if err != nil {
				logrus.Debugf("Error rendering namespace: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewNamespaceContext(), render)
}

func NewNamespaceContext() *NamespaceContext {
	namespaceCtx := NamespaceContext{}
	namespaceCtx.Header = SubHeaderContext{
		"Namespace": namespaceHeader,
		"TableType": tableTypeHeader,
	}
	return &namespaceCtx
}

func (n *NamespaceContext) Namespace() string {
	return n.n.GetName()
}

func (n *NamespaceContext) TableType() string {
	if n.n.GetTableType() == "PGSQL_TABLE_TYPE" {
		return "YSQL"
	}
	return "YCQL"
}

func (n *NamespaceContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.n)
}
