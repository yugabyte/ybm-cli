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
	"fmt"
	"runtime"

	"github.com/enescakir/emoji"
	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultRoleListing	= "table {{.Name}}\t{{.Description}}\t{{.IsUserDefined}}\t{{.UsersCount}}\t{{.ApiKeysCount}}"
	isUserDefinedHeader = "User Defined"
	usersCountHeader 	= "Users Count"
	apiKeysCountHeader 	= "API Keys Count"
)

type RoleContext struct {
	HeaderContext
	Context
	r ybmclient.RoleData
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

func NewRoleFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultRoleListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// RoleWrite renders the context for a list of roles
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


func (r *RoleContext) ID() string {
	return r.r.Info.Id
}

func (r *RoleContext) Name() string {
	return r.r.Info.GetDisplayName()
}

func (r *RoleContext) Description() string {
	return r.r.GetDescription()
}

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

func (r *RoleContext) UsersCount() int {
	return len(r.r.Info.GetUsers())
}

func (r *RoleContext) ApiKeysCount() int {
	return len(r.r.Info.GetApiKeys())
}

func (r *RoleContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.r)
}