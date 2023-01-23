package formatter

import (
	"encoding/json"
	"fmt"
	"strings"

	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultVPCListing = "table {{.Name}}\t{{.State}}\t{{.Provider}}\t{{.RegionsCIDR}}\t{{.Peerings}}\t{{.Clusters}}"
	vpcCIDRHeader     = "Region-CIDR"
	vpcPeeringHeader  = "Peerings"
)

type VPCContext struct {
	HeaderContext
	Context
	c openapi.SingleTenantVpcDataResponse
}

func NewVPCFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultVPCListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// VPCWrite renders the context for a list of containers
func VPCWrite(ctx Context, VPCs []openapi.SingleTenantVpcDataResponse) error {
	render := func(format func(subContext SubContext) error) error {
		for _, VPC := range VPCs {
			err := format(&VPCContext{c: VPC})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewVPCContext(), render)
}

// NewVPCContext creates a new context for rendering VPC
func NewVPCContext() *VPCContext {
	VPCCtx := VPCContext{}
	VPCCtx.Header = SubHeaderContext{
		"Name":        nameHeader,
		"State":       stateHeader,
		"RegionsCIDR": regionsHeader,
		"Provider":    providerHeader,
		"Peerings":    vpcPeeringHeader,
		"Clusters":    clustersHeader,
	}
	return &VPCCtx
}

func (c *VPCContext) Name() string {
	return c.c.Spec.Name
}

func (c *VPCContext) RegionsCIDR() string {
	var RegionsCIDRList []string
	for _, regionSpec := range c.c.GetSpec().RegionSpecs {
		RegionsCIDRList = append(RegionsCIDRList, fmt.Sprintf("%s[%s]", regionSpec.GetRegion(), regionSpec.GetCidr()))
	}
	return strings.Join(RegionsCIDRList, ",")
}

func (c *VPCContext) State() string {
	if v, ok := c.c.Info.GetStateOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCContext) Provider() string {
	if v, ok := c.c.Spec.GetCloudOk(); ok {
		return string(*v.Ptr())
	}
	return ""
}

func (c *VPCContext) Peerings() string {
	if v, ok := c.c.Info.GetPeeringIdsOk(); ok {
		return fmt.Sprint(len(*v))
	}
	return ""
}

func (c *VPCContext) Clusters() string {
	if v, ok := c.c.Info.GetClusterIdsOk(); ok {
		return fmt.Sprint(len(*v))
	}
	return ""
}

func (c *VPCContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}