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
	"fmt"
	"strings"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultVPCListing   = "table {{.Name}}\t{{.State}}\t{{.Provider}}\t{{.RegionsCIDR}}\t{{.Peerings}}\t{{.Clusters}}"
	vpcRegionCIDRHeader = "Region[CIDR]"
	vpcCIDRHeader       = "CIDR"
	vpcPeeringHeader    = "Peerings"
)

type VPCContext struct {
	HeaderContext
	Context
	c ybmclient.SingleTenantVpcDataResponse
}

func NewVPCFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultVPCListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// VPCWrite renders the context for a list of VPCs
func VPCWrite(ctx Context, VPCs []ybmclient.SingleTenantVpcDataResponse) error {
	render := func(format func(subContext SubContext) error) error {
		for _, VPC := range VPCs {
			err := format(&VPCContext{c: VPC})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewVPCContext(), render)
}

// NewVPCContext creates a new context for rendering VPC
func NewVPCContext() *VPCContext {
	VPCCtx := VPCContext{}
	VPCCtx.Header = SubHeaderContext{
		"Name":        nameHeader,
		"State":       stateHeader,
		"RegionsCIDR": vpcRegionCIDRHeader,
		"Regions":     regionsHeader,
		"CIDR":        vpcCIDRHeader,
		"Provider":    providerHeader,
		"Peerings":    vpcPeeringHeader,
		"Clusters":    clustersHeader,
	}
	return &VPCCtx
}

func (c *VPCContext) Name() string {
	return c.c.Spec.Name
}

func (c *VPCContext) CIDR() string {
	var CIDRList []string
	for _, regionSpec := range c.c.GetSpec().RegionSpecs {
		CIDRList = append(CIDRList, regionSpec.GetCidr())
	}
	return strings.Join(CIDRList, ",")
}

func (c *VPCContext) Regions() string {
	var RegionsList []string
	for _, regionSpec := range c.c.GetSpec().RegionSpecs {
		RegionsList = append(RegionsList, regionSpec.GetRegion())
	}
	return strings.Join(RegionsList, ",")
}

func (c *VPCContext) RegionsCIDR() string {
	var RegionsCIDRList []string
	for _, regionSpec := range c.c.GetSpec().RegionSpecs {
		RegionsCIDRList = append(RegionsCIDRList, fmt.Sprintf("%s[%s]", regionSpec.GetRegion(), regionSpec.GetCidr()))
	}
	return strings.Join(RegionsCIDRList, ",")
}

func (c *VPCContext) State() string {
	if v, ok := c.c.Info.GetStateOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCContext) Provider() string {
	if v, ok := c.c.Spec.GetCloudOk(); ok {
		return string(*v.Ptr())
	}
	return ""
}

func (c *VPCContext) Peerings() string {
	if v, ok := c.c.Info.GetPeeringIdsOk(); ok {
		return fmt.Sprint(len(*v))
	}
	return ""
}

func (c *VPCContext) Clusters() string {
	if v, ok := c.c.Info.GetClusterIdsOk(); ok {
		return fmt.Sprint(len(*v))
	}
	return ""
}

func (c *VPCContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
