// Copyright (c) YugaByte, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package formatter

import (
	"encoding/json"
	"strings"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultCdcStreamListing = "table {{.Name}}\t{{.DBName}}\t{{.Tables}}\t{{.KafkaPrefix}}\t{{.State}}\t{{.LagTime}}"
	dbNameHeader            = "Database Name"
	tablesHeader            = "Tables"
	kafkaPrefixHeader       = "Kafka Prefix"
	lagTimeHeader           = "Lag Time(sec)"
)

type CdcStreamContext struct {
	HeaderContext
	Context
	c ybmclient.CdcStreamData
}

func NewCdcStreamFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultCdcStreamListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// CdStreamWrite renders the context for a list of cdc streams
func CdcStreamWrite(ctx Context, cdcStreams []ybmclient.CdcStreamData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, cdcStream := range cdcStreams {
			err := format(&CdcStreamContext{c: cdcStream})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewCdcStreamContext(), render)
}

// NewCdcStreamContext creates a new context for rendering cdc stream
func NewCdcStreamContext() *CdcStreamContext {
	cdcStreamCtx := CdcStreamContext{}
	cdcStreamCtx.Header = SubHeaderContext{
		"Tables":      tablesHeader,
		"DBName":      dbNameHeader,
		"State":       stateHeader,
		"Name":        nameHeader,
		"KafkaPrefix": kafkaPrefixHeader,
		"LagTime":     lagTimeHeader,
	}
	return &cdcStreamCtx
}

func (c *CdcStreamContext) Name() string {
	if v, ok := c.c.Spec.GetNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) DBName() string {
	if v, ok := c.c.Spec.GetDbNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) State() string {
	if v, ok := c.c.Info.GetStatusOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) Tables() string {
	if v, ok := c.c.Spec.GetTablesOk(); ok {
		return strings.Join(*v, ",")
	}
	return ""
}

func (c *CdcStreamContext) KafkaPrefix() string {
	if v, ok := c.c.Spec.GetKafkaPrefixOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) LagTime() string {
	if v, ok := c.c.Info.GetStatusOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
