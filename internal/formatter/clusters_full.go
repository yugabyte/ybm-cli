// Copyright (c) YugaByte, Inc.
//
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
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022-present Yugabyte, Inc.

package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"text/template"

	"github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/cluster"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultFullClusterGeneral       = "table {{.Name}}\t{{.ID}}\t{{.SoftwareVersion}}\t{{.State}}\t{{.HealthState}}"
	defaultFullClusterGeneral2      = "table {{.Provider}}\t{{.FaultTolerance}}\t{{.DataDistribution}}\t{{.Nodes}}\t{{.NodesSpec}}"
	defaultVPCListingCluster        = "table {{.Name}}\t{{.State}}\t{{.Provider}}\t{{.Regions}}\t{{.CIDR}}\t{{.Peerings}}"
	defaultDefaultFullClusterRegion = "table {{.Region}}\t{{.NumNode}}\t{{.NumCores}}\t{{.MemoryGb}}\t{{.DiskSizeGb}}\t{{.VpcName}}"
	defaultFullClusterNalListing    = "table {{.Name}}\t{{.Desc}}\t{{.AllowedList}}"
	faultToleranceHeader            = "Fault Tolerance"
	dataDistributionHeader          = "Data Distribution"
	vcpuByNodeHeader                = "vCPU/Node"
	memoryByNodeHeader              = "Mem/Node"
	diskByNodeHeader                = "Disk/Node"
)

type FullClusterContext struct {
	HeaderContext
	Context
	fullCluster *cluster.FullCluster
}

func (c *FullClusterContext) SetFullCluster(authApi client.AuthApiClient, clusterData ybmclient.ClusterData) {
	fc := cluster.NewFullCluster(authApi, clusterData)
	c.fullCluster = fc
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
	VPCContext []*VPCContext
	CIRContext []*clusterInfoRegionsContext
	NalContext []*NetworkAllowListContext
}

func (c *FullClusterContext) Write() error {
	fcc := &fullClusterContext{
		Cluster:    &ClusterContext{},
		VPCContext: make([]*VPCContext, 0, len(c.fullCluster.Vpc)),
		CIRContext: make([]*clusterInfoRegionsContext, 0, len(c.fullCluster.Cluster.Spec.ClusterRegionInfo)),
		NalContext: make([]*NetworkAllowListContext, 0, len(c.fullCluster.AllowList)),
	}

	fcc.Cluster.c = c.fullCluster.Cluster

	//Adding VPC information
	for _, vpc := range c.fullCluster.Vpc {
		fcc.VPCContext = append(fcc.VPCContext, &VPCContext{c: vpc})
	}

	//Adding Regions - Node information
	sort.Slice(c.fullCluster.Cluster.GetSpec().ClusterRegionInfo, func(i, j int) bool {
		return string(c.fullCluster.Cluster.GetSpec().ClusterRegionInfo[i].PlacementInfo.CloudInfo.Region) < string(c.fullCluster.Cluster.GetSpec().ClusterRegionInfo[j].PlacementInfo.CloudInfo.Region)
	})
	for _, cir := range c.fullCluster.Cluster.GetSpec().ClusterRegionInfo {
		fcc.CIRContext = append(fcc.CIRContext,
			&clusterInfoRegionsContext{
				clusterInfoRegion: cir,
				clusterInfo:       c.fullCluster.Cluster.GetSpec().ClusterInfo,
				vpcName:           c.fullCluster.Vpc[cir.PlacementInfo.GetVpcId()].Spec.Name,
			})
	}

	//Adding AllowList
	for _, nal := range c.fullCluster.AllowList {
		fcc.NalContext = append(fcc.NalContext, &NetworkAllowListContext{c: nal})
	}

	//First Section
	tmpl, err := c.startSubsection(defaultFullClusterGeneral)
	if err != nil {
		return err
	}
	c.Output.Write([]byte(Colorize("General", GREEN_COLOR)))
	c.Output.Write([]byte("\n"))
	if err := c.contextFormat(tmpl, fcc.Cluster); err != nil {
		return err
	}
	c.postFormat(tmpl, NewClusterContext())

	tmpl, err = c.startSubsection(defaultFullClusterGeneral2)
	if err != nil {
		return err
	}
	c.Output.Write([]byte("\n"))
	if err := c.contextFormat(tmpl, fcc.Cluster); err != nil {
		return err
	}
	c.postFormat(tmpl, NewClusterContext())

	//NAL subsection if any
	if len(fcc.NalContext) > 0 {
		tmpl, err = c.startSubsection(defaultFullClusterNalListing)
		if err != nil {
			return err
		}
		c.SubSection("Network AllowList")
		for _, v := range fcc.NalContext {
			if err := c.contextFormat(tmpl, v); err != nil {
				return err
			}
		}
		c.postFormat(tmpl, NewNetworkAllowListContext())
	}

	//VPC subsection if any
	if len(fcc.VPCContext) > 0 {
		tmpl, err = c.startSubsection(defaultVPCListingCluster)
		if err != nil {
			return err
		}
		c.SubSection("VPC")
		for _, v := range fcc.VPCContext {
			if err := c.contextFormat(tmpl, v); err != nil {
				return err
			}
		}
		c.postFormat(tmpl, NewVPCContext())
	}

	//Regions Subsection
	tmpl, err = c.startSubsection(defaultDefaultFullClusterRegion)
	if err != nil {
		return err
	}
	c.SubSection("Regions")
	for _, v := range fcc.CIRContext {
		if err := c.contextFormat(tmpl, v); err != nil {
			return err
		}
	}
	c.postFormat(tmpl, NewClusterInfoRegionsContext())

	return nil
}

