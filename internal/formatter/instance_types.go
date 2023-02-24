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
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022- Yugabyte, Inc.

package formatter

import (
	"encoding/json"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaulInstanceTypeListing = "table {{.Cores}}\t{{.Memory}}\t{{.DiskSize}}\t{{.AZs}}\t{{.IsEnabled}}"
	coresHeader               = "Number of Cores"
	memoryHeader              = "Memory (MB)"
	diskSizeHeader            = "Disk Size (GB)"
	azsHeader                 = "Number of Availability Zones"
	isEnabledHeader           = "Is Enabled"
)

type InstanceTypeContext struct {
	HeaderContext
	Context
	c ybmclient.NodeConfigurationResponseItem
}

func NewInstanceTypeFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaulInstanceTypeListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// InstanceTypeWrite renders the context for a list of instance types
func InstanceTypeWrite(ctx Context, instanceTypes []ybmclient.NodeConfigurationResponseItem) error {
	render := func(format func(subContext SubContext) error) error {
		for _, instanceType := range instanceTypes {
			err := format(&InstanceTypeContext{c: instanceType})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewInstanceTypeContext(), render)
}

// NewInstanceTypeContext creates a new context for rendering cloud regions
func NewInstanceTypeContext() *InstanceTypeContext {
	instanceTypeCtx := InstanceTypeContext{}
	instanceTypeCtx.Header = SubHeaderContext{
		"Cores":     coresHeader,
		"Memory":    memoryHeader,
		"DiskSize":  diskSizeHeader,
		"AZs":       azsHeader,
		"IsEnabled": isEnabledHeader,
	}
	return &instanceTypeCtx
}

func (c *InstanceTypeContext) Cores() int32 {
	return c.c.GetNumCores()
}

func (c *InstanceTypeContext) Memory() int32 {
	return c.c.GetMemoryMb()
}

func (c *InstanceTypeContext) DiskSize() int32 {
	return c.c.GetIncludedDiskSizeGb()
}

func (c *InstanceTypeContext) AZs() int32 {
	return c.c.GetNumAzs()
}

func (c *InstanceTypeContext) IsEnabled() bool {
	return c.c.GetIsEnabled()
}

func (c *InstanceTypeContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
