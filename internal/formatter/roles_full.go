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
	"sort"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/yugabyte/ybm-cli/internal/role"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultFullRoleListing    = "table {{.Name}}\t{{.ID}}\t{{.Description}}\t{{.RoleType}}"
	defaultRoleUsersListing   = "table {{.UserEmail}}\t{{.UserName}}\t{{.UserState}}"
	userEmailHeader           = "Email"
	userNameHeader            = "Display Name"
	userStateHeader           = "Status"
	defaultRoleApiKeysListing = "table {{.ApiKeyName}}\t{{.ApiKeyIssuer}}\t{{.ApiKeyStatus}}"
	apiKeyNameHeader          = "Name"
	apiKeyIssuerHeader        = "Created By"
	apiKeyStatusHeader        = "Status"
)

type FullRoleContext struct {
	HeaderContext
	Context
	fullRole *role.FullRole
}

// NewFullRoleContext creates a new context for rendering all role details
func NewFullRoleContext() *FullRoleContext {
	roleCtx := FullRoleContext{}
	roleCtx.Header = SubHeaderContext{
		"Name":        nameHeader,
		"ID":          "ID",
		"Description": descriptionHeader,
		"RoleType":    roleTypeHeader,
	}
	return &roleCtx
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

// SingleRoleWrite renders the context for a single role
func SingleRoleWrite(ctx Context, role ybmclient.RoleData) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&RoleContext{r: role})
		if err != nil {
			logrus.Debugf("Error rendering role: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewFullRoleContext(), render)
}

func (r *FullRoleContext) SetFullRole(roleData ybmclient.RoleData) {
	fr := role.NewFullRole(roleData)
	r.fullRole = fr
}

func (fr *FullRoleContext) startSubsection(format string) (*template.Template, error) {
	fr.buffer = bytes.NewBufferString("")
	fr.header = ""
	fr.Format = Format(format)
	fr.preFormat()

	return fr.parseFormat()
}

type fullRoleContext struct {
	Role                        *RoleContext
	PermissionsContext          []*rolePermissionContext
	EffectivePermissionsContext []*rolePermissionContext
	RoleUsersContext            []*roleUsersContext
	RoleApiKeysContext          []*roleApiKeysContext
}

type rolePermissionContext struct {
	HeaderContext
	r        ybmclient.ResourcePermissionInfo
	opsIndex int
}

func NewRolePermissionFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultResourcePermissionListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewRolePermissionContext() *rolePermissionContext {
	rpCtx := rolePermissionContext{}
	rpCtx.Header = SubHeaderContext{
		"ResourceName":         resourceNameHeader,
		"ResourceType":         resourceTypeHeader,
		"OperationDescription": operationDescriptionHeader,
		"OperationType":        operationTypeHeader,
	}
	return &rpCtx
}

func (r *rolePermissionContext) ResourceName() string {
	return r.r.GetResourceName()
}

func (r *rolePermissionContext) ResourceType() string {
	return string(r.r.GetResourceType())
}

func (r *rolePermissionContext) OperationDescription() string {
	return r.r.OperationGroups[r.opsIndex].GetOperationGroupDescription()
}

func (r *rolePermissionContext) OperationType() string {
	return string(r.r.OperationGroups[r.opsIndex].GetOperationGroup())
}

func (r *rolePermissionContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}

type roleUsersContext struct {
	HeaderContext
	r ybmclient.UserSpecWithStateInfo
}

func NewRoleUsersContext() *roleUsersContext {
	roleUsersCtx := roleUsersContext{}
	roleUsersCtx.Header = SubHeaderContext{
		"UserEmail": userEmailHeader,
		"UserName":  userNameHeader,
		"UserState": userStateHeader,
	}
	return &roleUsersCtx
}

func (r *roleUsersContext) UserEmail() string {
	return r.r.GetEmail()
}

func (r *roleUsersContext) UserName() string {
	return fmt.Sprintf("%s %s", r.r.GetFirstName(), r.r.GetLastName())
}

func (r *roleUsersContext) UserState() string {
	return string(r.r.GetState())
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
		"ApiKeyName":   apiKeyNameHeader,
		"ApiKeyIssuer": apiKeyIssuerHeader,
		"ApiKeyStatus": apiKeyStatusHeader,
	}
	return &roleApiKeysCtx
}

