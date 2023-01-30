/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
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
