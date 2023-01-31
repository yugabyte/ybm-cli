package formatter

import (
	"encoding/json"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultCdcSinkListing = "table {{.Name}}\t{{.Type}}\t{{.HostName}}\t{{.State}}"
	typeHeader            = "Type"
	hostNameHeader        = "Host Name"
)

type CdcSinkContext struct {
	HeaderContext
	Context
	c ybmclient.CdcSinkData
}

func NewCdcSinkFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultCdcSinkListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// CdcSinWrite renders the context for a list of containers
func CdcSinkWrite(ctx Context, cdcSinks []ybmclient.CdcSinkData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, cdcSink := range cdcSinks {
			err := format(&CdcSinkContext{c: cdcSink})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewCdcSinkContext(), render)
}

// NewCdcSinkContext creates a new context for rendering cdc sink
func NewCdcSinkContext() *CdcSinkContext {
	cdcSinkCtx := CdcSinkContext{}
	cdcSinkCtx.Header = SubHeaderContext{
		"Type":     typeHeader,
		"HostName": hostNameHeader,
		"State":    stateHeader,
		"Name":     nameHeader,
	}
	return &cdcSinkCtx
}

func (c *CdcSinkContext) Name() string {
	if v, ok := c.c.Spec.GetNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcSinkContext) Type() string {
	if v, ok := c.c.Spec.GetSinkTypeOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcSinkContext) HostName() string {
	if v, ok := c.c.Spec.Kafka.GetHostnameOk(); ok {
		return string(*v)
	}
	return ""
}
func (c *CdcSinkContext) State() string {
	if v, ok := c.c.Info.GetStatusOk(); ok {
		return string(*v)
	}
	return ""
}
func (c *CdcSinkContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
