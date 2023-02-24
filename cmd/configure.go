// Copyright (c) YugaByte, Inc.
//
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
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022-present Yugabyte, Inc.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure ybm CLI",
	Long:  "Configure the ybm CLI through this command by providing the API Key.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Enter API Key: ")
		var apiKey string
		var host string
		fmt.Scanln(&apiKey)
		viper.GetViper().Set("apikey", &apiKey)
		fmt.Print("Enter Host: ")
		fmt.Scanln(&host)
		viper.GetViper().Set("host", &host)
		err := viper.WriteConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Fprintln(os.Stdout, "No config was found a new one will be created.")
				//Try to create the file
				err = viper.SafeWriteConfig()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error when writing new config file: %v", err)

				}
			} else {
				fmt.Fprintf(os.Stderr, "Error when writing config file: %v", err)
				return
			}
		}
		fmt.Println("Configuration file sucessfully updated.")
	},
}
