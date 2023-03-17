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
	"fmt"
	"runtime"
	"sort"

	"github.com/enescakir/emoji"
	"github.com/inhies/go-bytesize"
	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultNodeListing = "table {{.Name}}\t{{.RegionZone}}\t{{.IsNodeUp}}\t{{.IsMaster}}\t{{.IsTserver}}\t{{.IsRR}}\t{{.MemoryUsed}}"
	regionZoneHeader   = "Region[zone]"
	isMasterHeader     = "Master"
	isTserverHeader    = "Tserver"
	isRRHeader         = "ReadReplica"
	memoryUsedHeader   = "Used Memory(MB)"
)

type NodeContext struct {
	HeaderContext
	Context
	n ybmclient.NodeData
}

func NewNodeFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultNodeListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NodeWrite(ctx Context, node []ybmclient.NodeData) error {
	//Sort by name
	sort.Slice(node, func(i, j int) bool {
		return string(node[i].Name) < string(node[j].Name)
	})

	render := func(format func(subContext SubContext) error) error {
		for _, node := range node {
			err := format(&NodeContext{n: node})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewNodeContext(), render)
}

func NewNodeContext() *NodeContext {
	npCtx := NodeContext{}
	npCtx.Header = SubHeaderContext{
		"Name":       nameHeader,
		"RegionZone": regionZoneHeader,
		"IsTserver":  isTserverHeader,
		"IsMaster":   isMasterHeader,
		"IsNodeUp":   healthStateHeader,
		"IsRR":       isRRHeader,
		"MemoryUsed": memoryUsedHeader,
	}
	return &npCtx
}

func (n *NodeContext) RegionZone() string {
	if v, ok := n.n.GetCloudInfoOk(); ok {
		return fmt.Sprintf("%s[%s]", v.GetRegion(), v.GetZone())
	}
	return ""
}

func (n *NodeContext) Name() string {
	return n.n.GetName()
}

func (n *NodeContext) IsMaster() string {
	return NodeTypeToEmoji(n.n.IsMaster)
}

func (n *NodeContext) IsTserver() string {
	return NodeTypeToEmoji(n.n.IsTserver)
}

func (n *NodeContext) IsNodeUp() string {
	return NodeHealthToEmoji(n.n.IsNodeUp)
}

func (n *NodeContext) IsRR() string {
	return NodeTypeToEmoji(*n.n.IsReadReplica)
}

func (n *NodeContext) MemoryUsed() string {
	if m, ok := n.n.GetMetricsOk(); ok {
		return ConvertBytestoGb(m.GetMemoryUsedBytes())
	}
	return ""
}

// func (n *NodeContext) Host() string {
// 	return e.e.GetHost()
// }

// func (n *NodeContext) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(e.e)
// }

func NodeTypeToEmoji(nodeType bool) string {

	// Windows terminal do not support emoji
	// So we return directly the healthstate
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%t", nodeType)
	}
	switch nodeType {
	case true:
		return emoji.Parse(":white_check_mark:")
	case false:
		return emoji.CrossMark.String()
	default:
		return fmt.Sprintf("%t", nodeType)
	}
}

func NodeHealthToEmoji(status bool) string {

	// Windows terminal do not support emoji
	// So we return directly the healthstate
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%t", status)
	}
	switch status {
	case true:
		return emoji.GreenHeart.String()
	case false:
		return emoji.CrossMark.String()
	default:
		return fmt.Sprintf("%t", status)
	}
}

// convertMbtoGb convert MB to GB
func ConvertBytestoGb(sizeInB int64) string {
	b, err := bytesize.Parse(fmt.Sprintf("%d B", sizeInB))

	if err != nil {
		logrus.Errorf("could not parse size: %v", err)
		return ""
	}
	return b.Format("%.0f", "MB", false)
}
