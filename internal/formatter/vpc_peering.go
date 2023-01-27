package formatter

import (
	"encoding/json"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultVPCPeeringListing = "table {{.Name}}\t{{.Provider}}\t{{.AppVPC}}\t{{.YbVPC}}\t{{.Status}}"
	appVPCHeader             = "Application VPC ID/Name"
	ybVPCHeader              = "YugabyteDB VPC Name"
	statusHeader             = "Status"
)

type VPCPeeringContext struct {
	HeaderContext
	Context
	c ybmclient.VpcPeeringData
}

func NewVPCPeeringFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultVPCPeeringListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// VPCPeeringWrite renders the context for a list of VPC Peerings
func VPCPeeringWrite(ctx Context, VPCPeerings []ybmclient.VpcPeeringData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, VPCPeering := range VPCPeerings {
			err := format(&VPCPeeringContext{c: VPCPeering})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewVPCPeeringContext(), render)
}

// NewVPCPeeringContext creates a new context for rendering VPC Peering
func NewVPCPeeringContext() *VPCPeeringContext {
	VPCPeeringCtx := VPCPeeringContext{}
	VPCPeeringCtx.Header = SubHeaderContext{
		"Name":     nameHeader,
		"Provider": providerHeader,
		"AppVPC":   appVPCHeader,
		"YbVPC":    ybVPCHeader,
		"Status":   statusHeader,
	}
	return &VPCPeeringCtx
}

func (c *VPCPeeringContext) Name() string {
	if v, ok := c.c.Spec.GetNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) Provider() string {
	if v, ok := c.c.Spec.CustomerVpc.CloudInfo.GetCodeOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) AppVPC() string {
	if v, ok := c.c.Spec.CustomerVpc.GetExternalVpcIdOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) YbVPC() string {
	if v, ok := c.c.Info.GetYugabyteVpcNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) Status() string {
	if v, ok := c.c.Info.GetStateOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *VPCPeeringContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
