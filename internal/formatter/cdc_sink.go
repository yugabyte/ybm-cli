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
	defaultCdcSinkListing = "table {{.Name}}\t{{.Type}}\t{{.HostName}}\t{{.State}}"
	typeHeader            = "Type"
	hostNameHeader        = "Host Name"
)

type CdcSinkContext struct {
	HeaderContext
	Context
	c ybmclient.CdcSinkData
}

func NewCdcSinkFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultCdcSinkListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// CdcSinkWrite renders the context for a list of cdc sinks
func CdcSinkWrite(ctx Context, cdcSinks []ybmclient.CdcSinkData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, cdcSink := range cdcSinks {
			err := format(&CdcSinkContext{c: cdcSink})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewCdcSinkContext(), render)
}

// NewCdcSinkContext creates a new context for rendering cdc sink
func NewCdcSinkContext() *CdcSinkContext {
	cdcSinkCtx := CdcSinkContext{}
	cdcSinkCtx.Header = SubHeaderContext{
		"Type":     typeHeader,
		"HostName": hostNameHeader,
		"State":    stateHeader,
		"Name":     nameHeader,
	}
	return &cdcSinkCtx
}

func (c *CdcSinkContext) Name() string {
	if v, ok := c.c.Spec.GetNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcSinkContext) Type() string {
	if v, ok := c.c.Spec.GetSinkTypeOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcSinkContext) HostName() string {
	if v, ok := c.c.Spec.Kafka.GetHostnameOk(); ok {
		return string(*v)
	}
	return ""
}
func (c *CdcSinkContext) State() string {
	if v, ok := c.c.Info.GetStatusOk(); ok {
		return string(*v)
	}
	return ""
}
func (c *CdcSinkContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
