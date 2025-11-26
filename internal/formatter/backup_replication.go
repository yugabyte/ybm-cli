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
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultBackupReplicationListing                      = "table {{.Region}}\t{{.ConfigState}}\t{{.BucketName}}\t{{.LatestTransferOperationStatus}}\t{{.LastRun}}\t{{.NextRun}}"
	extendedBackupReplicationListing                     = "table {{.Region}}\t{{.ConfigState}}\t{{.BucketName}}\t{{.LatestTransferOperationStatus}}\t{{.LastRun}}\t{{.NextRun}}\t{{.ExpiryTime}}"
	backupReplicationRegionHeader                        = "Region"
	backupReplicationConfigStateHeader                   = "Config State"
	backupReplicationBucketHeader                        = "Bucket Name"
	backupReplicationLatestTransferOperationStatusHeader = "Latest Operation Status"
	backupReplicationLastRunHeader                       = "Last Run"
	backupReplicationNextRunHeader                       = "Next Run"
	backupReplicationExpiryTimeHeader                    = "Expiry Time"
	overallStateHeader                                   = "Overall State"
	noActiveConfigurationsMessage                        = "No active configurations are present for the cluster."
	noExpiredConfigurationsMessage                       = "No configurations are present for expiry."
)

type backupReplicationReportData struct {
	region          string
	bucketName      string
	report          *ybmclient.GcpBackupReplicationRegionReport
	isRemovedRegion bool
}

type BackupReplicationContext struct {
	HeaderContext
	Context
	r backupReplicationReportData
}

type ExpiredReport struct {
	r backupReplicationReportData
}

type BackupReplicationFullContext struct {
	HeaderContext
	Context
	data ybmclient.GcpBackupReplicationData
}

func NewBackupReplicationFormat(source string, showAll bool) Format {
	if source == "table" || source == "" {
		if showAll {
			return Format(extendedBackupReplicationListing)
		}
		return Format(defaultBackupReplicationListing)
	}
	return Format(source)
}

func processExpiredReport(report ybmclient.GcpBackupReplicationRegionReport, regionName, bucketName string, isRemovedRegion bool) *ExpiredReport {
	return &ExpiredReport{
		r: backupReplicationReportData{
			region:          regionName,
			bucketName:      bucketName,
			report:          &report,
			isRemovedRegion: isRemovedRegion,
		},
	}
}

func sortExpiredReports(reports []*ExpiredReport, specRegionOrder []string) []*ExpiredReport {
	sortedReports := make([]*ExpiredReport, 0, len(reports))
	removedRegionReports := make([]*ExpiredReport, 0)

	for _, specRegion := range specRegionOrder {
		for _, report := range reports {
			if report.r.region == specRegion && !report.r.isRemovedRegion {
				sortedReports = append(sortedReports, report)
			}
		}
	}

	for _, report := range reports {
		if report.r.isRemovedRegion {
			removedRegionReports = append(removedRegionReports, report)
		}
	}
	sortedReports = append(sortedReports, removedRegionReports...)

	return sortedReports
}

func renderActiveReports(ctx Context, reports []*BackupReplicationContext, format func(SubContext) error) error {
	if len(reports) == 0 {
		if ctx.Format.IsTable() {
			fmt.Fprintf(ctx.Output, "%s\n\n", noActiveConfigurationsMessage)
		}
		return nil
	}

	for _, report := range reports {
		if err := format(report); err != nil {
			logrus.Debugf("Error rendering active backup replication: %v", err)
			return err
		}
	}
	return nil
}

func renderExpiredReports(ctx Context, reports []*ExpiredReport, specRegionOrder []string, format func(SubContext) error) error {
	if len(reports) == 0 {
		if ctx.Format.IsTable() {
			fmt.Fprintf(ctx.Output, "%s\n", noExpiredConfigurationsMessage)
		}
		return nil
	}

	sortedReports := sortExpiredReports(reports, specRegionOrder)

	expiredFormat := NewBackupReplicationFormat(viper.GetString("output"), true)
	expiredFormatCtx := Context{
		Output: ctx.Output,
		Format: expiredFormat,
	}
	expiredHeaderCtx := NewBackupReplicationContext()
	if headerMap, ok := expiredHeaderCtx.Header.(SubHeaderContext); ok {
		headerMap["ExpiryTime"] = backupReplicationExpiryTimeHeader
	}

	expiredRender := func(formatFunc func(subContext SubContext) error) error {
		for _, report := range sortedReports {
			regionName := report.r.region
			if report.r.isRemovedRegion {
				regionName = fmt.Sprintf("%s (Removed Region)", report.r.region)
			}

			ctx := &BackupReplicationContext{
				r: backupReplicationReportData{
					region:          regionName,
					bucketName:      report.r.bucketName,
					report:          report.r.report,
					isRemovedRegion: report.r.isRemovedRegion,
				},
			}

			if err := formatFunc(ctx); err != nil {
				logrus.Debugf("Error rendering expired backup replication: %v", err)
				return err
			}
		}
		return nil
	}

	return expiredFormatCtx.Write(expiredHeaderCtx, expiredRender)
}

