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
