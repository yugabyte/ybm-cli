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
	defaultResourcePermissionListing = "table {{.ResourceName}}\t{{.OperationDescription}}\t{{.ResourceType}}\t{{.OperationType}}"
	resourceNameHeader               = "Resource Name"
	resourceTypeHeader               = "Resource Type"
	operationDescriptionHeader       = "Operation Description"
	operationTypeHeader              = "Operation Group"
)

type ResourcePermissionContext struct {
	HeaderContext
	Context
	r        ybmclient.ResourcePermissionsData
	opsIndex int
}

func NewResourcePermissionFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultResourcePermissionListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// ResourcePermissionWrite renders the context for a list of resource permissions
func ResourcePermissionWrite(ctx Context, resourcePermissions []ybmclient.ResourcePermissionsData) error {
	sort.Slice(resourcePermissions, func(i, j int) bool {
		return string(resourcePermissions[i].Info.ResourceType) < string(resourcePermissions[j].Info.ResourceType)
	})
	render := func(format func(subContext SubContext) error) error {
		for _, resourcePermission := range resourcePermissions {
			for i := 0; i < len(resourcePermission.Info.OperationGroups); i++ {
				err := format(&ResourcePermissionContext{r: resourcePermission, opsIndex: i})
				if err != nil {
					logrus.Debugf("Error rendering available resource permissions: %v", err)
					return err
				}
			}
		}
		return nil
	}
	return ctx.Write(NewResourcePermissionContext(), render)
}

// NewResourcePermissionContext creates a new context for rendering resource permissions
func NewResourcePermissionContext() *ResourcePermissionContext {
	resourcePermissionCtx := ResourcePermissionContext{}
	resourcePermissionCtx.Header = SubHeaderContext{
		"ResourceName":         resourceNameHeader,
		"ResourceType":         resourceTypeHeader,
		"OperationDescription": operationDescriptionHeader,
		"OperationType":        operationTypeHeader,
	}
	return &resourcePermissionCtx
}

func (r *ResourcePermissionContext) ResourceName() string {
	return r.r.Info.GetResourceName()
}

func (r *ResourcePermissionContext) ResourceType() string {
	return string(r.r.Info.GetResourceType())
}

func (r *ResourcePermissionContext) OperationDescription() string {
	return r.r.Info.OperationGroups[r.opsIndex].GetOperationGroupDescription()
}

func (r *ResourcePermissionContext) OperationType() string {
	return string(r.r.Info.OperationGroups[r.opsIndex].GetOperationGroup())
}

func (r *ResourcePermissionContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}
