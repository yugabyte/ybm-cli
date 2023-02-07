/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package greet

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

// greetCmd represents the greet command
var GreetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Greet the users of YBM CLI",
	Long:  "Greet the users of YBM CLI",
	Run: func(cmd *cobra.Command, args []string) {

		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Errorf("could not initiate api client: %s", err.Error())
			os.Exit(1)
		}
		authApi.GetInfo("", "")

		resp, r, err := authApi.GetGreetings().Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `GreetingsApi.GetGreetings`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}
		fmt.Println(resp.GetData())

	},
}

func init() {

}
