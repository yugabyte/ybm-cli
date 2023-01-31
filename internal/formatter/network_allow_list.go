package formatter

import (
	"encoding/json"
	"strings"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultNalListing      = "table {{.Name}}\t{{.Desc}}\t{{.AllowedList}}\t{{.Clusters}}"
	networkAllowListHeader = "Allow List"
)

type NetworkAllowListContext struct {
	HeaderContext
	Context
	c ybmclient.NetworkAllowListData
}

func NewNetworkAllowListFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultNalListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// NetworkAllowListWrite renders the context for a list of containers
func NetworkAllowListWrite(ctx Context, nals []ybmclient.NetworkAllowListData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, nal := range nals {
			err := format(&NetworkAllowListContext{c: nal})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewNetworkAllowListContext(), render)
}

// NewNetworkAllowListContext creates a new context for rendering nal
func NewNetworkAllowListContext() *NetworkAllowListContext {
	nalCtx := NetworkAllowListContext{}
	nalCtx.Header = SubHeaderContext{
		"AllowedList": networkAllowListHeader,
		"Clusters":    clustersHeader,
		"Desc":        descriptionHeader,
		"Name":        nameHeader,
	}
	return &nalCtx
}

func (c *NetworkAllowListContext) Name() string {
	return c.c.Spec.Name
}

func (c *NetworkAllowListContext) AllowedList() string {
	return strings.Join(c.c.GetSpec().AllowList, ",")
}

func (c *NetworkAllowListContext) Desc() string {
	return c.c.GetSpec().Description
}
func (c *NetworkAllowListContext) Clusters() string {
	var clusterNameList []string
	for _, cluster := range c.c.GetInfo().ClusterList {
		clusterNameList = append(clusterNameList, cluster.Name)
	}
	return strings.Join(clusterNameList, ",")
}
func (c *NetworkAllowListContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
