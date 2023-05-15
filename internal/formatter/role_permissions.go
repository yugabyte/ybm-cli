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
	"fmt"
	"encoding/json"

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaulRolePermissionListing = "table {{.ResourceName}}\t{{.OperationDescription}}\t{{.ResourceType}}\t{{.OperationType}}"
	resourceNameHeader         = "Resource Name"
	resourceTypeHeader         = "Resource Type"
	operationDescriptionHeader        = "Operation Description"
	operationTypeHeader        = "Operation Group Type"
)

type RolePermissionContext struct {
	HeaderContext
	Context
	r ybmclient.ResourcePermissionsData
	opsIndex int
}

func NewRolePermissionFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaulRolePermissionListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// CloudRegionWrite renders the context for a list of cloud regions
func RolePermissionWrite(ctx Context, rolePermissions []ybmclient.ResourcePermissionsData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, rolePermission := range rolePermissions {
			for i := 0 ; i < len(rolePermission.Info.OperationGroups); i++ { 
				// logrus.Info(i)
				err := format(&RolePermissionContext{r: rolePermission, opsIndex: i})
				if err != nil {
					logrus.Debugf("Error rendering cloud region: %v", err)
					return err
				}
			}
		}
		return nil
	}
	return ctx.Write(NewRolePermissionContext(), render)
}

// NewCloudRegionContext creates a new context for rendering cloud regions
func NewRolePermissionContext() *RolePermissionContext {
	rolePermissionCtx := RolePermissionContext{}
	rolePermissionCtx.Header = SubHeaderContext{
		"ResourceName":  resourceNameHeader,
		"ResourceType":  resourceTypeHeader,
		"OperationDescription":  operationDescriptionHeader,
		"OperationType":  operationTypeHeader,
	}
	return &rolePermissionCtx
}

func (r *RolePermissionContext) ResourceName() string {
	return fmt.Sprintf("%s", r.r.Info.GetResourceName())
}

func (r *RolePermissionContext) ResourceType() string {
	return fmt.Sprintf("%s", r.r.Info.GetResourceType())
}

func (r *RolePermissionContext) OperationDescription() string {
	return fmt.Sprintf("%s", r.r.Info.OperationGroups[r.opsIndex].GetOperationGroupDescription())
}

func (r *RolePermissionContext) OperationType() string {
	return fmt.Sprintf("%s", r.r.Info.OperationGroups[r.opsIndex].GetOperationGroup())
}


func (r *RolePermissionContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}
