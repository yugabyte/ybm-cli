// // Licensed to Yugabyte, Inc. under one or more contributor license
// // agreements. See the NOTICE file distributed with this work for
// // additional information regarding copyright ownership. Yugabyte
// // licenses this file to you under the Apache License, Version 2.0
// // (the "License"); you may not use this file except in compliance
// // with the License.  You may obtain a copy of the License at
// // http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing,
// // software distributed under the License is distributed on an
// // "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// // KIND, either express or implied.  See the License for the
// // specific language governing permissions and limitations
// // under the License.

package formatter

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"sort"
// 	"text/template"

// 	"github.com/sirupsen/logrus"
// 	"github.com/yugabyte/ybm-cli/internal/client"
// 	"github.com/yugabyte/ybm-cli/internal/cluster"
// 	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
// )

// const (
// 	defaultFullRoleGeneral       = "table {{.Name}}\t{{.ID}}\t{{.Description}}"
// 	defaultFullRoleGeneral2      = "table {{.IsUserDefined}}\t{{.UsersCount}}\t{{.ApiKeysCount}}"
// 	defaultRolePermissionsListing		 = "table {{.ResourceName}}\t{{.ResourceType}}\t{{.OperationDescription}}\t{{.OperationType}}"
// 	resourceNameHeader         = "Resource Name"
// 	resourceTypeHeader         = "Resource Type"
// 	operationDescriptionHeader        = "Operation Description"
// 	operationTypeHeader        = "Operation Group Type"
// 	isUserDefinedHeader = "User Defined"
// 	usersCountHeader = "Users Count"
// 	apiKeysCountHeader = "API Keys Count"
// 	usersHeader = "Users"
// 	apiKeysHeader = "API Keys"
// )

// type FullRoleContext struct {
// 	HeaderContext
// 	Context
// 	fullRole *role.FullRole
// }

// func (r *FullRoleContext) SetFullRole(authApi client.AuthApiClient, roleData ybmclient.RoleData) {
// 	fr := role.NewFullRole(authApi, roleData)
// 	r.fullRole = fr
// }

// func (r *FullRoleContext) startSubsection(format string) (*template.Template, error) {
// 	c.buffer = bytes.NewBufferString("")
// 	c.header = ""
// 	c.Format = Format(format)
// 	c.preFormat()

// 	return c.parseFormat()
// }

// func NewFullRoleFormat(source string) Format {
// 	switch source {
// 	case "table", "":
// 		format := defaultRoleListing
// 		return Format(format)
// 	default: // custom format or json or pretty
// 		return Format(source)
// 	}
// }

// type fullRoleContext struct {
// 	Role         *RoleContext
// 	PermissionContext     []*RolePermissionContext
// 	EffectivePermissionContext     []*RolePermissionContext
// }

// func (r *FullRoleContext) Write() error {
// 	frc := &FullRoleContext{
// 		Role:         &RoleContext{},
// 		PermissionContext:      make([]*RolePermissionContext, 0, len(r.fullRole.Permissions)),
// 		EffectivePermissionContext:      make([]*RolePermissionContext, 0, len(r.fullRole.EffectivePermissions)),
// 	}

// 	frc.Role.c = r.fullRole.Role
// 	fr.PermissionContext.

// 	//Adding Permissions
// 	sort.Slice(r.fullRole.Permissions, func(i, j int) bool {
// 		return string(r.fullRole.Permissions[i].ResourceType) < string(c.Permissions[i].ResourceType)
// 	})
// 	for _, permission := range r.fullRole.Permissions {
// 		frc.PermissionContext = append(frc.PermissionContext, &NodeContext{n: node})
// 	}

// 	//First Section
// 	tmpl, err := c.startSubsection(defaultFullClusterGeneral)
// 	if err != nil {
// 		return err
// 	}
// 	c.Output.Write([]byte(Colorize("General", GREEN_COLOR)))
// 	c.Output.Write([]byte("\n"))
// 	if err := c.contextFormat(tmpl, fcc.Cluster); err != nil {
// 		return err
// 	}
// 	c.postFormat(tmpl, NewClusterContext())

// 	tmpl, err = c.startSubsection(defaultFullClusterGeneral2)
// 	if err != nil {
// 		return err
// 	}
// 	c.Output.Write([]byte("\n"))
// 	if err := c.contextFormat(tmpl, fcc.Cluster); err != nil {
// 		return err
// 	}
// 	c.postFormat(tmpl, NewClusterContext())

// 	//Regions Subsection
// 	tmpl, err = c.startSubsection(defaultDefaultFullClusterRegion)
// 	if err != nil {
// 		return err
// 	}
// 	c.SubSection("Regions")
// 	for _, v := range fcc.CIRContext {
// 		if err := c.contextFormat(tmpl, v); err != nil {
// 			return err
// 		}
// 	}
// 	c.postFormat(tmpl, NewClusterInfoRegionsContext())