func BackupReplicationWrite(ctx Context, backupReplicationData ybmclient.GcpBackupReplicationData, showAll bool) error {
	if ctx.Format.IsJSON() || ctx.Format.IsPrettyJson() {
		fullCtx := &BackupReplicationFullContext{
			data: backupReplicationData,
		}
		fullCtx.Header = SubHeaderContext{}

		render := func(formatFunc func(subContext SubContext) error) error {
			return formatFunc(fullCtx)
		}

		return ctx.Write(fullCtx, render)
	}

	if ctx.Format.IsTable() {
		overallState := string(backupReplicationData.Info.GetState())
		fmt.Fprintf(ctx.Output, "%s: %s\n\n", overallStateHeader, overallState)
	}

	regionConfigMap := make(map[string]ybmclient.GcpBackupReplicationRegionConfig)
	if regionConfigs, ok := backupReplicationData.Info.GetRegionConfigsOk(); ok && regionConfigs != nil {
		for _, regionConfig := range *regionConfigs {
			regionConfigMap[regionConfig.GetRegion()] = regionConfig
		}
	}

	specRegions := make(map[string]bool)
	var specRegionOrder []string

	var activeReports []*BackupReplicationContext
	var expiredReports []*ExpiredReport

	if spec, ok := backupReplicationData.GetSpecOk(); ok && spec != nil {
		if regionalTargets, ok := spec.GetRegionalTargetsOk(); ok && regionalTargets != nil {
			for _, regionalTarget := range *regionalTargets {
				regionName := regionalTarget.GetRegion()
				specRegions[regionName] = true
				specRegionOrder = append(specRegionOrder, regionName)

				bucketName := "N/A"
				if target, ok := regionalTarget.GetTargetOk(); ok && target != nil {
					bucketName = *target
				}

				regionConfig, exists := regionConfigMap[regionName]
				hasActiveReport := false

				if exists {
					for _, report := range regionConfig.GetReports() {
						metadata, ok := report.GetMetadataOk()
						if !ok || metadata == nil {
							continue
						}

						expiryOn, isSet := metadata.GetExpiryOnOk()
						isExpiryNull := !isSet || (isSet && expiryOn == nil)

						if isExpiryNull {
							hasActiveReport = true
							activeReports = append(activeReports, &BackupReplicationContext{
								r: backupReplicationReportData{
									region:          regionName,
									bucketName:      bucketName,
									report:          &report,
									isRemovedRegion: false,
								},
							})
						} else if showAll {
							reportBucketName := bucketName
							if reportTarget, ok := report.GetTargetOk(); ok && reportTarget != nil {
								reportBucketName = *reportTarget
							}
							expiredReports = append(expiredReports, processExpiredReport(report, regionName, reportBucketName, false))
						}
					}
				}

				if !hasActiveReport && !showAll {
					activeReports = append(activeReports, &BackupReplicationContext{
						r: backupReplicationReportData{
							region:          regionName,
							bucketName:      bucketName,
							report:          nil,
							isRemovedRegion: false,
						},
					})
				}
			}
		}
	}

	if showAll {
		for regionName, regionConfig := range regionConfigMap {
			if !specRegions[regionName] {
				for _, report := range regionConfig.GetReports() {
					reportBucketName := "N/A"
					if reportTarget, ok := report.GetTargetOk(); ok && reportTarget != nil {
						reportBucketName = *reportTarget
					}
					expiredReports = append(expiredReports, processExpiredReport(report, regionName, reportBucketName, true))
				}
			}
		}
	}

	if !showAll {
		format := NewBackupReplicationFormat(viper.GetString("output"), false)
		ctx.Format = format
		headerCtx := NewBackupReplicationContext()

		render := func(formatFunc func(subContext SubContext) error) error {
			return renderActiveReports(ctx, activeReports, formatFunc)
		}

		return ctx.Write(headerCtx, render)
	}

	if ctx.Format.IsTable() {
		fmt.Fprintf(ctx.Output, "%s\n", Colorize("=== Active Configurations ===", GREEN_COLOR))

		activeFormat := NewBackupReplicationFormat(viper.GetString("output"), false)
		activeFormatCtx := Context{
			Output: ctx.Output,
			Format: activeFormat,
		}
		activeHeaderCtx := NewBackupReplicationContext()

		activeRender := func(formatFunc func(subContext SubContext) error) error {
			return renderActiveReports(activeFormatCtx, activeReports, formatFunc)
		}

		if err := activeFormatCtx.Write(activeHeaderCtx, activeRender); err != nil {
			return err
		}

		fmt.Fprintf(ctx.Output, "\n%s\n", Colorize("=== Configurations Set for Expiry ===", GREEN_COLOR))

		return renderExpiredReports(ctx, expiredReports, specRegionOrder, func(subContext SubContext) error {
			return nil
		})
	}

	format := NewBackupReplicationFormat(viper.GetString("output"), false)
	ctx.Format = format
	headerCtx := NewBackupReplicationContext()

	render := func(formatFunc func(subContext SubContext) error) error {
		if err := renderActiveReports(ctx, activeReports, formatFunc); err != nil {
			return err
		}
		return renderExpiredReports(ctx, expiredReports, specRegionOrder, formatFunc)
	}

	return ctx.Write(headerCtx, render)
}