func (r *roleApiKeysContext) ApiKeyName() string {
	return r.r.GetName()
}

func (r *roleApiKeysContext) ApiKeyIssuer() string {
	return r.r.GetIssuer()
}

func (r *roleApiKeysContext) ApiKeyStatus() string {
	return string(r.r.GetStatus())
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
		Role:                        &RoleContext{},
		PermissionsContext:          make([]*rolePermissionContext, 0, len(r.fullRole.Permissions)),
		EffectivePermissionsContext: make([]*rolePermissionContext, 0, len(r.fullRole.EffectivePermissions)),
		RoleUsersContext:            make([]*roleUsersContext, 0, len(r.fullRole.RoleUsers)),
		RoleApiKeysContext:          make([]*roleApiKeysContext, 0, len(r.fullRole.RoleApiKeys)),
	}

	frc.Role.r = r.fullRole.Role

	//Adding Permissions information
	sort.Slice(r.fullRole.Permissions, func(i, j int) bool {
		return string(r.fullRole.Permissions[i].ResourceType) < string(r.fullRole.Permissions[j].ResourceType)
	})
	for _, permission := range r.fullRole.Permissions {
		for i := 0; i < len(permission.OperationGroups); i++ {
			frc.PermissionsContext = append(frc.PermissionsContext, &rolePermissionContext{r: permission, opsIndex: i})
		}
	}

	//Adding Effective Permissions information
	sort.Slice(r.fullRole.EffectivePermissions, func(i, j int) bool {
		return string(r.fullRole.EffectivePermissions[i].ResourceType) < string(r.fullRole.EffectivePermissions[j].ResourceType)
	})
	for _, effectivePermission := range r.fullRole.EffectivePermissions {
		for i := 0; i < len(effectivePermission.OperationGroups); i++ {
			frc.EffectivePermissionsContext = append(frc.EffectivePermissionsContext, &rolePermissionContext{r: effectivePermission, opsIndex: i})
		}
	}

	//Adding Users information
	for _, user := range r.fullRole.RoleUsers {
		frc.RoleUsersContext = append(frc.RoleUsersContext, &roleUsersContext{r: user})
	}

	//Adding account API keys information
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
	tmpl, err = r.startSubsection(defaultResourcePermissionListing)
	if err != nil {
		return err
	}
	r.SubSection("Permissions")
	for _, v := range frc.PermissionsContext {
		if err := r.contextFormat(tmpl, v); err != nil {
			return err
		}
	}
	r.postFormat(tmpl, NewRolePermissionContext())

	//Effective Permissions Subsection
	tmpl, err = r.startSubsection(defaultResourcePermissionListing)
	if err != nil {
		return err
	}
	r.SubSection("Effective Permissions")
	for _, v := range frc.EffectivePermissionsContext {
		if err := r.contextFormat(tmpl, v); err != nil {
			return err
		}
	}
	r.postFormat(tmpl, NewRolePermissionContext())

	// Role Users
	if len(frc.RoleUsersContext) > 0 {
		tmpl, err = r.startSubsection(defaultRoleUsersListing)
		if err != nil {
			return err
		}
		r.SubSection("Users")
		for _, v := range frc.RoleUsersContext {
			if err := r.contextFormat(tmpl, v); err != nil {
				return err
			}
		}
		r.postFormat(tmpl, NewRoleUsersContext())
	}

	// Role account API keys
	if len(frc.RoleApiKeysContext) > 0 {
		tmpl, err = r.startSubsection(defaultRoleApiKeysListing)
		if err != nil {
			return err
		}
		r.SubSection("API Keys")
		for _, v := range frc.RoleApiKeysContext {
			if err := r.contextFormat(tmpl, v); err != nil {
				return err
			}
		}
		r.postFormat(tmpl, NewRoleApiKeysContext())
	}

	return nil
}
