package formatter

import (
	"encoding/json"
	"fmt"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultClusterListing = "table {{.Name}}\t{{.Nodes}}\t{{.NodesSpec}}"
	nameHeader            = "Name"
	numNodesHeader        = "Nodes"
	nodeInfoHeader        = "Node_Spec"
)

type ClusterContext struct {
	HeaderContext
	Context
	c ybmclient.ClusterData
}

func NewClusterFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultClusterListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// ClusterWrite renders the context for a list of containers
func ClusterWrite(ctx Context, clusters []ybmclient.ClusterData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, cluster := range clusters {
			err := format(&ClusterContext{c: cluster})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewClusterContext(), render)
}

// NewClusterContext creates a new context for rendering cluster
func NewClusterContext() *ClusterContext {
	clusterCtx := ClusterContext{}
	clusterCtx.Header = SubHeaderContext{
		"Name":      nameHeader,
		"Nodes":     numNodesHeader,
		"NodesSpec": nodeInfoHeader,
	}
	return &clusterCtx
}

func (c *ClusterContext) Name() string {
	return c.c.Spec.Name
}

func (c *ClusterContext) NodesSpec() string {
	return fmt.Sprintf("%d Vcpu,%d Mb Mem, %d Gb Disk",
		c.c.GetSpec().ClusterInfo.NodeInfo.NumCores,
		c.c.GetSpec().ClusterInfo.NodeInfo.MemoryMb,
		c.c.GetSpec().ClusterInfo.NodeInfo.DiskSizeGb)
}

func (c *ClusterContext) Nodes() string {
	return fmt.Sprintf("%d", c.c.GetSpec().ClusterInfo.NumNodes)
}

func (c *ClusterContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
