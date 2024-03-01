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
	"fmt"
	"strconv"
	"time"

	"github.com/enescakir/emoji"
	"github.com/sirupsen/logrus"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultBackupListing   = "table {{.Id}}\t{{.CreatedOn}}\t{{.Incremental}}\t{{.ExpireOn}}\t{{.ClusterName}}\t{{.BackupState}}\t{{.BackupType}}"
	incrementalHeader      = "Inc"
	backupIdCreateOnHeader = "Created On"
	backupIdExpireOnHeader = "Expire On"
	backupIdHeader         = "ID"
	backupTypeHeader       = "Type"
	retainInDaysHeader     = "Retains(day)"
	sizeHeader             = "Size"
)

type BackupContext struct {
	HeaderContext
	Context
	c ybmclient.BackupData
}

func NewBackupFormat(source string) Format {
	switch source {
	case "table", "":
		format := defaultBackupListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

// BackupWrite renders the context for a list of backups
func BackupWrite(ctx Context, Backups []ybmclient.BackupData) error {
	render := func(format func(subContext SubContext) error) error {
		for _, Backup := range Backups {
			err := format(&BackupContext{c: Backup})
			if err != nil {
				logrus.Debugf("Error in rendering backup context: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewBackupContext(), render)
}

// NewBackupContext creates a new context for rendering Backup
func NewBackupContext() *BackupContext {
	BackupCtx := BackupContext{}
	BackupCtx.Header = SubHeaderContext{
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
	}
	return &BackupCtx
}

func (c *BackupContext) Incremental() string {
	if *c.c.GetSpec().BackupType.Ptr() == ybmclient.BACKUPTYPEENUM_INCREMENTAL {
		return emoji.Pizza.String()
	}
	return emoji.YellowCircle.String()
}

func (c *BackupContext) Size() string {
	// Convert the size to string
	return strconv.FormatInt(*c.c.GetInfo().SizeInBytes.Get(), 10)
}

func (c *BackupContext) ExpireOn() string {
	CreatedOn := c.CreatedOn()
	if len(CreatedOn) > 0 {
		t, _ := time.Parse("2006-01-02,15:04", CreatedOn)
		retainInDay, _ := strconv.Atoi(c.RetainInDays())
		//CreatedOn is already formatted in Local time
		return t.AddDate(0, 0, retainInDay).Format("2006-01-02,15:04")
	}
	return ""
}

func (c *BackupContext) CreatedOn() string {
	if !c.c.GetInfo().Metadata.Get().HasCreatedOn() {
		return ""

	}
	return FormatDate(c.c.GetInfo().Metadata.Get().GetCreatedOn())
}

func (c *BackupContext) ClusterName() string {
	return *c.c.GetInfo().ClusterName
}

func (c *BackupContext) Description() string {
	if v, ok := c.c.Spec.GetDescriptionOk(); ok {
		if v == nil {
			return ""
		}
		return Truncate(*v, 10)
	}
	return ""
}

func (c *BackupContext) Id() string {
	return *c.c.GetInfo().Id
}

func (c *BackupContext) BackupState() string {
	switch *c.c.GetInfo().State {
	case ybmclient.BACKUPSTATEENUM_SUCCEEDED:
		return emoji.CheckMarkButton.String()
	case ybmclient.BACKUPSTATEENUM_FAILED:
		return emoji.CrossMark.String()
	case ybmclient.BACKUPSTATEENUM_IN_PROGRESS:
		return emoji.HourglassNotDone.String()
	default:
		return emoji.QuestionMark.String()
	}
}

func (c *BackupContext) BackupType() string {
	switch *c.c.GetInfo().ActionType {
	case ybmclient.BACKUPACTIONTYPEENUM_MANUAL:
		return emoji.Person.String()
	case ybmclient.BACKUPACTIONTYPEENUM_SCHEDULED:
		return emoji.Calendar.String()
	default:
		return emoji.QuestionMark.String()
	}
}

func (c *BackupContext) RetainInDays() string {
	return fmt.Sprint(*c.c.GetSpec().RetentionPeriodInDays)
}

func (c *BackupContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}

// TODO add error handling
func FormatDate(dateToBeFormatted string) string {
	t, _ := time.Parse(time.RFC3339Nano, dateToBeFormatted)
	return t.Local().Format("2006-01-02,15:04")
}
