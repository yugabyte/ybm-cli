package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/enescakir/emoji"
	"github.com/inhies/go-bytesize"
	"github.com/sirupsen/logrus"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultClusterListing = "table {{.Name}}\t{{.SoftwareVersion}}\t{{.State}}\t{{.HealthState}}\t{{.Regions}}\t{{.Nodes}}\t{{.NodesSpec}}"
	numNodesHeader        = "Nodes"
	nodeInfoHeader        = "Total Res.(Vcpu/Mem/Disk)"
	healthStateHeader     = "Health"
)

type ClusterContext struct {
	HeaderContext
	Context
	c openapi.ClusterData
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
func ClusterWrite(ctx Context, clusters []openapi.ClusterData) error {
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
		"Name":            nameHeader,
		"Regions":         regionsHeader,
		"Nodes":           numNodesHeader,
		"NodesSpec":       nodeInfoHeader,
		"SoftwareVersion": softwareVersionHeader,
		"State":           stateHeader,
		"HealthState":     healthStateHeader,
	}
	return &clusterCtx
}

func (c *ClusterContext) Name() string {
	return c.c.Spec.Name
}

func (c *ClusterContext) Regions() string {
	return c.c.GetSpec().CloudInfo.Region
}

func (c *ClusterContext) State() string {
	if v, ok := c.c.Info.GetStateOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *ClusterContext) SoftwareVersion() string {
	if v, ok := c.c.Info.GetSoftwareVersionOk(); ok {
		return *v
	}
	return ""
}

func (c *ClusterContext) HealthState() string {
	if v, ok := c.c.Info.GetHealthInfoOk(); ok {
		return clusterHealthStateToEmoji(v.GetState())
	}
	return ""
}

func (c *ClusterContext) NodesSpec() string {
	return fmt.Sprintf("%d / %s / %dGB",
		c.totalResource(c.c.GetSpec().ClusterInfo.NodeInfo.NumCores),
		convertMbtoGb(c.totalResource(c.c.GetSpec().ClusterInfo.NodeInfo.MemoryMb)),
		c.totalResource(c.c.GetSpec().ClusterInfo.NodeInfo.DiskSizeGb))
}

func (c *ClusterContext) Nodes() string {
	return fmt.Sprintf("%d", c.c.GetSpec().ClusterInfo.NumNodes)
}

func (c *ClusterContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}

func (c *ClusterContext) totalResource(resource int32) int32 {
	return c.c.GetSpec().ClusterInfo.NumNodes * resource
}

// clusterHealthStateToEmoji return emoji based on cluster health state
// See http://www.unicode.org/emoji/charts/emoji-list.html#1f49a
func clusterHealthStateToEmoji(state openapi.ClusterHealthState) string {
	switch state {
	case openapi.CLUSTERHEALTHSTATE_HEALTHY:
		return emoji.GreenHeart.String()
	case openapi.CLUSTERHEALTHSTATE_NEEDS_ATTENTION:
		return emoji.Warning.String()
	case openapi.CLUSTERHEALTHSTATE_UNHEALTHY:
		return emoji.Collision.String()
	case openapi.CLUSTERHEALTHSTATE_UNKNOWN:
		return emoji.QuestionMark.String()
	default:
		return ""
	}
}

// convertMbtoGb convert MB to GB
func convertMbtoGb(sizeInMB int32) string {
	b, err := bytesize.Parse(fmt.Sprintf("%d MB", sizeInMB))

	if err != nil {
		logrus.Errorf("could not parse size: ", err)
		return ""
	}
	return b.Format("%.0f", "GB", false)
}