func (c *FullClusterContext) SubSection(name string) {
	c.Output.Write([]byte("\n\n"))
	c.Output.Write([]byte(Colorize(name, GREEN_COLOR)))
	c.Output.Write([]byte("\n"))
}

// NewFullClusterContext creates a new context for rendering cluster
func NewFullClusterContext() *FullClusterContext {
	clusterCtx := FullClusterContext{}
	clusterCtx.Header = SubHeaderContext{}
	return &clusterCtx
}

type clusterInfoRegionsContext struct {
	HeaderContext
	clusterInfoRegion ybmclient.ClusterRegionInfo
	clusterInfo       ybmclient.ClusterInfo
	vpcName           string
}

func NewClusterInfoRegionsContext() *clusterInfoRegionsContext {
	clusterCtx := clusterInfoRegionsContext{}
	clusterCtx.Header = SubHeaderContext{
		"Region":     "Region",
		"NumNode":    numNodesHeader,
		"NumCores":   vcpuByNodeHeader,
		"MemoryGb":   memoryByNodeHeader,
		"DiskSizeGb": diskByNodeHeader,
		"VpcName":    "VPC",
	}
	return &clusterCtx
}

func (c *clusterInfoRegionsContext) NumNode() string {
	return fmt.Sprintf("%d", c.clusterInfoRegion.GetPlacementInfo().NumNodes)
}

func (c *clusterInfoRegionsContext) NumCores() string {
	return fmt.Sprintf("%d", c.clusterInfo.NodeInfo.NumCores)
}

func (c *clusterInfoRegionsContext) MemoryGb() string {
	return convertMbtoGb(c.clusterInfo.NodeInfo.MemoryMb)
}

func (c *clusterInfoRegionsContext) DiskSizeGb() string {
	return fmt.Sprintf("%dGB", c.clusterInfo.NodeInfo.DiskSizeGb)
}

func (c *clusterInfoRegionsContext) Region() string {
	return c.clusterInfoRegion.GetPlacementInfo().CloudInfo.Region
}

func (c *clusterInfoRegionsContext) VpcName() string {
	return c.vpcName
}
func (c *clusterInfoRegionsContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.clusterInfoRegion)
}
