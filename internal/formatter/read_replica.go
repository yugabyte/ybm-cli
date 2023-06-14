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
	"strconv"

	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultReadReplicaListing = "table {{.Region}}\t{{.Endpoint}}\t{{.State}}\t{{.Nodes}}\t{{.NodesSpec}}"
	regionHeader              = "Region"
	endpointHeader            = "Endpoint"
)

type ReadReplicaContext struct {
	HeaderContext
	Context
	rrSpec     ybmclient.ReadReplicaSpec
	rrEndpoint ybmclient.Endpoint
}

func NewReadReplicaFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultReadReplicaListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// ReadReplicaWrite renders the context for a list of read replicas
func ReadReplicaWrite(ctx Context, rrSpecs []ybmclient.ReadReplicaSpec, rrEndpoints []ybmclient.Endpoint) error {
	render := func(format func(subContext SubContext) error) error {
		for index, rrSpec := range rrSpecs {
			err := format(&ReadReplicaContext{rrSpec: rrSpec, rrEndpoint: rrEndpoints[index]})
			if err != nil {
				logrus.Debugf("Error rendering read replica: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewReadReplicaContext(), render)
}

// NewReadReplicaContext creates a new context for rendering readReplica
func NewReadReplicaContext() *ReadReplicaContext {
	readReplicaCtx := ReadReplicaContext{}
	readReplicaCtx.Header = SubHeaderContext{
		"Region":    regionHeader,
		"Nodes":     numNodesHeader,
		"NodesSpec": nodeInfoHeader,
		"State":     stateHeader,
		"Endpoint":  endpointHeader,
	}
	return &readReplicaCtx
}

func (c *ReadReplicaContext) Region() string {
	if v, ok := c.rrEndpoint.GetRegionOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *ReadReplicaContext) State() string {
	if v, ok := c.rrEndpoint.GetStateOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *ReadReplicaContext) Endpoint() string {
	if v, ok := c.rrEndpoint.GetHostOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *ReadReplicaContext) NodesSpec() string {
	iops := "-"
	if c.rrSpec.NodeInfo.DiskIops.Get() != nil {
		iops = strconv.Itoa(int(*c.rrSpec.NodeInfo.DiskIops.Get()))
	}
	return fmt.Sprintf("%d / %s / %dGB / %s",
		c.rrSpec.NodeInfo.NumCores,
		convertMbtoGb(c.rrSpec.NodeInfo.MemoryMb),
		c.rrSpec.NodeInfo.DiskSizeGb,
		iops)
}

func (c *ReadReplicaContext) Nodes() string {
	return fmt.Sprintf("%d", c.rrSpec.PlacementInfo.NumNodes)
}

func (c *ReadReplicaContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Spec     ybmclient.ReadReplicaSpec
		Endpoint ybmclient.Endpoint
	}{
		Spec:     c.rrSpec,
		Endpoint: c.rrEndpoint,
	})
}
