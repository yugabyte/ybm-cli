package formatter

import (
	"encoding/json"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaulCloudRegionListing = "table {{.RegionName}}\t{{.RegionCode}}\t{{.CountryCode}}"
	regionNameHeader         = "Region Name"
	regionCodeHeader         = "Region Code"
	countryCodeHeader        = "Country Code"
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
	render := func(format func(subContext SubContext) error) error {
		for _, cloudRegion := range cloudRegions {
			err := format(&CloudRegionContext{c: cloudRegion})
			if err != nil {
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
	return c.c.GetCountryCode()
}

func (c *CloudRegionContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
