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
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultDrListing    = "table {{.Id}}\t{{.Name}}\t{{.SourceCluster}}\t{{.TargetCluster}}\t{{.Databases}}\t{{.State}}\t{{.CreatedOn}}"
	sourceClusterHeader = "Source Cluster"
	targetClusterHeader = "Target Cluster"
	databasesHeader     = "Databases"
	drCreatedOnHeader   = "Created On"
)

type DrContext struct {
	HeaderContext
	Context
	c ybmclient.XClusterDrData
	a ybmAuthClient.AuthApiClient
}

func NewDrFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultDrListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// DrWrite renders the context for a list of DRs
func DrWrite(ctx Context, Drs []ybmclient.XClusterDrData, authApi ybmAuthClient.AuthApiClient) error {
	render := func(format func(subContext SubContext) error) error {
		for _, Dr := range Drs {
			err := format(&DrContext{c: Dr, a: authApi})
			if err != nil {
				logrus.Debugf("Error in rendering DR context: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewDrContext(), render)
}

// NewDrContext creates a new context for rendering DR
func NewDrContext() *DrContext {
	DrCtx := DrContext{}
	DrCtx.Header = SubHeaderContext{
		"Id":            idHeader,
		"Name":          nameHeader,
		"SourceCluster": sourceClusterHeader,
		"TargetCluster": targetClusterHeader,
		"Databases":     databasesHeader,
		"State":         stateHeader,
		"CreatedOn":     drCreatedOnHeader,
	}
	return &DrCtx
}

func (c *DrContext) Name() string {
	if v, ok := c.c.Spec.GetNameOk(); ok {
		return *v
	}
	return ""
}

func (c *DrContext) Id() string {
	if v, ok := c.c.Info.GetIdOk(); ok {
		return *v
	}
	return ""
}

func (c *DrContext) TargetCluster() string {
	if v, ok := c.c.Info.GetTargetClusterIdOk(); ok {
		clusterResp, _, err := c.a.GetCluster(*v).Execute()
		if err == nil {
			return clusterResp.Data.Spec.GetName()
		}
	}
	return ""
}

func (c *DrContext) SourceCluster() string {
	if v, ok := c.c.Info.GetSourceClusterIdOk(); ok {
		clusterResp, _, err := c.a.GetCluster(*v).Execute()
		if err == nil {
			return clusterResp.Data.Spec.GetName()
		}
	}
	return ""
}

func (c *DrContext) CreatedOn() string {
	if !c.c.GetInfo().Metadata.Get().HasCreatedOn() {
		return ""
	}
	return FormatDate(c.c.GetInfo().Metadata.Get().GetCreatedOn())
}

func (c *DrContext) Databases() string {
	if v, ok := c.c.Spec.GetDatabaseIdsOk(); ok {
		namespacesResp, _, err := c.a.GetClusterNamespaces(c.c.Info.GetSourceClusterId()).Execute()
		if err == nil {
			dbIdToNameMap := map[string]string{}
			for _, namespace := range namespacesResp.Data {
				dbIdToNameMap[namespace.GetId()] = namespace.GetName()
			}
			databaseNames := []string{}
			for _, databaseId := range *v {
				if databaseName, exists := dbIdToNameMap[databaseId]; exists {
					databaseNames = append(databaseNames, databaseName)
				} else {
					continue
				}
			}
			return strings.Join(databaseNames, ",")
		}
	}
	return ""
}

func (c *DrContext) State() string {
	if _, ok := c.c.Info.GetStateOk(); ok {
		return string(*c.c.Info.State)
	}
	return ""
}

func (c *DrContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
