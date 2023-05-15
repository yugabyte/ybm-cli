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
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"text/template"

	"github.com/enescakir/emoji"
	"github.com/sirupsen/logrus"
	// "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/role"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultRoleListing = "table {{.Name}}\t{{.Description}}\t{{.IsUserDefined}}\t{{.UsersCount}}\t{{.ApiKeysCount}}"
	isUserDefinedHeader = "User Defined"
	usersCountHeader = "Users Count"
	apiKeysCountHeader = "API Keys Count"
	defaultFullRoleListing = "table {{.Name}}\t{{.ID}}\t{{.Description}}\t{{.IsUserDefined}}"
	defaultRoleUsersListing = "table {{.UserEmail}}\t{{.UserFirstName}}\t{{.UserLastName}}\t{{.UserState}}"
	userEmailHeader = "Email"
	userFirstNameHeader = "First Name"
	userLastNameHeader = "Last Name"
	userStateHeader = "User State"
	defaultRoleApiKeysListing = "table {{.ApiKeyName}}\t{{.ApiKeyIssuer}}\t{{.ApiKeyStatus}}"
	apiKeyNameHeader = "API Key Name"
	apiKeyIssuerHeader = "Issuer"
	apiKeyStatusHeader  = "Status"
)

type RoleContext struct {
	HeaderContext
	Context
	r ybmclient.RoleData
}

type FullRoleContext struct {
	HeaderContext
	Context
	fullRole *role.FullRole
}

// NewRoleContext creates a new context for rendering cluster
func NewRoleContext() *RoleContext {
	roleCtx := RoleContext{}
	roleCtx.Header = SubHeaderContext{
		"Name":             nameHeader,
		"ID":				"ID",
		"Description":      descriptionHeader,
		"IsUserDefined":    isUserDefinedHeader,
		"UsersCount": 		usersCountHeader,
		"ApiKeysCount":		apiKeysCountHeader,
	}
	return &roleCtx
}

// func NewFullRoleContext() *FullRoleContext {
// 	fullRoleCtx := FullRoleContext{}
// 	fullRoleCtx.Header = SubHeaderContext{
// 		"Name":             nameHeader,
// 		"ID":               "ID",
// 		"Description":      descriptionHeader,
// 		"IsUserDefined":    isUserDefinedHeader,
// 	}
// 	return &fullRoleCtx
// }

func NewFullRoleContext() *FullRoleContext {
	roleCtx := FullRoleContext{}
	roleCtx.Header = SubHeaderContext{
	"Name":             nameHeader,
	"ID":				"ID",
	"Description":      descriptionHeader,
	"IsUserDefined":    isUserDefinedHeader,}
	return &roleCtx
}

func NewRoleFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultRoleListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewFullRoleFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultFullRoleListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// ClusterWrite renders the context for a list of clusters
func RoleWrite(ctx Context, roles []ybmclient.RoleData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, role := range roles {
			err := format(&RoleContext{r: role})
			if err != nil {
				logrus.Debugf("Error rendering role: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewRoleContext(), render)
}

func SingleRoleWrite(ctx Context, role ybmclient.RoleData) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&RoleContext{r: role})
		//, p: role.Spec.Permissions, ep: role.Info.GetEffectivePermissions()
		if err != nil {
			logrus.Debugf("Error rendering role: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewFullRoleContext(), render)
}


func (r *RoleContext) ID() string {
	return r.r.Info.Id
}

// func (r *FullRoleContext) ID() string {
// 	return r.r.Info.Id
// }

func (r *RoleContext) Name() string {
	return r.r.Info.GetDisplayName()
}

// func (r *FullRoleContext) Name() string {
// 	return r.r.Info.GetDisplayName()
// }

func (r *RoleContext) Description() string {
	return r.r.GetDescription()
}

// func (r *FullRoleContext) Description() string {
// 	return r.r.GetDescription()
// }

func (r *RoleContext) IsUserDefined() string {
	isUserDefined := r.r.Info.GetIsUserDefined()
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%t", isUserDefined)
	}
	switch isUserDefined {
	case true:
		return emoji.Parse(":white_check_mark:")
	case false:
		return emoji.CrossMark.String()
	default:
		return fmt.Sprintf("%t", isUserDefined)
	}
}

// func (r *FullRoleContext) IsUserDefined() string {
// 	isUserDefined := r.r.Info.GetIsUserDefined()
// 	if runtime.GOOS == "windows" {
// 		return fmt.Sprintf("%t", isUserDefined)
// 	}
// 	switch isUserDefined {
// 	case true:
// 		return emoji.Parse(":white_check_mark:")
// 	case false:
// 		return emoji.CrossMark.String()
// 	default:
// 		return fmt.Sprintf("%t", isUserDefined)
// 	}
// }

func (r *RoleContext) UsersCount() int {
	return len(r.r.Info.GetUsers())
}

func (r *RoleContext) ApiKeysCount() int {
	return len(r.r.Info.GetApiKeys())
}

func (r *RoleContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}


// const (
// 	psEndpointAccess         = "table {{.Az}}\t{{.SecurityPrincipals}}"
// 	psEndpointConn           = "table {{.ServiceName}}\t{{.State}}\t{{.ActiveConnections}}"
// 	azHeader                 = "Availability Zones"
// 	securityPrincipalsHeader = "Security Principals"
// 	serviceNameHeader        = "Service Name"
// 	activeConnectionsHeader  = "Active Connections"
// )

// type PSEndpointFullContext struct {
// 	HeaderContext
// 	Context
// 	psEndpoint ybmclient.PrivateServiceEndpointRegionData
// 	endpoint   ybmclient.Endpoint
// }

// func NewPSEndpointContext() *PSEndpointFullContext {
// 	psEndpointCtx := PSEndpointFullContext{}
// 	psEndpointCtx.Header = SubHeaderContext{
// 		"Az":                 azHeader,
// 		"SecurityPrincipals": securityPrincipalsHeader,
// 		"ServiceName":        serviceNameHeader,
// 		"State":              stateHeader,
// 		"ActiveConnections":  activeConnectionsHeader,
// 	}
// 	return &psEndpointCtx
// }

// func NewPSEndpointFormat(source string) Format {
// 	switch source {
// 	case "table", "":
// 		format := defaultEndpointListing
// 		return Format(format)
// 	default: // custom format or json or pretty
// 		return Format(source)
// 	}
// }

// func FullRoleWrite(ctx Context, role ybmclient.RoleData) error {
// 	render := func(format func(subContext SubContext) error) error {
// 		err := format(&FullRoleContext{r: role, p: role.Spec.Permissions, ep: role.Info.GetEffectivePermissions()})
// 		if err != nil {
// 			logrus.Debugf("Error while rendering role: %v", err)
// 			return err
// 		}
// 		return nil
// 	}
// 	return ctx.Write(NewFullRoleContext(), render)
// }

// func SingleRoleWrite(ctx Context, role ybmclient.RoleData) error {
// 	render := func(format func(subContext SubContext) error) error {
// 		err := format(&RoleContext{r: role})
// 		//, p: role.Spec.Permissions, ep: role.Info.GetEffectivePermissions()
// 		if err != nil {
// 			logrus.Debugf("Error rendering role: %v", err)
// 			return err
// 		}
// 		return nil
// 	}
// 	return ctx.Write(NewRoleContext(), render)
// }

// func (fr *FullRoleContext) MarshalJSON() ([]byte, error) {
// 	// r ybmclient.RoleData
// 	// p []ybmclient.ResourcePermissionInfo
// 	// ep []ybmclient.ResourcePermissionInfo
// 	return json.Marshal(struct {
// 		Role   ybmclient.RoleData
// 		Permissions []ybmclient.ResourcePermissionInfo
// 		EffectivePermissions []ybmclient.ResourcePermissionInfo
// 	}{
// 		Role:   fr.r,
// 		Permissions: fr.p,
// 		EffectivePermissions: fr.ep,
// 	})
// }

func (r *FullRoleContext) SetFullRole(roleData ybmclient.RoleData) {
	fr:= role.NewFullRole(roleData)
	r.fullRole = fr
}

// func (c *FullClusterContext) SetFullCluster(authApi client.AuthApiClient, clusterData ybmclient.ClusterData) {
// 	fc := cluster.NewFullCluster(authApi, clusterData)
// 	c.fullCluster = fc
// }

func (fr *FullRoleContext) startSubsection(format string) (*template.Template, error) {
	fr.buffer = bytes.NewBufferString("")
	fr.header = ""
	fr.Format = Format(format)
	fr.preFormat()

	return fr.parseFormat()
}

// type fullClusterContext struct {
// 	Cluster         *ClusterContext
// 	VPCContext      []*VPCContext
// 	CIRContext      []*clusterInfoRegionsContext
// 	NalContext      []*NetworkAllowListContext
// 	EndpointContext []*EndpointContext
// 	NodeContext     []*NodeContext
// 	CmkContext      []*CMKContext
// }

type fullRoleContext struct {
	Role         	*RoleContext
	PermissionsContext      []*resourcePermissionContext
	EffectivePermissionsContext      []*resourcePermissionContext
	RoleUsersContext			[]*roleUsersContext
	RoleApiKeysContext			[]*roleApiKeysContext
}

type resourcePermissionContext struct {
	HeaderContext
	r ybmclient.ResourcePermissionInfo
	opsIndex int
}

func NewResourcePermissionFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaulRolePermissionListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewResourcePermissionContext() *resourcePermissionContext {
	rpCtx := resourcePermissionContext{}
	rpCtx.Header = SubHeaderContext{
		"ResourceName":  resourceNameHeader,
		"ResourceType":  resourceTypeHeader,
		"OperationDescription":  operationDescriptionHeader,
		"OperationType":  operationTypeHeader,
	}
	return &rpCtx
}

func (r *resourcePermissionContext) ResourceName() string {
	return fmt.Sprintf("%s", r.r.GetResourceName())
}

func (r *resourcePermissionContext) ResourceType() string {
	return fmt.Sprintf("%s", r.r.GetResourceType())
}

func (r *resourcePermissionContext) OperationDescription() string {
	return fmt.Sprintf("%s", r.r.OperationGroups[r.opsIndex].GetOperationGroupDescription())
}

func (r *resourcePermissionContext) OperationType() string {
	return fmt.Sprintf("%s", r.r.OperationGroups[r.opsIndex].GetOperationGroup())
}


func (r *resourcePermissionContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}

type roleUsersContext struct {
	HeaderContext
	r ybmclient.UserSpecWithStateInfo
}

func NewRoleUsersContext() *roleUsersContext {
	roleUsersCtx := roleUsersContext{}
	roleUsersCtx.Header = SubHeaderContext{
		"UserEmail":  userEmailHeader,
		"UserFirstName":  userFirstNameHeader,
		"UserLastName":  userLastNameHeader,
		"UserState":  userStateHeader,
	}
	return &roleUsersCtx
}

func (r *roleUsersContext) UserEmail() string {
	return fmt.Sprintf("%s", r.r.GetEmail())
}

func (r *roleUsersContext) UserFirstName() string {
	return fmt.Sprintf("%s", r.r.GetFirstName())
}

func (r *roleUsersContext) UserLastName() string {
	return fmt.Sprintf("%s", r.r.GetLastName())
}

func (r *roleUsersContext) UserState() string {
	return fmt.Sprintf("%s", r.r.GetState())
}


func (r *roleUsersContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}

type roleApiKeysContext struct {
	HeaderContext
	r ybmclient.ApiKeyBasicInfo
}

func NewRoleApiKeysContext() *roleApiKeysContext {
	roleApiKeysCtx := roleApiKeysContext{}
	roleApiKeysCtx.Header = SubHeaderContext{
		"ApiKeyName":  apiKeyNameHeader,
		"ApiKeyIssuer":  apiKeyIssuerHeader,
		"ApiKeyStatus":  apiKeyStatusHeader,
	}
	return &roleApiKeysCtx
}

func (r *roleApiKeysContext) ApiKeyName() string {
	return fmt.Sprintf("%s", r.r.GetName())
}

func (r *roleApiKeysContext) ApiKeyIssuer() string {
	return fmt.Sprintf("%s", r.r.GetIssuer())
}

func (r *roleApiKeysContext) ApiKeyStatus() string {
	return fmt.Sprintf("%s", r.r.GetStatus())
}


func (r *roleApiKeysContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}

func (r *FullRoleContext) SubSection(name string) {
	r.Output.Write([]byte("\n\n"))
	r.Output.Write([]byte(Colorize(name, GREEN_COLOR)))
	r.Output.Write([]byte("\n"))
}


func (r *FullRoleContext) Write() error {
	frc := &fullRoleContext{
		Role:         &RoleContext{},
		PermissionsContext:      make([]*resourcePermissionContext, 0, len(r.fullRole.Permissions)),
		EffectivePermissionsContext:      make([]*resourcePermissionContext, 0, len(r.fullRole.EffectivePermissions)),
		RoleUsersContext:      make([]*roleUsersContext, 0, len(r.fullRole.RoleUsers)),
		RoleApiKeysContext: make([]*roleApiKeysContext, 0, len(r.fullRole.RoleApiKeys)),
	}

	frc.Role.r = r.fullRole.Role

	//Adding Permissions information
	for _, permission := range r.fullRole.Permissions {
		for i := 0 ; i < len(permission.OperationGroups); i++ { 
			frc.PermissionsContext = append(frc.PermissionsContext, &resourcePermissionContext{r: permission, opsIndex: i})
		}
	}

	//Adding Effective Permissions information
	for _, effectivePermission := range r.fullRole.EffectivePermissions {
		for i := 0 ; i < len(effectivePermission.OperationGroups); i++ { 
			frc.EffectivePermissionsContext = append(frc.EffectivePermissionsContext, &resourcePermissionContext{r: effectivePermission, opsIndex: i})
		}
	}

	//Adding Users information
	for _, user := range r.fullRole.RoleUsers {
		frc.RoleUsersContext = append(frc.RoleUsersContext, &roleUsersContext{r: user})
	}

	//Adding Api Keys information
	for _, apiKey := range r.fullRole.RoleApiKeys {
		frc.RoleApiKeysContext = append(frc.RoleApiKeysContext, &roleApiKeysContext{r: apiKey})
	}

	//First Section
	tmpl, err := r.startSubsection(defaultFullRoleListing)
	if err != nil {
		return err
	}
	r.Output.Write([]byte(Colorize("General", GREEN_COLOR)))
	r.Output.Write([]byte("\n"))
	if err := r.contextFormat(tmpl, frc.Role); err != nil {
		return err
	}
	r.postFormat(tmpl, NewFullRoleContext())


	//Permissions Subsection
	tmpl, err = r.startSubsection(defaulRolePermissionListing)
	if err != nil {
		return err
	}
	r.SubSection("Permissions")
	for _, v := range frc.PermissionsContext {
		if err := r.contextFormat(tmpl, v); err != nil {
			return err
		}
	}
	r.postFormat(tmpl, NewResourcePermissionContext())

	//Permissions Subsection
	tmpl, err = r.startSubsection(defaulRolePermissionListing)
	if err != nil {
		return err
	}
	r.SubSection("Effective Permissions")
	for _, v := range frc.EffectivePermissionsContext {
		if err := r.contextFormat(tmpl, v); err != nil {
			return err
		}
	}
	r.postFormat(tmpl, NewResourcePermissionContext())



	// Role Users
	if len(frc.RoleUsersContext) > 0 {
		tmpl, err = r.startSubsection(defaultRoleUsersListing)
		if err != nil {
			return err
		}
		r.SubSection("Role Users")
		for _, v := range frc.RoleUsersContext {
			if err := r.contextFormat(tmpl, v); err != nil {
				return err
			}
		}
		r.postFormat(tmpl, NewRoleUsersContext())
	}

	// Role Api Keys
	if len(frc.RoleApiKeysContext) > 0 {
		tmpl, err = r.startSubsection(defaultRoleApiKeysListing)
		if err != nil {
			return err
		}
		r.SubSection("Role API Keys")
		for _, v := range frc.RoleApiKeysContext {
			if err := r.contextFormat(tmpl, v); err != nil {
				return err
			}
		}
		r.postFormat(tmpl, NewRoleApiKeysContext())
	}

	return nil
}

// // NewFullClusterContext creates a new context for rendering cluster
// func NewFullClusterContext() *FullClusterContext {
// 	clusterCtx := FullClusterContext{}
// 	clusterCtx.Header = SubHeaderContext{}
// 	return &clusterCtx
// }

// func (fr *FullRoleContext) Write() error {
// 	frContext := &FullRoleContext{}
// 	frContext.r = fr.r

// 	// Role listing
// 	tmpl, err := fr.startSubsection(defaultFullRoleListing)
// 	if err != nil {
// 		return err
// 	}
// 	fr.Output.Write([]byte(Colorize("General", GREEN_COLOR)))
// 	fr.Output.Write([]byte("\n"))
// 	if err := fr.contextFormat(tmpl, frContext); err != nil {
// 		return err
// 	}
// 	fr.postFormat(tmpl, NewFullRoleContext())

// 	// Permissions
// 	tmpl, err = ep.startSubsection(psEndpointAccess)
// 	if err != nil {
// 		return err
// 	}
// 	ep.Output.Write([]byte(Colorize("Access", GREEN_COLOR)))
// 	ep.Output.Write([]byte("\n"))
// 	if err := ep.contextFormat(tmpl, ep); err != nil {
// 		return err
// 	}
// 	ep.postFormat(tmpl, ep)

// 	// PSEndpoint connectivity
// 	tmpl, err = ep.startSubsection(psEndpointConn)
// 	if err != nil {
// 		return err
// 	}
// 	ep.Output.Write([]byte(Colorize("Connectivity", GREEN_COLOR)))
// 	ep.Output.Write([]byte("\n"))
// 	if err := ep.contextFormat(tmpl, ep); err != nil {
// 		return err
// 	}
// 	ep.postFormat(tmpl, ep)

// 	return nil
// }

// func (ep *PSEndpointFullContext) Az() string {
// 	return strings.Join(ep.psEndpoint.Info.AvailabilityZones, ",")
// }

// func (ep *PSEndpointFullContext) SecurityPrincipals() string {
// 	return strings.Join(ep.psEndpoint.Spec.SecurityPrincipals, ",")
// }

// func (ep *PSEndpointFullContext) ServiceName() string {
// 	return ep.psEndpoint.Info.GetServiceName()
// }

// func (ep *PSEndpointFullContext) State() string {
// 	return string(ep.psEndpoint.Info.State)
// }

// func (ep *PSEndpointFullContext) ActiveConnections() string {
// 	return strings.Join(ep.psEndpoint.Info.GetConnections(), ",")
//bc7f5141-a7a5-4eb5-a054-bff8ff97796e
// }


