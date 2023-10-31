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
	"path/filepath"
	"sort"
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
		force := cmd.Flags().Changed("force")

		startDateTime, endDateTime, err := validateDates(startDate, endDate)
		if err != nil {
			logrus.Fatalf("Error: %v\n", err)
		}

		// Format dates
		startDate = startDateTime.Format("2006-01-02T15:04:05.000Z")
		endDate = endDateTime.Format("2006-01-02T15:04:05.000Z")

		combinedFilename, fileExtension, err := checkOutputFile(outputFile, outputFormat, startDateTime, endDateTime, force)
		if err != nil {
			logrus.Fatalf("Error: %v", err)
		}

		selectedUUIDs, selectedClusterNames, err := getSelectedUUIDs(startDate, endDate, clusters, authApi)
		if err != nil {
			logrus.Fatalf("Error: %v", err)
		}

		resp, r, err := authApi.GetBillingUsage(startDate, endDate, selectedUUIDs).Execute()

		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		handleOutput(resp, combinedFilename, fileExtension, selectedClusterNames)
	},
}

func handleOutput(usageResponse ybmclient.BillingUsageResponse, combinedFilename, fileExtension string, selectedClusterNames []string) {
	switch fileExtension {
	case "csv":
		if err := outputCSV(usageResponse.GetData(), combinedFilename, selectedClusterNames); err != nil {
			logrus.Fatalf("Error outputting CSV: %v", err)
		}
	case "json":
		usageResponse.Data.Get().SetClusterIds(selectedClusterNames)
		if err := outputJSON(usageResponse, combinedFilename); err != nil {
			logrus.Fatalf("Error outputting JSON: %v", err)
		}
	default:
		logrus.Warnf("Unsupported format: %s. Defaulting to CSV.", fileExtension)
		if err := outputCSV(usageResponse.GetData(), combinedFilename, selectedClusterNames); err != nil {
			logrus.Fatalf("Error outputting CSV: %v", err)
		}
	}
}

func getSelectedUUIDs(startDate, endDate string, clusters []string, authApi *ybmAuthClient.AuthApiClient) ([]string, []string, error) {
	respC, r, err := authApi.ListClustersByDateRange(startDate, endDate).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		return nil, nil, fmt.Errorf(ybmAuthClient.GetApiErrorDetails(err))
	}

	clusterData := respC.Data.GetClusters()
	clusterUUIDMap := make(map[string]string)

	// Map of <cluster uuid, cluster name> for all clusters
	for _, cluster := range clusterData {
		clusterName := cluster.Name
		clusterUUIDMap[clusterName] = cluster.Id
	}

	var selectedClusterUUIDs []string
	var selectedClusterNames []string
	var notFoundClusters []string

	if len(clusters) > 0 {
		selectedClusterUUIDs = make([]string, 0, len(clusters))
		for _, clusterName := range clusters {
			if clusterID, ok := clusterUUIDMap[clusterName]; ok {
				selectedClusterUUIDs = append(selectedClusterUUIDs, clusterID)
				selectedClusterNames = append(selectedClusterNames, clusterName)
			} else {
				notFoundClusters = append(notFoundClusters, clusterName)
			}
		}
	} else { // By default select all clusters
		for clusterName, _ := range clusterUUIDMap {
			selectedClusterNames = append(selectedClusterNames, clusterName)
		}
	}

	if len(notFoundClusters) > 0 {
		return nil, nil, fmt.Errorf("clusters '%s' not found within the specified time range", strings.Join(notFoundClusters, ", "))
	}
	sort.Strings(selectedClusterNames)
	return selectedClusterUUIDs, selectedClusterNames, nil
}

func validateDates(startDate, endDate string) (startDateTime, endDateTime time.Time, err error) {
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("both start date and end date are required")
	}
	startDateTime, err = parseAndFormatDate(startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format. Use either RFC3339 format (e.g., '2023-09-01T12:30:45.000Z') or 'yyyy-MM-dd' format (e.g., '2023-09-01')")
	}

	endDateTime, err = parseAndFormatDate(endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format. Use either RFC3339 format (e.g., '2023-09-01T12:30:45.000Z') or 'yyyy-MM-dd' format (e.g., '2023-09-01')")
	}

	if startDateTime.After(endDateTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("start date must be before end date")
	}

	return startDateTime, endDateTime, nil
}

