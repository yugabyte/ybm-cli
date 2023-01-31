package formatter

import (
	"encoding/json"
	"strings"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultCdcStreamListing = "table {{.Name}}\t{{.DBName}}\t{{.Tables}}\t{{.KafkaPrefix}}\t{{.State}}\t{{.LagTime}}"
	dbNameHeader            = "Database Name"
	tablesHeader            = "Tables"
	kafkaPrefixHeader       = "Kafka Prefix"
	lagTimeHeader           = "Lag Time(sec)"
)

type CdcStreamContext struct {
	HeaderContext
	Context
	c ybmclient.CdcStreamData
}

func NewCdcStreamFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultCdcStreamListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// CdStreamWrite renders the context for a list of cdc streams
func CdcStreamWrite(ctx Context, cdcStreams []ybmclient.CdcStreamData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, cdcStream := range cdcStreams {
			err := format(&CdcStreamContext{c: cdcStream})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewCdcStreamContext(), render)
}

// NewCdcStreamContext creates a new context for rendering cdc stream
func NewCdcStreamContext() *CdcStreamContext {
	cdcStreamCtx := CdcStreamContext{}
	cdcStreamCtx.Header = SubHeaderContext{
		"Tables":      tablesHeader,
		"DBName":      dbNameHeader,
		"State":       stateHeader,
		"Name":        nameHeader,
		"KafkaPrefix": kafkaPrefixHeader,
		"LagTime":     lagTimeHeader,
	}
	return &cdcStreamCtx
}

func (c *CdcStreamContext) Name() string {
	if v, ok := c.c.Spec.GetNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) DBName() string {
	if v, ok := c.c.Spec.GetDbNameOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) State() string {
	if v, ok := c.c.Info.GetStatusOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) Tables() string {
	if v, ok := c.c.Spec.GetTablesOk(); ok {
		return strings.Join(*v, ",")
	}
	return ""
}

func (c *CdcStreamContext) KafkaPrefix() string {
	if v, ok := c.c.Spec.GetKafkaPrefixOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) LagTime() string {
	if v, ok := c.c.Info.GetStatusOk(); ok {
		return string(*v)
	}
	return ""
}

func (c *CdcStreamContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}
