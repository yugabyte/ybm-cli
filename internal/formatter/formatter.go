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
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/internal/formatter/tabwriter"
	"github.com/yugabyte/ybm-cli/internal/formatter/templates"
	"golang.org/x/exp/utf8string"
)

// Format keys used to specify certain kinds of output formats
const (
	TableFormatKey  = "table"
	RawFormatKey    = "raw"
	PrettyFormatKey = "pretty"
	JSONFormatKey   = "json"

	DefaultQuietFormat = "{{.ID}}"
	jsonFormat         = "{{json .}}"
	prettyFormat       = "{{. | toPrettyJson}}"

	// default header use accross multiple formatter
	clustersHeader        = "Clusters"
	descriptionHeader     = "Description"
	nameHeader            = "Name"
	regionsHeader         = "Regions"
	softwareVersionHeader = "Version"
	stateHeader           = "State"
	providerHeader        = "Provider"

	GREEN_COLOR = "green"
	RED_COLOR   = "red"
)

// Format is the format string rendered using the Context
type Format string

// IsTable returns true if the format is a table-type format
func (f Format) IsTable() bool {
	return strings.HasPrefix(string(f), TableFormatKey)
}

// IsJSON returns true if the format is the json format
func (f Format) IsJSON() bool {
	return string(f) == JSONFormatKey
}

// IsJSON returns true if the format is the json format
func (f Format) IsPrettyJson() bool {
	return string(f) == PrettyFormatKey
}

// Contains returns true if the format contains the substring
func (f Format) Contains(sub string) bool {
	return strings.Contains(string(f), sub)
}

// Context contains information required by the formatter to print the output as desired.
type Context struct {
	// Output is the output stream to which the formatted string is written.
	Output io.Writer
	// Format is used to choose raw, table or custom format for the output.
	Format Format

	// internal element
	finalFormat string
	header      interface{}
	buffer      *bytes.Buffer
}

func (c *Context) preFormat() {
	c.finalFormat = string(c.Format)
	// TODO: handle this in the Format type
	switch {
	case c.Format.IsTable():
		c.finalFormat = c.finalFormat[len(TableFormatKey):]
	case c.Format.IsJSON():
		c.finalFormat = jsonFormat
	case c.Format.IsPrettyJson():
		c.finalFormat = prettyFormat
	}

	c.finalFormat = strings.Trim(c.finalFormat, " ")
	r := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
	c.finalFormat = r.Replace(c.finalFormat)
}

func (c *Context) parseFormat() (*template.Template, error) {
	tmpl, err := templates.Parse(c.finalFormat)
	if err != nil {
		return tmpl, errors.Wrap(err, "template parsing error")
	}
	return tmpl, err
}

func (c *Context) postFormat(tmpl *template.Template, subContext SubContext) {
	if c.Format.IsTable() {
		t := tabwriter.NewWriter(c.Output, 10, 1, 3, ' ', 0)
		buffer := bytes.NewBufferString("")
		tmpl.Funcs(templates.HeaderFunctions).Execute(buffer, subContext.FullHeader())
		buffer.WriteTo(t)
		t.Write([]byte("\n"))
		c.buffer.WriteTo(t)
		t.Flush()
	} else {
		c.buffer.WriteTo(c.Output)
	}
}

func (c *Context) contextFormat(tmpl *template.Template, subContext SubContext) error {
	if err := tmpl.Execute(c.buffer, subContext); err != nil {
		return errors.Wrap(err, "template parsing error")
	}
	if c.Format.IsTable() && c.header != nil {
		c.header = subContext.FullHeader()
	}
	c.buffer.WriteString("\n")
	return nil
}

// SubFormat is a function type accepted by Write()
type SubFormat func(func(SubContext) error) error

// Write the template to the buffer using this Context
func (c *Context) Write(sub SubContext, f SubFormat) error {
	c.buffer = bytes.NewBufferString("")
	c.preFormat()

	tmpl, err := c.parseFormat()
	if err != nil {
		return err
	}

	subFormat := func(subContext SubContext) error {
		return c.contextFormat(tmpl, subContext)
	}
	if err := f(subFormat); err != nil {
		return err
	}

	c.postFormat(tmpl, sub)
	return nil
}

// Colorize the message accoring the colors var
func Colorize(message string, colors string) string {
	//If Colors is disable return the message as it is.
	if viper.GetBool("no-color") {
		color.NoColor = true
	}
	switch colors {
	case GREEN_COLOR:
		return color.GreenString(message)
	case RED_COLOR:
		return color.RedString(message)
	default:
		return message
	}
}

func Truncate(text string, lenght int) string {
	if lenght <= 0 || len(text) <= 0 {
		return ""
	}
	s := utf8string.NewString(text)
	return fmt.Sprintf("%s...", s.Slice(0, lenght))
}
