package formatter

import (
	"fmt"

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
	return fmt.Sprintf("%d / %s / %dGB",
		c.totalResource(c.rrSpec.NodeInfo.NumCores),
		convertMbtoGb(c.totalResource(c.rrSpec.NodeInfo.MemoryMb)),
		c.totalResource(c.rrSpec.NodeInfo.DiskSizeGb))
}

func (c *ReadReplicaContext) Nodes() string {
	return fmt.Sprintf("%d", c.rrSpec.PlacementInfo.NumNodes)
}

func (c *ReadReplicaContext) totalResource(resource int32) int32 {
	return c.rrSpec.PlacementInfo.NumNodes * resource
}
