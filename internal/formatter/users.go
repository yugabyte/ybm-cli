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

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultUserListing = "table {{.UserEmail}}\t{{.UserName}}\t{{.UserRole}}\t{{.UserState}}"
	userRoleHeader     = "User Role"
)

type UserContext struct {
	HeaderContext
	Context
	u ybmclient.UserData
}

func NewUserFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultUserListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// UserWrite renders the context for a list of users
func UserWrite(ctx Context, users []ybmclient.UserData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, user := range users {
			err := format(&UserContext{u: user})
			if err != nil {
				logrus.Debugf("Error rendering user: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewUserContext(), render)
}

// NewUserContext creates a new context for rendering users
func NewUserContext() *UserContext {
	userCtx := UserContext{}
	userCtx.Header = SubHeaderContext{
		"UserEmail": userEmailHeader,
		"UserName":  userNameHeader,
		"UserState": userStateHeader,
		"UserRole":  userRoleHeader,
	}
	return &userCtx
}

func (u *UserContext) UserEmail() string {
	return u.u.Spec.Email
}

func (u *UserContext) UserName() string {
	return u.u.Spec.GetFirstName() + " " + u.u.Spec.GetLastName()
}

func (u *UserContext) UserState() string {
	return fmt.Sprintf("%s", u.u.Info.State)
}

func (u *UserContext) UserRole() string {
	return u.u.Info.GetRoleList()[0].GetRoles()[0].Info.GetDisplayName()
}

func (u *UserContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.u)
}
