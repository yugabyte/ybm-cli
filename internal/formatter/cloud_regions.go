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
	"sort"

	emoji "github.com/jayco/go-emoji-flag"
	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaulCloudRegionListing = "table {{.RegionName}}\t{{.RegionCode}}\t{{.CountryCode}}"
	regionNameHeader         = "Region Name"
	regionCodeHeader         = "Region Code"
	countryCodeHeader        = "Country"
)

type CloudRegionContext struct {
	HeaderContext
	Context
	c ybmclient.RegionListResponseDataItem
}

func NewCloudRegionFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaulCloudRegionListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// CloudRegionWrite renders the context for a list of cloud regions
func CloudRegionWrite(ctx Context, cloudRegions []ybmclient.RegionListResponseDataItem) error {
	sort.Slice(cloudRegions, func(i, j int) bool {
		return string(cloudRegions[i].Name) < string(cloudRegions[j].Name)
	})
	render := func(format func(subContext SubContext) error) error {
		for _, cloudRegion := range cloudRegions {
			err := format(&CloudRegionContext{c: cloudRegion})
			if err != nil {
				logrus.Debugf("Error rendering cloud region: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewCloudRegionContext(), render)
}

// NewCloudRegionContext creates a new context for rendering cloud regions
func NewCloudRegionContext() *CloudRegionContext {
	cloudRegionCtx := CloudRegionContext{}
	cloudRegionCtx.Header = SubHeaderContext{
		"RegionName":  regionNameHeader,
		"RegionCode":  regionCodeHeader,
		"CountryCode": countryCodeHeader,
	}
	return &cloudRegionCtx
}

func (c *CloudRegionContext) RegionName() string {
	return c.c.GetName()
}

func (c *CloudRegionContext) RegionCode() string {
	return c.c.GetCode()
}

func (c *CloudRegionContext) CountryCode() string {
	return emoji.GetFlag(c.c.GetCountryCode()) + "  " + c.c.GetCountryCode()
}

func (c *CloudRegionContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