// 	// Cluster endpoints
// 	if len(fcc.EndpointContext) > 0 {
// 		tmpl, err = c.startSubsection(defaultFullClusterEndpoints)
// 		if err != nil {
// 			return err
// 		}
// 		c.SubSection("Endpoints")
// 		for _, v := range fcc.EndpointContext {
// 			if err := c.contextFormat(tmpl, v); err != nil {
// 				return err
// 			}
// 		}
// 		c.postFormat(tmpl, NewEndpointContext())
// 	}

// 	//NAL subsection if any
// 	if len(fcc.NalContext) > 0 {
// 		tmpl, err = c.startSubsection(defaultFullClusterNalListing)
// 		if err != nil {
// 			return err
// 		}
// 		c.SubSection("Network AllowList")
// 		for _, v := range fcc.NalContext {
// 			if err := c.contextFormat(tmpl, v); err != nil {
// 				return err
// 			}
// 		}
// 		c.postFormat(tmpl, NewNetworkAllowListContext())
// 	}

// 	//VPC subsection if any
// 	if len(fcc.VPCContext) > 0 {
// 		tmpl, err = c.startSubsection(defaultVPCListingCluster)
// 		if err != nil {
// 			return err
// 		}
// 		c.SubSection("VPC")
// 		for _, v := range fcc.VPCContext {
// 			if err := c.contextFormat(tmpl, v); err != nil {
// 				return err
// 			}
// 		}
// 		c.postFormat(tmpl, NewVPCContext())
// 	}

// 	// CMK subsection if any
// 	if len(fcc.CmkContext) > 0 {
// 		tmpl, err = c.startSubsection(defaultFullClusterCMK)
// 		if err != nil {
// 			return err
// 		}
// 		c.SubSection("Encryption at Rest")
// 		for _, v := range fcc.CmkContext {
// 			if err := c.contextFormat(tmpl, v); err != nil {
// 				logrus.Fatal(err)
// 				return err
// 			}
// 		}
// 		c.postFormat(tmpl, NewCMKContext())
// 	}

// 	//Node subsection if any
// 	if len(fcc.NodeContext) > 0 {
// 		tmpl, err = c.startSubsection(defaultNodeListing)
// 		if err != nil {
// 			return err
// 		}
// 		c.SubSection("Nodes")
// 		for _, v := range fcc.NodeContext {
// 			if err := c.contextFormat(tmpl, v); err != nil {
// 				return err
// 			}
// 		}
// 		c.postFormat(tmpl, NewNodeContext())
// 	}

// 	return nil
// }

// func (c *FullRoleContext) SubSection(name string) {
// 	c.Output.Write([]byte("\n\n"))
// 	c.Output.Write([]byte(Colorize(name, GREEN_COLOR)))
// 	c.Output.Write([]byte("\n"))
// }

// // NewFullRoleContext creates a new context for rendering cluster
// func NewFullRoleContext() *FullRoleContext {
// 	clusterCtx := FullRoleContext{}
// 	clusterCtx.Header = SubHeaderContext{}
// 	return &clusterCtx
// }

// type clusterInfoRegionsContext struct {
// 	HeaderContext
// 	clusterInfoRegion ybmclient.ClusterRegionInfo
// 	clusterInfo       ybmclient.ClusterInfo
// 	vpcName           string
// }

// func NewClusterInfoRegionsContext() *clusterInfoRegionsContext {
// 	clusterCtx := clusterInfoRegionsContext{}
// 	clusterCtx.Header = SubHeaderContext{
// 		"Region":     "Region",
// 		"NumNode":    numNodesHeader,
// 		"NumCores":   vcpuByNodeHeader,
// 		"MemoryGb":   memoryByNodeHeader,
// 		"DiskSizeGb": diskByNodeHeader,
// 		"VpcName":    "VPC",
// 	}
// 	return &clusterCtx
// }

// func (c *clusterInfoRegionsContext) NumNode() string {
// 	return fmt.Sprintf("%d", c.clusterInfoRegion.GetPlacementInfo().NumNodes)
// }

// func (c *clusterInfoRegionsContext) NumCores() string {
// 	return fmt.Sprintf("%d", c.clusterInfo.NodeInfo.NumCores)
// }

// func (c *clusterInfoRegionsContext) MemoryGb() string {
// 	return convertMbtoGb(c.clusterInfo.NodeInfo.MemoryMb)
// }

// func (c *clusterInfoRegionsContext) DiskSizeGb() string {
// 	return fmt.Sprintf("%dGB", c.clusterInfo.NodeInfo.DiskSizeGb)
// }

// func (c *clusterInfoRegionsContext) Region() string {
// 	return c.clusterInfoRegion.GetPlacementInfo().CloudInfo.Region
// }

// func (c *clusterInfoRegionsContext) VpcName() string {
// 	return c.vpcName
// }
// func (c *clusterInfoRegionsContext) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(c.clusterInfoRegion)
// }
