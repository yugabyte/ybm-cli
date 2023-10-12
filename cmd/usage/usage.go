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

package usage

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var UsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Billing usage for the account in YugabyteDB Managed",
	Long:  "Billing usage for the account in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "View billing usage data available for the account in YugabyteDB Managed",
	Long:  "View billing usage data available for the account in YugabyteDB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")
		startDate, _ := cmd.Flags().GetString("start")
		endDate, _ := cmd.Flags().GetString("end")
		outputFormat, _ := cmd.Flags().GetString("output-format")
		outputFile, _ := cmd.Flags().GetString("output-file")
		clusters, _ := cmd.Flags().GetStringArray("cluster-name")

		if startDate == "" || endDate == "" {
			// If either start date or end date is missing, throw an error
			logrus.Fatalf("Both start date and end date are required.\n")
		}

		// Validate start date and end date
		startDateTime, err := parseAndFormatDate(startDate)
		if err != nil {
			logrus.Fatalf("Invalid start date format. Use either RFC3339 format (e.g., '2023-09-01T12:30:45.000Z') or 'yyyy-MM-dd' format (e.g., '2023-09-01').\n")
		}

		endDateTime, err := parseAndFormatDate(endDate)
		if err != nil {
			logrus.Fatalf("Invalid end date format. Use either RFC3339 format (e.g., '2023-09-01T12:30:45.000Z') or 'yyyy-MM-dd' format (e.g., '2023-09-01').\n")
		}

		if startDateTime.After(endDateTime) {
			logrus.Fatalf("Start date must be before end date.")
		}

		startDate = startDateTime.Format("2006-01-02T15:04:05.000Z")
		endDate = endDateTime.Format("2006-01-02T15:04:05.000Z")

		// Assigning default value to filename if not specified
		outputFileFormat := "usage_%s_%s"
		if outputFile == "" {
			startDateComponents := startDateTime.Format("20060102T150405") // Format as YYYYMMDDHHmmSS
			endDateComponents := endDateTime.Format("20060102T150405")     // Format as YYYYMMDDHHmmSS
			outputFile = fmt.Sprintf(outputFileFormat, startDateComponents, endDateComponents)
		}

		// Check if the file already exists
		if _, err := os.Stat(outputFile); err == nil {
			logrus.Fatalf("File %s already exists. Please choose a different output file name.\n", formatter.Colorize(outputFile, formatter.GREEN_COLOR))
		}

		respC, r, err := authApi.ListClustersByDateRange(startDate, endDate).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		clusterData := respC.Data.GetClusters()
		var clusterUUIDMap = make(map[string]string)

		for _, cluster := range clusterData {
			clusterName := cluster.Name
			clusterUUIDMap[clusterName] = cluster.Id
		}

		var selectedUUIDs []string
		if len(clusters) > 0 {
			selectedUUIDs = make([]string, 0, len(clusters))
			for _, clusterName := range clusters {
				if clusterID, ok := clusterUUIDMap[clusterName]; ok {
					selectedUUIDs = append(selectedUUIDs, clusterID)
				} else {
					logrus.Fatalf("Cluster name '%s' not found within the specified time range.\n", clusterName)
				}
			}
		} else {
			selectedUUIDs = nil
		}

		resp, r, err := authApi.GetBillingUsage(startDate, endDate, selectedUUIDs).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		usageData := resp.GetData()

		switch strings.ToLower(outputFormat) {
		case "csv":
			if err := outputCSV(usageData, outputFile); err != nil {
				logrus.Fatalf("Error outputting CSV: %v", err)
			}
		case "json":
			if err := outputJSON(usageData, outputFile); err != nil {
				logrus.Fatalf("Error outputting JSON: %v", err)
			}
		default:
			logrus.Warnf("Unsupported format: %s. Defaulting to CSV.", outputFormat)
			if err := outputCSV(usageData, outputFile); err != nil {
				logrus.Fatalf("Error outputting CSV: %v", err)
			}
		}
	},
}

func parseAndFormatDate(dateStr string) (time.Time, error) {
	// Try parsing in RFC3339 format
	parsedDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// If parsing fails, try parsing in "yyyy-MM-dd" format
		parsedDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return time.Time{}, err
		}
	}

	return parsedDate, nil
}

func outputCSV(resp ybmclient.BillingUsageData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	dimensionHeaders := []string{"Date"}
	for _, dimension := range resp.GetDimensions() {
		dimensionHeaders = append(dimensionHeaders, string(dimension.GetName())+"_Daily", string(dimension.GetName())+"_Cumulative")
	}

	err = writer.Write(dimensionHeaders)
	if err != nil {
		return err
	}

	for _, dataPoint := range resp.GetDimensions()[0].DataPoints {
		row := []string{dataPoint.GetStartDate()}

		for _, dimension := range resp.GetDimensions() {
			var found bool
			for _, point := range dimension.DataPoints {
				if point.StartDate == dataPoint.StartDate {
					row = append(row, fmt.Sprintf("%f", *point.DailyUsage))
					row = append(row, fmt.Sprintf("%f", *point.CumulativeUsage))
					found = true
					break
				}
			}

			// If no data found for the current dimension, add placeholders
			if !found {
				row = append(row, "0.000000", "0.000000")
			}
		}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	fmt.Printf("CSV data written to %s.csv.\n", formatter.Colorize(filename, formatter.GREEN_COLOR))
	return nil
}

func outputJSON(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create or open JSON file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write JSON data to file: %v", err)
	}

	fmt.Printf("JSON data written to %s.json.\n", formatter.Colorize(filename, formatter.GREEN_COLOR))
	return nil
}

func init() {
	UsageCmd.AddCommand(getCmd)
	getCmd.Flags().String("start", "", "[REQUIRED] Start date in RFC3339 format (e.g., '2023-09-01T12:30:45.000Z') or 'yyyy-MM-dd' format (e.g., '2023-09-01').")
	getCmd.Flags().String("end", "", "[REQUIRED] End date in RFC3339 format (e.g., '2023-09-30T23:59:59.999Z') or 'yyyy-MM-dd' format (e.g., '2023-09-30').")
	getCmd.Flags().StringArray("cluster-name", []string{}, "[REQUIRED] Cluster names. Multiple names can be specified by using multiple --cluster-name arguments.")
	getCmd.Flags().String("output-format", "csv", "[OPTIONAL] Output format. Possible values: csv, json.")
	getCmd.Flags().String("output-file", "", "[OPTIONAL] Output filename.")
}
