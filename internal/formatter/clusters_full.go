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

package formatter

import (
	"bytes"
	"text/template"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultFullClusterGeneral  = "table {{.Name}}\t{{.SoftwareVersion}}\t{{.State}}\t{{.HealthState}}\t{{.Regions}}\t{{.Nodes}}\t{{.NodesSpec}}"
	defaultFullClusterGeneral2 = "table {{.Provider}}\t{{.FaultTolerance}}\t{{.DataDistribution}}"
	faultToleranceHeader       = "Fault Tolerance"
	dataDistributionHeader     = "Data Distribution"
)

type FullClusterContext struct {
	HeaderContext
	Context
	c ybmclient.ClusterData
}

func (c *FullClusterContext) SetClusterData(cd ybmclient.ClusterData) {
	c.c = cd
}

func (c *FullClusterContext) startSubsection(format string) (*template.Template, error) {
	c.buffer = bytes.NewBufferString("")
	c.header = ""
	c.Format = Format(format)
	c.preFormat()

	return c.parseFormat()
}

func NewFullClusterFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultClusterListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

type fullClusterContext struct {
	Cluster    *ClusterContext
	VPCContext *VPCContext
}

func (c *FullClusterContext) Write() error {
	fcc := &fullClusterContext{
		Cluster:    &ClusterContext{},
		VPCContext: &VPCContext{},
	}
	fcc.Cluster.c = c.c

	tmpl, err := c.startSubsection(defaultFullClusterGeneral)
	if err != nil {
		return err
	}
	c.Output.Write([]byte("General:\n\n"))
	if err := c.contextFormat(tmpl, fcc.Cluster); err != nil {
		return err
	}
	c.postFormat(tmpl, NewClusterContext())

	tmpl, err = c.startSubsection(defaultFullClusterGeneral2)
	if err != nil {
		return err
	}
	c.Output.Write([]byte("\n\n"))
	if err := c.contextFormat(tmpl, fcc.Cluster); err != nil {
		return err
	}
	c.postFormat(tmpl, NewClusterContext())
	return nil
}

// NewFullClusterContext creates a new context for rendering cluster
func NewFullClusterContext() *FullClusterContext {
	clusterCtx := FullClusterContext{}
	clusterCtx.Header = SubHeaderContext{}
	return &clusterCtx
}
