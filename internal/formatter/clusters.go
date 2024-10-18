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
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	defaultClusterListing   = "table {{.Name}}\t{{.Tier}}\t{{.SoftwareVersion}}\t{{.State}}\t{{.HealthState}}\t{{.Provider}}\t{{.Regions}}\t{{.Nodes}}\t{{.NodesSpec}}"
	defaultClusterListingV2 = "table {{.Name}}\t{{.Tier}}\t{{.SoftwareVersion}}\t{{.State}}\t{{.HealthState}}\t{{.Provider}}\t{{.Regions}}\t{{.Nodes}}\t{{.NodesSpec}}\t{{.ConnectionPoolingStatus}}"
	numNodesHeader          = "Nodes"
	nodeInfoHeader          = "Node Res.(Vcpu/Mem/DiskGB/IOPS)"
	healthStateHeader       = "Health"
	tierHeader              = "Tier"
	connectionPoolingHeader = "Connection Pooling Enabled"
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
		if util.IsFeatureFlagEnabled(util.CONNECTION_POOLING) {
			format = defaultClusterListingV2
		}
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

	clusterContext := NewClusterContext()

	if util.IsFeatureFlagEnabled(util.CONNECTION_POOLING) {
		clusterContext = NewClusterContextV2()
	}

	return ctx.Write(clusterContext, render)
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

func NewClusterContextV2() *ClusterContext {
	clusterCtx := ClusterContext{}
	clusterCtx.Header = SubHeaderContext{
		"Name":                    nameHeader,
		"ID":                      "ID",
		"Regions":                 regionsHeader,
		"Nodes":                   numNodesHeader,
		"NodesSpec":               nodeInfoHeader,
		"SoftwareVersion":         softwareVersionHeader,
		"State":                   stateHeader,
		"HealthState":             healthStateHeader,
		"Provider":                providerHeader,
		"FaultTolerance":          faultToleranceHeader,
		"DataDistribution":        dataDistributionHeader,
		"Tier":                    tierHeader,
		"ConnectionPoolingStatus": connectionPoolingHeader,
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

func (c *ClusterContext) ConnectionPoolingStatus() bool {
	if v, ok := c.c.Info.GetIsConnectionPoolingEnabledOk(); ok {
		return *v
	}
	return false
}

func (c *ClusterContext) HealthState() string {
	if v, ok := c.c.Info.GetHealthInfoOk(); ok {
		return clusterHealthStateToEmoji(v.GetState())
	}
	return ""
}

func (c *ClusterContext) NodesSpec() string {
	iops := "-"
	if c.c.GetSpec().ClusterInfo.NodeInfo.Get().DiskIops.Get() != nil {
		iops = strconv.Itoa(int(*c.c.GetSpec().ClusterInfo.NodeInfo.Get().DiskIops.Get()))
	}
	return fmt.Sprintf("%d / %s / %dGB / %s",
		c.c.GetSpec().ClusterInfo.NodeInfo.Get().NumCores,
		convertMbtoGb(c.c.GetSpec().ClusterInfo.NodeInfo.Get().MemoryMb),
		c.c.GetSpec().ClusterInfo.NodeInfo.Get().DiskSizeGb,
		iops)
}

func (c *ClusterContext) Nodes() string {
	return fmt.Sprintf("%d", c.c.GetSpec().ClusterInfo.NumNodes)
}

func (c *ClusterContext) MarshalJSON() ([]byte, error) {
	//Removing Azure Private endpoit from Json output
	if len(c.c.Info.GetClusterEndpoints()) > 0 {
		c.c.Info.ClusterEndpoints = slices.DeleteFunc(c.c.Info.ClusterEndpoints, func(ep ybmclient.Endpoint) bool {
			if ep.AccessibilityType == ybmclient.ACCESSIBILITYTYPE_PRIVATE && c.c.Spec.CloudInfo.Code == ybmclient.CLOUDENUM_AZURE {
				return true
			}
			return false
		})

	}
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
	rf := *c.c.GetSpec().ClusterInfo.NumFaultsToTolerate.Get()*2 + 1
	return fmt.Sprintf("%s, RF %d", string(c.c.GetSpec().ClusterInfo.FaultTolerance), rf)
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
