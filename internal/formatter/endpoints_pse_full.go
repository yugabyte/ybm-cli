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
	"strings"
	"text/template"

	"github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	psEndpointAccess         = "table {{.Az}}\t{{.SecurityPrincipals}}"
	psEndpointConn           = "table {{.ServiceName}}\t{{.State}}\t{{.ActiveConnections}}"
	azHeader                 = "Availability Zones"
	securityPrincipalsHeader = "Security Principals"
	serviceNameHeader        = "Service Name"
	activeConnectionsHeader  = "Active Connections"
)

type PSEndpointFullContext struct {
	HeaderContext
	Context
	psEndpoint ybmclient.PrivateServiceEndpointRegionData
	endpoint   ybmclient.Endpoint
}

func NewPSEndpointContext() *PSEndpointFullContext {
	psEndpointCtx := PSEndpointFullContext{}
	psEndpointCtx.Header = SubHeaderContext{
		"Az":                 azHeader,
		"SecurityPrincipals": securityPrincipalsHeader,
		"ServiceName":        serviceNameHeader,
		"State":              stateHeader,
		"ActiveConnections":  activeConnectionsHeader,
	}
	return &psEndpointCtx
}

func NewPSEndpointFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultEndpointListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func PSEndpointWrite(ctx Context, pseData ybmclient.PrivateServiceEndpointRegionData, endpoint ybmclient.Endpoint) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&PSEndpointFullContext{
			psEndpoint: pseData,
			endpoint:   endpoint,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return ctx.Write(NewPSEndpointContext(), render)
}

func (ep *PSEndpointFullContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Endpoint   ybmclient.Endpoint
		PsEndpoint ybmclient.PrivateServiceEndpointRegionData
	}{
		Endpoint:   ep.endpoint,
		PsEndpoint: ep.psEndpoint,
	})
}

func (ep *PSEndpointFullContext) SetFullPSEndpoint(authApi client.AuthApiClient, pseData ybmclient.PrivateServiceEndpointRegionData, endpoint ybmclient.Endpoint) {
	ep.psEndpoint = pseData
	ep.endpoint = endpoint
}

func (ep *PSEndpointFullContext) startSubsection(format string) (*template.Template, error) {
	ep.buffer = bytes.NewBufferString("")
	ep.header = ""
	ep.Format = Format(format)
	ep.preFormat()

	return ep.parseFormat()
}

func (ep *PSEndpointFullContext) Write() error {
	epContext := &EndpointContext{}
	epContext.e = ep.endpoint

	// Endpoint listing
	tmpl, err := ep.startSubsection(defaultEndpointListing)
	if err != nil {
		return err
	}
	ep.Output.Write([]byte(Colorize("General", GREEN_COLOR)))
	ep.Output.Write([]byte("\n"))
	if err := ep.contextFormat(tmpl, epContext); err != nil {
		return err
	}
	ep.postFormat(tmpl, NewEndpointContext())

	// PSEndpoint accessibility
	tmpl, err = ep.startSubsection(psEndpointAccess)
	if err != nil {
		return err
	}
	ep.Output.Write([]byte(Colorize("Access", GREEN_COLOR)))
	ep.Output.Write([]byte("\n"))
	if err := ep.contextFormat(tmpl, ep); err != nil {
		return err
	}
	ep.postFormat(tmpl, ep)

	// PSEndpoint connectivity
	tmpl, err = ep.startSubsection(psEndpointConn)
	if err != nil {
		return err
	}
	ep.Output.Write([]byte(Colorize("Connectivity", GREEN_COLOR)))
	ep.Output.Write([]byte("\n"))
	if err := ep.contextFormat(tmpl, ep); err != nil {
		return err
	}
	ep.postFormat(tmpl, ep)

	return nil
}

func (ep *PSEndpointFullContext) Az() string {
	return strings.Join(ep.psEndpoint.Info.AvailabilityZones, ",")
}

func (ep *PSEndpointFullContext) SecurityPrincipals() string {
	return strings.Join(ep.psEndpoint.Spec.SecurityPrincipals, ",")
}

func (ep *PSEndpointFullContext) ServiceName() string {
	return ep.psEndpoint.Info.GetServiceName()
}

func (ep *PSEndpointFullContext) State() string {
	return string(ep.psEndpoint.Info.State)
}

func (ep *PSEndpointFullContext) ActiveConnections() string {
	return strings.Join(ep.psEndpoint.Info.GetConnections(), ",")
}