func checkOutputFile(outputFile, outputFormat string, startDateTime, endDateTime time.Time, force bool) (combinedFilename, fileExtension string, err error) {
	// Assigning default value to filename if not specified
	outputFileFormat := "usage_%s_%s"
	if outputFile == "" {
		startDateComponents := startDateTime.Format("20060102T150405") // Format as YYYYMMDDHHmmSS
		endDateComponents := endDateTime.Format("20060102T150405")     // Format as YYYYMMDDHHmmSS
		outputFile = fmt.Sprintf(outputFileFormat, startDateComponents, endDateComponents)
	}

	outputFormat = strings.ToLower(outputFormat)
	combinedFilename = getNormalizedFilePathForFormat(outputFile, outputFormat)
	fileExtension = strings.ToLower(strings.TrimPrefix(filepath.Ext(combinedFilename), "."))

	if outputFormat != "" && outputFormat != fileExtension {
		logrus.Warnf("Mismatch between output file extension and output format. Using file extension: %s\n", fileExtension)
	}

	// Check if the file already exists
	if !force {
		if _, err := os.Stat(combinedFilename); err == nil {
			return "", "", fmt.Errorf("file %s already exists", combinedFilename)
		}
	}

	if fileExtension != "csv" && fileExtension != "json" {
		return "", "", fmt.Errorf("file extension %s is not supported", fileExtension)
	}

	return combinedFilename, fileExtension, nil
}

func getNormalizedFilePathForFormat(filePath, outputFormat string) string {
	filePathWithoutExtension := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	validExtensions := map[string]struct{}{"json": {}, "csv": {}}
	_, isValid := validExtensions[extension]
	if isValid {
		return filePathWithoutExtension + "." + extension
	} else if extension != "" {
		logrus.Warnf("Unsupported extension type: %s\n", extension)
	}

	// If extension is not valid and outputFormat is empty, default to "csv"
	if outputFormat == "" {
		return filePathWithoutExtension + ".csv"
	}
	return filePathWithoutExtension + "." + strings.ToLower(outputFormat)
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

func outputCSV(resp ybmclient.BillingUsageData, filename string, selectedClusterNames []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	dimensionHeaders := []string{"Date", "Clusters"}
	for _, dimension := range resp.GetDimensions() {
		dimensionHeaders = append(dimensionHeaders, string(dimension.GetName())+"_Daily", string(dimension.GetName())+"_Cumulative")
	}

	err = writer.Write(dimensionHeaders)
	if err != nil {
		return err
	}

	for _, dataPoint := range resp.GetDimensions()[0].DataPoints {
		row := []string{dataPoint.GetStartDate()}
		row = append(row, "'"+strings.Join(selectedClusterNames, "','")+"'")

		for _, dimension := range resp.GetDimensions() {
			for _, point := range dimension.DataPoints {
				if point.GetStartDate() == dataPoint.GetStartDate() {
					row = append(row, fmt.Sprintf("%f", *point.DailyUsage))
					row = append(row, fmt.Sprintf("%f", *point.CumulativeUsage))
					break
				}
			}
		}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	fmt.Printf("CSV data written to %s\n", formatter.Colorize(filename, formatter.GREEN_COLOR))
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

	fmt.Printf("JSON data written to %s\n", formatter.Colorize(filename, formatter.GREEN_COLOR))
	return nil
}

func init() {
	UsageCmd.AddCommand(getCmd)
	getCmd.Flags().String("start", "", "[REQUIRED] Start date in RFC3339 format (e.g., '2023-09-01T12:30:45.000Z') or 'yyyy-MM-dd' format (e.g., '2023-09-01').")
	getCmd.Flags().String("end", "", "[REQUIRED] End date in RFC3339 format (e.g., '2023-09-30T23:59:59.999Z') or 'yyyy-MM-dd' format (e.g., '2023-09-30').")
	getCmd.Flags().StringArray("cluster-name", []string{}, "[OPTIONAL] Cluster names. Multiple names can be specified by using multiple --cluster-name arguments.")
	getCmd.Flags().String("output-format", "", "[OPTIONAL] Output format. Possible values: csv, json.")
	getCmd.Flags().String("output-file", "", "[OPTIONAL] Output filename.")
	getCmd.Flags().BoolP("force", "f", false, "[OPTIONAL] Overwrite the output file if it exists")
}
