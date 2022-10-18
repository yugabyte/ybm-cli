/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// greetCmd represents the greet command
var greetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Greet the users of YBM CLI",
	Long:  "Greet the users of YBM CLI",
	Run: func(cmd *cobra.Command, args []string) {

		apiClient, _ := getApiClient(context.Background())
		resp, r, err := apiClient.GreetingsApi.GetGreetings(context.Background()).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `GreetingsApi.GetGreetings`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}
		prettyPrintJson(resp)

	},
}

func init() {
	rootCmd.AddCommand(greetCmd)

}
