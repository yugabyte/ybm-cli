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

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultVPCPeeringListing = "table {{.Name}}\t{{.Provider}}\t{{.AppVPC}}\t{{.YbVPC}}\t{{.Status}}"
	appVPCHeader             = "Application VPC ID/Name"
	ybVPCHeader              = "YugabyteDB VPC Name"
	statusHeader             = "Status"
)

type VPCPeeringContext struct {
	HeaderContext
	Context
	c ybmclient.VpcPeeringData
}

func NewVPCPeeringFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultVPCPeeringListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// VPCPeeringWrite renders the context for a list of VPC Peerings
func VPCPeeringWrite(ctx Context, VPCPeerings []ybmclient.VpcPeeringData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, VPCPeering := range VPCPeerings {
			err := format(&VPCPeeringContext{c: VPCPeering})
			if err != nil {
				logrus.Debugf("Error rendering VPC Peering: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewVPCPeeringContext(), render)
}

// NewVPCPeeringContext creates a new context for rendering VPC Peering
func NewVPCPeeringContext() *VPCPeeringContext {
	VPCPeeringCtx := VPCPeeringContext{}
	VPCPeeringCtx.Header = SubHeaderContext{
		"Name":     nameHeader,
		"Provider": providerHeader,
		"AppVPC":   appVPCHeader,
		"YbVPC":    ybVPCHeader,
		"Status":   statusHeader,
	}
	return &VPCPeeringCtx
}

func (c *VPCPeeringContext) Name() string {
	if v, ok := c.c.Spec.GetNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) Provider() string {
	if v, ok := c.c.Spec.CustomerVpc.CloudInfo.GetCodeOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) AppVPC() string {
	if v, ok := c.c.Spec.CustomerVpc.GetExternalVpcIdOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) YbVPC() string {
	if v, ok := c.c.Info.GetYugabyteVpcNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) Status() string {
	if v, ok := c.c.Info.GetStateOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
