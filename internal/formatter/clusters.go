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
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/inhies/go-bytesize"
	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	"golang.org/x/exp/maps"
)

const (
	defaultClusterListing = "table {{.Name}}\t{{.Tier}}\t{{.SoftwareVersion}}\t{{.State}}\t{{.HealthState}}\t{{.Provider}}\t{{.Regions}}\t{{.Nodes}}\t{{.NodesSpec}}"
	numNodesHeader        = "Nodes"
	nodeInfoHeader        = "Node Res.(Vcpu/Mem/DiskGB/IOPS)"
	healthStateHeader     = "Health"
	tierHeader            = "Tier"
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

// ClusterWrite renders the context for a list of clusters
func ClusterWrite(ctx Context, clusters []ybmclient.ClusterData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, cluster := range clusters {
			err := format(&ClusterContext{c: cluster})
			if err != nil {
				logrus.Debugf("Error rendering cluster: %v", err)
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
		"Name":             nameHeader,
		"ID":               "ID",
		"Regions":          regionsHeader,
		"Nodes":            numNodesHeader,
		"NodesSpec":        nodeInfoHeader,
		"SoftwareVersion":  softwareVersionHeader,
		"State":            stateHeader,
		"HealthState":      healthStateHeader,
		"Provider":         providerHeader,
		"FaultTolerance":   faultToleranceHeader,
		"DataDistribution": dataDistributionHeader,
		"Tier":             tierHeader,
	}
	return &clusterCtx
}

func (c *ClusterContext) ID() string {
	return c.c.Info.Id
}

func (c *ClusterContext) Name() string {
	return c.c.Spec.Name
}

// Return single region or the first regions with +number of others region
func (c *ClusterContext) Regions() string {
	if ok := c.c.Spec.HasClusterRegionInfo(); ok {

		if len(c.c.GetSpec().ClusterRegionInfo) > 1 {
			sort.Slice(c.c.GetSpec().ClusterRegionInfo, func(i, j int) bool {
				return c.c.GetSpec().ClusterRegionInfo[i].PlacementInfo.CloudInfo.Region < c.c.GetSpec().ClusterRegionInfo[j].PlacementInfo.CloudInfo.Region
			})
			return fmt.Sprintf("%s,+%d", c.c.GetSpec().ClusterRegionInfo[0].PlacementInfo.CloudInfo.Region, len(c.c.GetSpec().ClusterRegionInfo)-1)
		} else {
			return c.c.GetSpec().ClusterRegionInfo[0].PlacementInfo.CloudInfo.Region
		}

	}
	return ""
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
	iops := "-"
	if c.c.GetSpec().ClusterInfo.NodeInfo.DiskIops.Get() != nil {
		iops = strconv.Itoa(int(*c.c.GetSpec().ClusterInfo.NodeInfo.DiskIops.Get()))
	}
	return fmt.Sprintf("%d / %s / %dGB / %s",
		c.c.GetSpec().ClusterInfo.NodeInfo.NumCores,
		convertMbtoGb(c.c.GetSpec().ClusterInfo.NodeInfo.MemoryMb),
		c.c.GetSpec().ClusterInfo.NodeInfo.DiskSizeGb,
		iops)
}

func (c *ClusterContext) Nodes() string {
	return fmt.Sprintf("%d", c.c.GetSpec().ClusterInfo.NumNodes)
}

func (c *ClusterContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}

func (c *ClusterContext) Provider() string {
	providers := make(map[string]string)

	if ok := c.c.Spec.HasClusterRegionInfo(); ok {
		if len(c.c.GetSpec().ClusterRegionInfo) > 0 {
			sort.Slice(c.c.GetSpec().ClusterRegionInfo, func(i, j int) bool {
				return string(c.c.GetSpec().ClusterRegionInfo[i].PlacementInfo.CloudInfo.Code) < string(c.c.GetSpec().ClusterRegionInfo[j].PlacementInfo.CloudInfo.Code)
			})
			for _, p := range c.c.GetSpec().ClusterRegionInfo {
				//Check uniqueness of Cloud (in case multi cloud with strange distribution, AWS, GCP,AWS)
				if _, ok := providers[string(p.PlacementInfo.CloudInfo.Code)]; !ok {
					providers[string(p.PlacementInfo.CloudInfo.Code)] = string(p.PlacementInfo.CloudInfo.Code)
				}
			}
		}
	}
	return strings.Join(maps.Keys(providers), ",")
}

func (c *ClusterContext) Tier() string {
	if c.c.GetSpec().ClusterInfo.ClusterTier == ybmclient.CLUSTERTIER_FREE {
		return "Sandbox"
	}
	return "Dedicated"
}
func (c *ClusterContext) FaultTolerance() string {
	return string(c.c.GetSpec().ClusterInfo.FaultTolerance)
}

func (c *ClusterContext) DataDistribution() string {
	return "No idea"
}

// clusterHealthStateToEmoji return emoji based on cluster health state
// See http://www.unicode.org/emoji/charts/emoji-list.html#1f49a
func clusterHealthStateToEmoji(healthState ybmclient.ClusterHealthState) string {

	// Windows terminal do not support emoji
	// So we return directly the healthstate
	if runtime.GOOS == "windows" {
		return string(healthState)
	}
	switch healthState {
	case ybmclient.CLUSTERHEALTHSTATE_HEALTHY:
		return emoji.GreenHeart.String()
	case ybmclient.CLUSTERHEALTHSTATE_NEEDS_ATTENTION:
		return emoji.Warning.String()
	case ybmclient.CLUSTERHEALTHSTATE_UNHEALTHY:
		return emoji.Collision.String()
	case ybmclient.CLUSTERHEALTHSTATE_UNKNOWN:
		return emoji.QuestionMark.String()
	default:
		return string(healthState)
	}
}

// convertMbtoGb convert MB to GB
func convertMbtoGb(sizeInMB int32) string {
	b, err := bytesize.Parse(fmt.Sprintf("%d MB", sizeInMB))

	if err != nil {
		logrus.Errorf("could not parse size: %v", err)
		return ""
	}
	return b.Format("%.0f", "GB", false)
}