func NewBackupReplicationContext() *BackupReplicationContext {
	backupReplicationCtx := BackupReplicationContext{}
	backupReplicationCtx.Header = SubHeaderContext{
		"Region":                        backupReplicationRegionHeader,
		"ConfigState":                   backupReplicationConfigStateHeader,
		"BucketName":                    backupReplicationBucketHeader,
		"LatestTransferOperationStatus": backupReplicationLatestTransferOperationStatusHeader,
		"LastRun":                       backupReplicationLastRunHeader,
		"NextRun":                       backupReplicationNextRunHeader,
	}
	return &backupReplicationCtx
}

func (b *BackupReplicationContext) Region() string {
	return b.r.region
}

func (b *BackupReplicationContext) ConfigState() string {
	if b.r.report == nil {
		return "DISABLED"
	}
	return string(b.r.report.GetConfigState())
}

func (b *BackupReplicationContext) BucketName() string {
	return b.r.bucketName
}

func (b *BackupReplicationContext) LatestTransferOperationStatus() string {
	if b.r.report == nil {
		return "N/A"
	}
	transferJobDetails, ok := b.r.report.GetTransferJobDetailsOk()
	if !ok || transferJobDetails == nil {
		return "N/A"
	}
	if latestOp, ok := transferJobDetails.GetLatestTransferOperationDetailsOk(); ok && latestOp != nil {
		return latestOp.GetStatus()
	}
	return "N/A"
}

func (b *BackupReplicationContext) LastRun() string {
	if b.r.report == nil {
		return "N/A"
	}
	transferJobDetails, ok := b.r.report.GetTransferJobDetailsOk()
	if !ok || transferJobDetails == nil {
		return "N/A"
	}
	if latestOp, ok := transferJobDetails.GetLatestTransferOperationDetailsOk(); ok && latestOp != nil {
		if endTime := latestOp.GetEndTime(); !endTime.IsZero() {
			return formatTime(endTime)
		}
	}
	return "N/A"
}

func (b *BackupReplicationContext) NextRun() string {
	if b.r.report == nil {
		return "N/A"
	}
	transferJobDetails, ok := b.r.report.GetTransferJobDetailsOk()
	if !ok || transferJobDetails == nil {
		return "N/A"
	}
	return formatTime(transferJobDetails.GetNextTransferOperationTime())
}

func (b *BackupReplicationContext) ExpiryTime() string {
	if b.r.report == nil {
		return "N/A"
	}
	metadata, ok := b.r.report.GetMetadataOk()
	if !ok || metadata == nil {
		return "N/A"
	}
	expiryOn, isSet := metadata.GetExpiryOnOk()
	if !isSet || expiryOn == nil {
		return "N/A"
	}

	parsedTime, err := time.Parse(time.RFC3339, *expiryOn)
	if err != nil {
		parsedTime, err = time.Parse("2006-01-02T15:04:05Z07:00", *expiryOn)
		if err != nil {
			return *expiryOn
		}
	}
	return formatTime(parsedTime)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Local().Format("2006-01-02,15:04")
}

func (b *BackupReplicationContext) MarshalJSON() ([]byte, error) {
	if b.r.report != nil {
		reportJSON, err := json.Marshal(b.r.report)
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(reportJSON, &result); err != nil {
			return nil, err
		}

		result["region"] = b.r.region
		if b.r.isRemovedRegion {
			result["isRemovedRegion"] = true
		}

		return json.Marshal(result)
	}

	result := map[string]interface{}{
		"region":      b.r.region,
		"bucketName":  b.r.bucketName,
		"configState": "DISABLED",
	}

	return json.Marshal(result)
}

func (b *BackupReplicationFullContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.data)
}
