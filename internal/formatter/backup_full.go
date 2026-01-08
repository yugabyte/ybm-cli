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
	"encoding/json"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/yugabyte/ybm-cli/internal/backup"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	backupGeneral1  = "table {{.Id}}\t{{.CreatedOn}}\t{{.Incremental}}\t{{.ClusterName}}\t{{.BackupState}}"
	backupGeneral2  = "table {{.BackupType}}\t{{.Size}}\t{{.ExpireOn}}\t{{.Duration}}\t{{.IncludeRoles}}"
	databaseListing = "table {{.Db}}\t{{.ApiType}}"
	dbHeader        = "Database/Keyspace"
	apiTypeHeader   = "API Type"
	durationHeader  = "Duration"
)

type FullBackupContext struct {
	HeaderContext
	Context
	fullBackup *backup.FullBackup
}

// NewFullBackupContext creates a new context for rendering all backup details
func NewFullBackupContext() *FullBackupContext {
	backupCtx := FullBackupContext{}
	backupCtx.Header = SubHeaderContext{
		"Id":           backupIdHeader,
		"BackupState":  stateHeader,
		"BackupType":   backupTypeHeader,
		"ClusterName":  clustersHeader,
		"CreatedOn":    backupIdCreateOnHeader,
		"Description":  descriptionHeader,
		"ExpireOn":     backupIdExpireOnHeader,
		"RetainInDays": retainInDaysHeader,
		"Incremental":  incrementalHeader,
		"Size":         sizeHeader,
		"Duration":     durationHeader,
		"IncludeRoles": includeRolesHeader,
	}
	return &backupCtx
}

func NewFullBackupFormat(source string) Format {
	switch source {
	case "table", "":
		format := backupGeneral1
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// SingleBackupWrite renders the context for a single backup
func SingleBackupWrite(ctx Context, backup ybmclient.BackupData) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&BackupContext{c: backup})
		if err != nil {
			logrus.Debugf("Error rendering backup: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewFullBackupContext(), render)
}

func (b *FullBackupContext) SetFullBackup(backupData ybmclient.BackupData) {
	fb := backup.NewFullBackup(backupData)
	b.fullBackup = fb
}

func (b *FullBackupContext) startSubsection(format string) (*template.Template, error) {
	b.buffer = bytes.NewBufferString("")
	b.header = ""
	b.Format = Format(format)
	b.preFormat()

	return b.parseFormat()
}

type fullBackupContext struct {
	Backup           *BackupContext
	DatabasesContext []*databasesContext
}

type databasesContext struct {
	HeaderContext
	d ybmclient.BackupKeyspaceInfo
}

func NewDatabasesContext() *databasesContext {
	databasesCtx := databasesContext{}
	databasesCtx.Header = SubHeaderContext{
		"Db":      dbHeader,
		"ApiType": apiTypeHeader,
	}
	return &databasesCtx
}

func (d *databasesContext) Db() string {
	return d.d.GetKeyspaces()[0]
}

func (d *databasesContext) ApiType() string {
	return string(d.d.GetTableType())
}

func (d *databasesContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.d)
}

func (b *FullBackupContext) SubSection(name string) {
	b.Output.Write([]byte("\n\n"))
	b.Output.Write([]byte(Colorize(name, GREEN_COLOR)))
	b.Output.Write([]byte("\n"))
}

func (b *FullBackupContext) Write() error {
	fbc := &fullBackupContext{
		Backup:           &BackupContext{},
		DatabasesContext: make([]*databasesContext, 0, len(b.fullBackup.Databases)),
	}

	fbc.Backup.c = b.fullBackup.Backup

	//Adding databases/keyspaces information
	for _, database := range b.fullBackup.Databases {
		fbc.DatabasesContext = append(fbc.DatabasesContext, &databasesContext{d: database})
	}

	//First Section
	tmpl, err := b.startSubsection(backupGeneral1)
	if err != nil {
		return err
	}
	b.Output.Write([]byte(Colorize("General", GREEN_COLOR)))
	b.Output.Write([]byte("\n"))
	if err := b.contextFormat(tmpl, fbc.Backup); err != nil {
		return err
	}
	b.postFormat(tmpl, NewFullBackupContext())

	tmpl, err = b.startSubsection(backupGeneral2)
	if err != nil {
		return err
	}
	b.Output.Write([]byte(Colorize("", GREEN_COLOR)))
	b.Output.Write([]byte("\n"))
	if err := b.contextFormat(tmpl, fbc.Backup); err != nil {
		return err
	}
	b.postFormat(tmpl, NewFullBackupContext())

	// Databases/Keyspaces
	if len(fbc.DatabasesContext) > 0 {
		tmpl, err = b.startSubsection(databaseListing)
		if err != nil {
			return err
		}
		b.SubSection("Databases/Keyspaces")
		for _, v := range fbc.DatabasesContext {
			if err := b.contextFormat(tmpl, v); err != nil {
				return err
			}
		}
		b.postFormat(tmpl, NewDatabasesContext())
	}

	return nil
}
