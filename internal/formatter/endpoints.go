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

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultEndpointListing = "table {{.Region}}\t{{.Accessibility}}\t{{.State}}\t{{.Host}}"
	accessibilityHeader    = "Accessibility"
	hostHeader             = "Host"
)

type EndpointContext struct {
	HeaderContext
	Context
	e ybmclient.Endpoint
}

func NewEndpointFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultEndpointListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func EndpointWrite(ctx Context, endpoints []ybmclient.Endpoint) error {
	render := func(format func(subContext SubContext) error) error {
		for _, endpoint := range endpoints {
			err := format(&EndpointContext{e: endpoint})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewEndpointContext(), render)
}

func NewEndpointContext() *EndpointContext {
	epCtx := EndpointContext{}
	epCtx.Header = SubHeaderContext{
		"Region":        regionHeader,
		"Accessibility": accessibilityHeader,
		"State":         stateHeader,
		"Host":          hostHeader,
	}
	return &epCtx
}

func (e *EndpointContext) Region() string {
	return e.e.GetRegion()
}

func (e *EndpointContext) Accessibility() string {
	return string(e.e.GetAccessibilityType())
}

func (e *EndpointContext) State() string {
	return string(e.e.GetState())
}

func (e *EndpointContext) Host() string {
	return e.e.GetHost()
}

func (e *EndpointContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.e)
}
